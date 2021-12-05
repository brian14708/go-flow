package flowdebug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"

	"github.com/brian14708/go-flow/flowdebug/internal/viewer"
	"github.com/brian14708/go-flow/flowdebug/types"
)

type Profiler struct {
	opt    *ProfilerOptions
	mux    *http.ServeMux
	cancel context.CancelFunc

	mu     sync.RWMutex
	graphs map[string]*graph

	metrics *metricSet
}

type graph struct {
	uuid     string
	topology json.RawMessage
	err      *error
	stopTime time.Time

	metricNames []string
}

type ProfilerOptions struct {
	Debug         bool
	HistoryLength int
	SampleRate    time.Duration

	Logger *logrus.Logger
}

func NewProfiler(opt *ProfilerOptions) *Profiler {
	if opt == nil {
		opt = new(ProfilerOptions)
	}
	if opt.SampleRate == time.Duration(0) {
		opt.SampleRate = 5 * time.Second
	}
	if opt.HistoryLength <= 0 {
		opt.HistoryLength = 16
	}
	if opt.Logger == nil {
		opt.Logger = logrus.StandardLogger()
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := &Profiler{
		opt:    opt,
		mux:    http.NewServeMux(),
		cancel: cancel,

		graphs: make(map[string]*graph),

		metrics: newMetricSet(ctx, opt),
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				old := now.Add(-time.Minute)
				p.mu.Lock()
				for id, g := range p.graphs {
					if g.err != nil && g.stopTime.Before(old) {
						delete(p.graphs, id)
					}
				}
				p.mu.Unlock()
			}
		}
	}()

	p.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		p.index(w, r)
	})
	p.mux.HandleFunc("/graph/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/graph/")
		p.getGraph(id, w, r)
	})
	p.mux.HandleFunc("/metrics/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/metrics/")
		p.getMetrics(id, w, r)
	})

	p.mux.HandleFunc("/viewer/dot/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/viewer/dot/")
		p.viewerDot(id, w, r)
	})
	p.mux.Handle("/viewer/", http.StripPrefix("/viewer/", viewer.Handler()))
	return p
}

func (p *Profiler) Stop() {
	p.cancel()
}

func (p *Profiler) AddMetric(id string, isRate bool) Metric {
	return p.metrics.add(id, isRate)
}

func (p *Profiler) AddGraph(id string, topology *types.Topology, metricsPrefix []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	j, err := json.Marshal(topology)
	if err != nil {
		return err
	}

	if _, ok := p.graphs[id]; ok {
		return fmt.Errorf("graph `%s' already exist in profiler", id)
	}
	names := p.metrics.getByPrefix(metricsPrefix)
	p.graphs[id] = &graph{
		uuid:        xid.New().String(),
		topology:    j,
		metricNames: names,
	}
	return nil
}

func (p *Profiler) RemoveGraph(id string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if g, ok := p.graphs[id]; ok {
		g.err = &err
		for _, n := range g.metricNames {
			p.metrics.remove(n)
		}
		g.stopTime = time.Now()
	}
}

func (p *Profiler) DeleteMetrics(metricsPrefix []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	names := p.metrics.getByPrefix(metricsPrefix)
	for _, n := range names {
		p.metrics.remove(n)
	}
}

func (p *Profiler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mux.ServeHTTP(w, r)
}

func (p *Profiler) getGraph(id string, w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	g, ok := p.graphs[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	if im := r.Header.Get("If-None-Match"); im != "" {
		if g.uuid == im {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	p.metrics.attach(g.metricNames, p.opt.SampleRate*2)
	state := struct {
		ID           string          `json:"id"`
		UUID         string          `json:"uuid"`
		Running      bool            `json:"running"`
		Error        *string         `json:"error"`
		Topology     json.RawMessage `json:"topology"`
		MetricsCount int             `json:"metrics_history_length"`
		Metrics      []string        `json:"metrics"`
	}{
		ID:           id,
		Topology:     g.topology,
		MetricsCount: p.opt.HistoryLength,
		Metrics:      g.metricNames,
	}
	if g.err == nil {
		state.Running = true
	} else if *g.err != nil {
		s := (*g.err).Error()
		state.Error = &s
	}

	w.Header().Add("ETag", g.uuid)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		panic(err)
	}
}

func (p *Profiler) getMetrics(id string, w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	g, ok := p.graphs[id]
	if !ok {
		p.mu.RUnlock()
		http.NotFound(w, r)
		return
	}
	names := g.metricNames
	p.mu.RUnlock()

	timeout := 0
	if t, err := strconv.Atoi(r.URL.Query().Get("timeout")); err == nil {
		timeout = t
	}

	p.metrics.attach(names, p.opt.SampleRate*2)
	etag, ch := p.metrics.etag()
	if im := r.Header.Get("If-None-Match"); im != "" {
		if etag == im {
			if timeout == 0 {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			select {
			case <-ch:
				etag = ""
			case <-time.After(time.Duration(timeout) * time.Second):
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	if etag == "" {
		etag, _ = p.metrics.etag()
	}
	w.Header().Add("ETag", etag)
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/x-flow-metrics") {
		w.Header().Add("Content-Type", "application/x-flow-metrics")
		err := p.metrics.getMetrics(names, w)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		p.metrics.getMetricsJSON(names, w)
	}
}

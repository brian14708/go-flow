package flowdebug

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/rs/xid"
	"go.uber.org/atomic"
)

var guid = xid.New().String()

type Metric struct {
	*atomic.Int32
	isRate bool
}

type History struct {
	values []float32
	metric Metric
	expire time.Time
}

type metricSet struct {
	count int

	mu         sync.RWMutex
	tailIndex  int
	timeMillis []int64
	history    map[string]*History // ring buffer
	metrics    map[string]Metric

	historyWatch chan struct{}
	etagWatch    chan struct{}
}

func newMetricSet(ctx context.Context, opt *ProfilerOptions) *metricSet {
	cnt := opt.HistoryLength
	s := &metricSet{
		count: cnt,

		history:    make(map[string]*History),
		timeMillis: make([]int64, cnt),
		metrics:    make(map[string]Metric),
	}
	go s.watch(ctx, opt.SampleRate)
	return s
}

func (s *metricSet) add(id string, isRate bool) (m Metric) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if metric, ok := s.metrics[id]; !ok {
		m = Metric{
			Int32:  atomic.NewInt32(0),
			isRate: isRate,
		}
		s.metrics[id] = m
	} else {
		m = metric
	}
	return m
}

func (s *metricSet) remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.metrics, id)
}

func (s *metricSet) etag() (string, chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.etagWatch == nil {
		s.etagWatch = make(chan struct{})
	}
	tag := fmt.Sprintf("%s-%d", guid, s.timeMillis[(s.tailIndex+s.count-1)%s.count])
	return tag, s.etagWatch
}

func (s *metricSet) getByPrefix(prefix []string) (names []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k := range s.metrics {
		for _, pre := range prefix {
			if strings.HasPrefix(k, pre) {
				names = append(names, k)
			}
		}
	}
	return
}

func (s *metricSet) getMetricsJSON(names []string, w io.Writer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type metric struct {
		Name   string    `json:"name"`
		Values []float32 `json:"values"`
	}
	result := struct {
		Timestamp []int64  `json:"timestamp_millis"`
		Metrics   []metric `json:"metrics"`
	}{}

	result.Timestamp = append(append(result.Timestamp,
		s.timeMillis[s.tailIndex:]...),
		s.timeMillis[:s.tailIndex]...)

	for _, name := range names {
		m := metric{
			Name: name,
		}
		hist := s.history[name]
		m.Values = append(append(m.Values,
			hist.values[s.tailIndex:]...),
			hist.values[:s.tailIndex]...)
		result.Metrics = append(result.Metrics, m)
	}
	if err := json.NewEncoder(w).Encode(result); err != nil {
		panic(err)
	}
}

func (s *metricSet) getMetrics(names []string, w io.Writer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, name := range names {
		if _, ok := s.history[name]; !ok {
			return errors.New("missing")
		}
	}

	err := binary.Write(w, binary.LittleEndian, s.timeMillis[s.tailIndex:])
	if err != nil {
		panic(err)
	}
	err = binary.Write(w, binary.LittleEndian, s.timeMillis[:s.tailIndex])
	if err != nil {
		panic(err)
	}

	for _, name := range names {
		hist := s.history[name]
		err = binary.Write(w, binary.LittleEndian, hist.values[s.tailIndex:])
		if err != nil {
			panic(err)
		}
		err = binary.Write(w, binary.LittleEndian, hist.values[:s.tailIndex])
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func (s *metricSet) watch(ctx context.Context, interval time.Duration) {
	tick := time.NewTicker(interval)
	prev := time.Now()

	hasHistory := false
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}

		if !hasHistory {
			s.mu.RLock()
			var ch chan struct{}
			if len(s.history) == 0 {
				ch = make(chan struct{})
				s.historyWatch = ch
			}
			s.mu.RUnlock()
			if ch != nil {
				tick.Stop()
				tick = nil
				select {
				case <-ctx.Done():
					return
				case <-ch:
				}
				tick = time.NewTicker(interval)
			}
			hasHistory = true
		}

		s.mu.Lock()
		now := time.Now()
		duration := float32(now.Sub(prev).Seconds())
		prev = now

		idx := s.tailIndex
		for k, hist := range s.history {
			if hist.expire.Before(now) {
				delete(s.history, k)
				continue
			}
			if hist.metric.Int32 == nil {
				continue
			}
			if hist.metric.isRate {
				val := float32(hist.metric.Swap(int32(0)))
				hist.values[idx] = val / duration
			} else {
				val := float32(hist.metric.Load())
				hist.values[idx] = val
			}
		}
		s.timeMillis[idx] = now.UnixNano() / int64(time.Millisecond)
		s.tailIndex = (idx + 1) % s.count

		if s.etagWatch != nil {
			close(s.etagWatch)
			s.etagWatch = nil
		}

		if len(s.history) == 0 {
			s.tailIndex = 0
			for i := range s.timeMillis {
				s.timeMillis[i] = 0
			}
			hasHistory = false
		}
		s.mu.Unlock()
	}
}

func (s *metricSet) attach(ids []string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expire := time.Now().Add(duration)
	for _, id := range ids {
		if hist, ok := s.history[id]; ok {
			hist.expire = expire
		} else {
			m := s.metrics[id]
			s.history[id] = &History{
				values: make([]float32, s.count),
				metric: m,
				expire: expire,
			}
		}
	}

	if s.historyWatch != nil {
		close(s.historyWatch)
		s.historyWatch = nil
	}
}

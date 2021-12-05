package flowdebug

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/channel"
)

type intercept struct {
	profiler *Profiler

	nodes []*logWrapper
	// connection name -> total elem count
	totalElem map[string]*int

	metricPrefix []string
	chanMon      []*chanMonitor
}

type logWrapper struct {
	flow.NodeWrapper
	stopped chan<- string
	name    string
	err     error
}

func (n *logWrapper) Run(ctx context.Context) error {
	pprof.Do(ctx, pprof.Labels("node", n.name), func(ctx context.Context) {
		n.err = n.NodeWrapper.Run(ctx)
		n.stopped <- n.name
	})
	return n.err
}

func GraphInterceptor(p *Profiler) flow.Interceptor {
	pp := &intercept{
		profiler:  p,
		totalElem: make(map[string]*int),
	}
	runtime.SetFinalizer(pp, func(p *intercept) {
		if p.profiler == nil {
			return
		}
		p.profiler.DeleteMetrics(p.metricPrefix)
	})
	return pp
}

func (p *intercept) AddNode(next flow.ChainInterceptor, name string, node flow.Node) error {
	w := &logWrapper{
		NodeWrapper: flow.WrapNode(node),
		name:        name,
	}
	p.nodes = append(p.nodes, w)
	return next.AddNode(name, w)
}

func (p *intercept) Connect(next flow.ChainInterceptor,
	id string, srcs, dsts []string, opts ...flow.ConnectOption,
) (flow.Chan, error) {
	rate := p.profiler.AddMetric(id+"/rate", true)
	p.metricPrefix = append(p.metricPrefix, id+"/")
	cnt := new(int)
	p.totalElem[id] = cnt

	opts = append(opts, channel.WithObserver(func(interface{}) {
		*cnt++
		rate.Inc()
	}))

	ch, err := next.Connect(id, srcs, dsts, opts...)
	if err != nil {
		return nil, err
	}

	p.chanMon = append(p.chanMon, newChanMonitor(
		ch, p.profiler.AddMetric(id+"/size", false)))
	return ch, err
}

func (p *intercept) Run(next flow.ChainInterceptor, ctx context.Context) error {
	log := p.profiler.opt.Logger
	g := next.Graph()
	topo := g.Topology()

	nodeStopped := make(chan string, len(topo.Nodes))
	for _, n := range p.nodes {
		n.stopped = nodeStopped
	}

	if err := p.profiler.AddGraph(topo.ID, topo, p.metricPrefix); err != nil {
		return err
	}
	go func() {
		update := func() {
			for _, m := range p.chanMon {
				m.Record()
			}
		}
		const TICK = 128 * time.Millisecond
		tick := time.NewTicker(TICK)
		defer tick.Stop()

		duration := p.profiler.opt.SampleRate
		if duration < TICK {
			duration = TICK
		}
		weight := 2 / float32(1+duration/TICK)

		deadline := time.Unix(0, 0)
		now := time.Now()
		for {
			for _, m := range p.chanMon {
				m.Sample(weight)
			}
			if now.After(deadline) {
				update()
				deadline = now.Add(duration)
			}
			select {
			case <-ctx.Done():
				update()
				return
			case now = <-tick.C:
			}
		}
	}()
	go func() {
		nodes := make(map[string]struct{})
		for _, v := range topo.Nodes {
			nodes[v.Name] = struct{}{}
		}
		delete(nodes, <-nodeStopped)

		tk := time.NewTicker(30 * time.Second)
		defer tk.Stop()

		begin := time.Now()
		for {
			select {
			case id, ok := <-nodeStopped:
				if !ok {
					return
				}
				delete(nodes, id)
			case t := <-tk.C:
				b := new(bytes.Buffer)
				fmt.Fprintf(b, "Graph `%s' still running for %ds after node exited, nodes active:", topo.ID, t.Sub(begin)/time.Second)
				for _, n := range nodes {
					fmt.Fprintf(b, "\n\t* %s", n)
				}
				log.Warn(b.String())
			}
		}
	}()

	start := time.Now()
	var err error
	pprof.Do(ctx, pprof.Labels("graph", g.ID()), func(ctx context.Context) {
		err = next.Run(ctx)
	})
	duration := time.Since(start).Seconds()
	p.profiler.RemoveGraph(g.ID(), err)

	close(nodeStopped)

	if err != nil {
		b := new(bytes.Buffer)
		fmt.Fprintf(b, "Graph `%s' error: %s", g.ID(), err)
		for _, n := range p.nodes {
			err := n.err
			if err != nil {
				for _, nn := range topo.Nodes {
					if nn.Name == n.name {
						fmt.Fprintf(b, "\n\t* node `%s' error: %s", nn.Name, err)
						break
					}
				}
			}
		}
		log.Warn(b.String())
	}

	if log.IsLevelEnabled(logrus.DebugLevel) {
		b := new(bytes.Buffer)
		fmt.Fprintf(b, "Graph `%s' summary\n", g.ID())
		w := tablewriter.NewWriter(b)
		w.SetHeader([]string{"From", "To", "Count", "Rate"})
		for _, c := range topo.Connections {
			w.Append([]string{
				strings.Join(c.Source, "\n"),
				strings.Join(c.Destination, "\n"),
				strconv.Itoa(*p.totalElem[c.ID]),
				fmt.Sprintf("%.2f/s", float64(*p.totalElem[c.ID])/duration),
			})
			if len(c.Source) > 1 || len(c.Destination) > 1 {
				w.SetRowLine(true)
			}
		}
		w.Render()
		log.Debug(b)
	}

	return err
}

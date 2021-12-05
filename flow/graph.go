package flow

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"go.uber.org/atomic"

	"github.com/brian14708/go-flow/flow/channel"
	"github.com/brian14708/go-flow/flow/internal/ident"
)

type connection struct {
	id       string
	src, dst []string
	ch       Chan
}

type chanKey struct {
	addr uintptr
	dir  reflect.ChanDir
}

type Graph struct {
	nodes      map[string]*nodeContainer
	nodeHashes map[interface{}]struct{}
	conns      []*connection

	background []func(context.Context) error
	chanMap    map[chanKey]*atomic.Int32

	opt         GraphOptions
	interceptor ChainInterceptor
	initialized bool
}

type GraphOptions struct {
	ID string

	Interceptors          []Interceptor
	DefaultConnectOptions []ConnectOption
}

func NewGraph(opt *GraphOptions) (*Graph, error) {
	if opt == nil {
		opt = new(GraphOptions)
	}
	if opt.ID == "" {
		opt.ID = ident.UniqueID()
	}

	g := &Graph{
		nodes:      make(map[string]*nodeContainer),
		nodeHashes: make(map[interface{}]struct{}),
		chanMap:    make(map[chanKey]*atomic.Int32),

		opt: *opt,
	}
	g.interceptor = newChainInterceptor(g, g.opt.Interceptors)
	g.opt.Interceptors = nil
	return g, nil
}

func (g *Graph) ID() string {
	return g.opt.ID
}

func (g *Graph) AddNode(name string, node Node) error {
	return g.interceptor.AddNode(name, node)
}

func (g *Graph) addNode(name string, node Node) error {
	if !ident.Check(name) {
		return errors.New("invalid name")
	}
	if _, ok := g.nodes[name]; ok {
		return errors.New("name already exists")
	}

	nc, err := newNodeContainer(name, node)
	if err != nil {
		return err
	}

	if _, ok := g.nodeHashes[nc.nodeHash]; ok {
		found := ""
		for name, node := range g.nodes {
			if node.nodeHash == nc.nodeHash {
				found = name
			}
		}
		return fmt.Errorf("duplicate node, already added as `%s'", found)
	}

	g.nodes[name] = nc
	g.nodeHashes[nc.nodeHash] = struct{}{}
	return nil
}

func (g *Graph) Run(ctx context.Context) error {
	if g.initialized {
		return errors.New("graph already initialized")
	}
	g.initialized = true

	return g.interceptor.Run(ctx)
}

func (g *Graph) run(origCtx context.Context) error {
	var (
		gctx = newGraphContext(origCtx, g)
		wg   sync.WaitGroup

		errCh    = make(chan GraphError, 2)
		out2node = make(map[channel.Channel][]*nodeContainer)
	)
	defer func() {
		wg.Wait()
		gctx.Cancel()
	}()

	for _, n := range g.nodes {
		n.setContext(gctx)

		cnt := 0
		for _, o := range n.out {
			if o.ch != nil {
				cnt++
				out2node[o.ch] = append(out2node[o.ch], n)
			}
		}
		n.numActiveOut.Store(int32(cnt))
	}

	for _, n := range g.nodes {
		n := n
		wg.Add(1)
		go func() {
			defer func() {
				n.stop()
				wg.Done()
			}()

			err := n.run()
			for _, port := range n.in {
				ptr := port.addr()
				if ptr == 0 {
					continue
				}
				if g.chanMap[chanKey{ptr, reflect.RecvDir}].Dec() != 0 {
					continue
				}

				ch := port.ch
				wg.Add(1)
				go func() {
					defer wg.Done()
					ch.Drain()
				}()

				for _, o := range out2node[ch] {
					// all out port of a certain node is draining
					if o.numActiveOut.Dec() == 0 {
						// use -1 as special marker
						o.numActiveOut.Store(-1)
						o.stop()
					}
				}
			}
			for _, port := range n.out {
				ptr := port.addr()
				if ptr == 0 {
					continue
				}
				if g.chanMap[chanKey{ptr, reflect.SendDir}].Dec() == 0 {
					port.ch.Close()
				}
			}

			if n.numActiveOut.Load() == -1 {
				// ignore error if stop is triggered by downstream
				err = nil
			}
			errCh <- GraphError{
				name: n.name,
				err:  err,
			}
		}()
	}

	for _, bg := range g.background {
		bg := bg
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- GraphError{
				err: bg(gctx),
			}
		}()
	}

	var outErr []GraphError

	pendingNode := len(g.nodes)
	pending := pendingNode + len(g.background)
	for pending > 0 {
		select {
		case <-origCtx.Done():
			gctx.Cancel()
			// drain
			for i := 0; i < pending; i++ {
				<-errCh
			}
			return origCtx.Err()
		case err := <-errCh:
			pending--
			if err.name != "" {
				pendingNode--
			}

			if err.err != nil {
				gctx.Cancel()
				outErr = append(outErr, err)
			}
		}

		if pendingNode == 0 {
			// cancel background
			gctx.Cancel()
		}
	}

	if origCtx.Err() != nil {
		return origCtx.Err()
	}
	if len(outErr) == 0 {
		return nil
	}
	// chain error
	it := &outErr[0]
	for _, e := range outErr[1:] {
		it.next = &e
		it = it.next
	}
	return &outErr[0]
}

package flow

import "context"

var (
	ctxKey  int
	gctxKey int
)

type NodeExecContext interface {
	GetChan(s string) Chan
	NodeName() string
}

type GraphExecContext interface {
	GraphID() string
}

func FromContext(ctx context.Context) (GraphExecContext, NodeExecContext) {
	n, _ := ctx.Value(&ctxKey).(NodeExecContext)
	g, _ := ctx.Value(&gctxKey).(GraphExecContext)
	return g, n
}

// each port has a buffered channel that actually contains pending elements.
func (nc *nodeContainer) GetChan(s string) Chan {
	for _, n := range nc.in {
		if n.name == s {
			return n.ch
		}
	}
	for _, n := range nc.out {
		if n.name == s {
			return n.ch
		}
	}
	return nil
}

func (nc *nodeContainer) NodeName() string {
	return nc.name
}

func (g *Graph) GraphID() string {
	return g.opt.ID
}

type nodeContext struct {
	context.Context
	Cancel context.CancelFunc
	n      NodeExecContext
}

func (n *nodeContext) Value(key interface{}) interface{} {
	if key == &ctxKey {
		return n.n
	}
	return n.Context.Value(key)
}

func newNodeContext(ctx context.Context, nc NodeExecContext) *nodeContext {
	cctx, ccancel := context.WithCancel(ctx)
	return &nodeContext{
		Context: cctx,
		Cancel:  ccancel,
		n:       nc,
	}
}

type graphContext struct {
	context.Context
	Cancel context.CancelFunc
	g      GraphExecContext
}

func (n *graphContext) Value(key interface{}) interface{} {
	if key == &gctxKey {
		return n.g
	}
	return n.Context.Value(key)
}

func newGraphContext(ctx context.Context, gc GraphExecContext) *graphContext {
	cctx, ccancel := context.WithCancel(ctx)
	return &graphContext{
		Context: cctx,
		Cancel:  ccancel,
		g:       gc,
	}
}

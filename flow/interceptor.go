package flow

import (
	"context"
)

type Interceptor interface {
	AddNode(next ChainInterceptor, name string, node Node) error
	Connect(next ChainInterceptor,
		id string, srcs, dsts []string, opts ...ConnectOption,
	) (Chan, error)
	Run(next ChainInterceptor, ctx context.Context) error
}

type ChainInterceptor []Interceptor

func newChainInterceptor(g *Graph, interceptors []Interceptor) ChainInterceptor {
	chain := make([]Interceptor, 0, len(interceptors)+1)
	for _, i := range interceptors {
		if i != nil {
			chain = append(chain, i)
		}
	}
	chain = append(chain, (*graphCaller)(g))
	return chain
}

func (c ChainInterceptor) Graph() *Graph {
	return (*Graph)(c[len(c)-1].(*graphCaller))
}

func (c ChainInterceptor) AddNode(name string, node Node) error {
	return c[0].AddNode(c[1:], name, node)
}

func (c ChainInterceptor) Connect(
	id string, srcs, dsts []string, opts ...ConnectOption,
) (Chan, error) {
	return c[0].Connect(c[1:], id, srcs, dsts, opts...)
}

func (c ChainInterceptor) Run(ctx context.Context) error {
	return c[0].Run(c[1:], ctx)
}

//

type NoopInterceptor struct{}

func (NoopInterceptor) AddNode(next ChainInterceptor, name string, node Node) error {
	return next.AddNode(name, node)
}

func (NoopInterceptor) Connect(next ChainInterceptor,
	id string, srcs, dsts []string, opts ...ConnectOption,
) (Chan, error) {
	return next.Connect(id, srcs, dsts, opts...)
}

func (NoopInterceptor) Run(next ChainInterceptor, ctx context.Context) error {
	return next.Run(ctx)
}

//

type graphCaller Graph

func (g *graphCaller) AddNode(next ChainInterceptor, name string, node Node) error {
	return (*Graph)(g).addNode(name, node)
}

func (g *graphCaller) Connect(next ChainInterceptor,
	id string, srcs, dsts []string, opts ...ConnectOption,
) (Chan, error) {
	return (*Graph)(g).connect(id, srcs, dsts, opts...)
}

func (g *graphCaller) Run(next ChainInterceptor, ctx context.Context) error {
	return (*Graph)(g).run(ctx)
}

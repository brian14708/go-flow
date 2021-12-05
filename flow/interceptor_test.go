package flow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testInterceptor struct {
	NoopInterceptor
	cnt int
	g   *Graph
}

func (t *testInterceptor) AddNode(next ChainInterceptor, name string, node Node) error {
	t.cnt++
	t.g = next.Graph()
	return next.AddNode(name, node)
}

func TestInterceptor(t *testing.T) {
	i := new(testInterceptor)
	g, _ := NewGraph(&GraphOptions{
		Interceptors: []Interceptor{i, nil, NoopInterceptor{}},
	})
	out := make(chan int, 2)
	assert.NoError(t, g.AddNode("g1", GeneratorNode(1)))
	assert.NoError(t, g.AddNode("sum1", SumNode(out)))
	assert.NoError(t, g.Connect([]string{"g1:out"}, []string{"sum1:in"}))
	assert.Equal(t, 2, i.cnt)
	assert.Equal(t, g, i.g)
	assert.NoError(t, g.Run(context.Background()))
}

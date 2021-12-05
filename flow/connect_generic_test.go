package flow

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow/port"
	"github.com/brian14708/go-flow/flowtype/testutil"
)

type tplNode struct {
	in, out *Port
}

func (n *tplNode) Ports() (in, out PortMap) {
	if n.in != nil {
		in = PortMap{"in": n.in}
	}
	if n.out != nil {
		out = PortMap{"out": n.out}
	}
	return
}

func (*tplNode) Run(context.Context) error { return nil }

func TestTemplateNode(t *testing.T) {
	g, err := NewGraph(nil)
	assert.NoError(t, err)

	_ = g.AddNode("a", GeneratorNode(1))
	assert.Error(t, g.AddNode("b", &tplNode{
		port.TemplatePort(new(chan int), reflect.TypeOf("")),
		nil,
	}))
	assert.Error(t, g.AddNode("c", &tplNode{
		nil,
		port.TemplatePort(new(chan int), reflect.TypeOf("")),
	}))
	assert.NoError(t, g.AddNode("b", &tplNode{
		port.TemplatePort(new(chan interface{}), reflect.TypeOf(0)),
		nil,
	}))
	assert.NoError(t, g.AddNode("c", &tplNode{
		nil,
		port.TemplatePort(new(chan interface{}), reflect.TypeOf(0)),
	}))
	assert.NoError(t, g.Connect([]string{"c:out"}, []string{"a:in"}))
	assert.NoError(t, g.Connect([]string{"a:out"}, []string{"b:in"}))
}

func TestAnyMessage(t *testing.T) {
	testutil.TestWithNilRegistry(t, func(t *testing.T) {
		out := make(chan int, 5)
		g, _ := NewGraph(nil)
		cnt := 0
		assert.NoError(t, g.AddNode("g1", GeneratorNode(1)))
		assert.NoError(t, g.AddNode("w1", FnNode(
			func(_ context.Context, in <-chan int, out chan<- int) error {
				for i := range in {
					cnt++
					out <- i
				}
				return nil
			},
		)))
		assert.NoError(t, g.AddNode("sum1", SumNode(out)))
		assert.NoError(t, g.Connect([]string{"g1:out"}, []string{"w1:in"}))
		assert.NoError(t, g.Connect([]string{"w1:out"}, []string{"sum1:in"}))
		assert.NoError(t, g.Run(context.Background()))
		assert.Equal(t, 5, cnt)
	})
}

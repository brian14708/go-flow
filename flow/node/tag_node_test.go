package node

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flowtype"
)

type runnable struct{}

func (*runnable) Run(context.Context) error {
	return nil
}

func TestTagNode(t *testing.T) {
	type singleChan struct {
		In <-chan int `flow:"in"`
	}
	type recursion struct {
		R *recursion
	}
	rr := new(recursion)
	rr.R = rr

	testcase := [...]struct {
		success bool
		in, out int
		n       interface{}
	}{
		{false, 0, 0, new(struct{})},
		{false, 0, 0, new(struct {
			runnable
		})},
		{false, 0, 0, &struct {
			runnable
			R *recursion
		}{R: rr}},
		{false, 0, 0, new(struct {
			runnable
			BadType int `flow:"in"`
		})},
		{false, 0, 0, new(struct {
			runnable
			BiChannel chan int `flow:"in"`
		})},
		{false, 0, 0, new(struct {
			runnable
			In  <-chan int `flow:"in"`
			InX <-chan int `flow:"in"`
		})},
		{false, 0, 0, new(struct {
			runnable
			Out  chan<- int `flow:"out"`
			OutX chan<- int `flow:"out"`
		})},
		{true, 1, 0, new(struct {
			runnable
			In   <-chan int `flow:"in"`
			OutX chan<- int `flow:""` // empty
		})},
		{true, 0, 1, new(struct {
			runnable
			Out chan<- int `flow:"out"`
		})},
		{false, 0, 0, new(struct {
			runnable
			*singleChan
		})},
		{true, 1, 0, &struct {
			runnable
			*singleChan
		}{singleChan: new(singleChan)}},
	}

	for _, test := range testcase {
		n, err := NewTagNode(test.n, "flow")
		if test.success {
			assert.NoError(t, err)
			in, out := n.Ports()
			assert.Equal(t, len(in), test.in)
			assert.Equal(t, len(out), test.out)
			assert.Equal(t, reflect.TypeOf(test.n).String(), n.(interface {
				NodeType() string
			}).NodeType())
		} else {
			assert.Error(t, err)
		}
	}
}

func TestTagNodeEqual(t *testing.T) {
	s := new(struct {
		runnable
		In <-chan int `flow:"in"`
	})

	n1, err := NewTagNode(s, "flow")
	assert.NoError(t, err)
	n2, err := NewTagNode(s, "flow")
	assert.NoError(t, err)
	assert.Equal(t, n1.NodeHash(), n2.NodeHash())
}

type tagNodeFn struct {
	runnable
	In      <-chan int `flow:"in"`
	_, _, _ int
}

func (tagNodeFn) Fn()              {}
func (tagNodeFn) NodeType() string { return "T" }

func TestTagNodeWrap(t *testing.T) {
	n, err := NewTagNode(new(tagNodeFn), "flow")
	assert.NoError(t, err)

	var fn interface{ Fn() }
	assert.True(t, flowtype.As(n, &fn))

	assert.Equal(t, "T", n.(interface {
		NodeType() string
	}).NodeType())
}

func BenchmarkTag(b *testing.B) {
	t := new(tagNodeFn)
	for i := 0; i < b.N; i++ {
		_, _ = NewTagNode(t, "flow")
	}
}

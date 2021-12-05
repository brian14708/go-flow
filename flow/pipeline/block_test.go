package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/internal/ident"
)

func TestMakeBlock(t *testing.T) {
	var id ident.Generator
	ppl := New(nil)

	_, err := ppl.addBlock(id.Generate(""), new(SimpleNode))
	assert.NoError(t, err)

	{
		blk, err := ppl.addBlock(id.Generate(""), (<-chan int)(make(chan int)))
		assert.NoError(t, err)
		blk2, err := ppl.addBlock(id.Generate(""), blk)
		assert.NoError(t, err)
		assert.Equal(t, blk, blk2)
	}

	// bad nodes
	_, err = ppl.addBlock(id.Generate(""), "bad type")
	assert.Error(t, err)
	_, err = ppl.addBlock(id.Generate(""), new(BadTagNode))
	assert.Error(t, err)
	_, err = ppl.addBlock(id.Generate(""), make(chan int))
	assert.Error(t, err)
	_, err = ppl.addBlock(id.Generate(""), func() {})
	assert.Error(t, err)
	errMakerFail := errors.New("maker failed")
	_, err = ppl.addBlock(id.Generate(""), func() (flow.Node, error) {
		return nil, errMakerFail
	})
	assert.Equal(t, errMakerFail, err)

	_, err = ppl.addBlock(id.Generate(""), New(nil))
	assert.Error(t, err)
}

type BadTagNode struct {
	In  <-chan int `pipeline:"in"`
	InX <-chan int `pipeline:"in"`
}

func (*BadTagNode) Run(context.Context) error { return nil }

type SimpleNode struct {
	In  <-chan int
	Out chan<- int
}

func (s *SimpleNode) Ports() (flow.PortMap, flow.PortMap) {
	return flow.PortMap{
			"in": &s.In,
		}, flow.PortMap{
			"out": &s.Out,
		}
}

func (s *SimpleNode) Run(context.Context) error { return nil }

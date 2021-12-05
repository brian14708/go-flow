package flow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testNode struct{}

func (testNode) Run(context.Context) error { return nil }
func (testNode) Ports() (in, out PortMap)  { return nil, nil }
func (testNode) Fn() int                   { return 99 }

func TestNodeDedup(t *testing.T) {
	g, _ := NewGraph(nil)
	var n Node = testNode{}
	assert.NoError(t, g.AddNode("n", n))

	w := WrapNode(WrapNode(n))
	// duplicate
	assert.NotEqual(t, n, w)
	assert.Error(t, g.AddNode("w", w))

	var fnVal interface{ Fn() int }
	assert.True(t, w.As(&fnVal))
	assert.Equal(t, 99, fnVal.Fn())
}

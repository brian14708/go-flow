package flow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeContext(t *testing.T) {
	cc, _ := newNodeContainer("abcd", new(testNode))
	ctx := context.Background()
	_, fc := FromContext(ctx)
	assert.Nil(t, fc)
	nctx := newNodeContext(ctx, cc)
	_, fc = FromContext(nctx)
	assert.NotNil(t, fc)
	assert.Equal(t, fc.NodeName(), cc.NodeName())
}

func TestGraphContext(t *testing.T) {
	cc, _ := NewGraph(nil)
	ctx := context.Background()
	fc, _ := FromContext(ctx)
	assert.Nil(t, fc)
	gctx := newGraphContext(ctx, cc)
	fc, _ = FromContext(gctx)
	assert.NotNil(t, fc)
	assert.Equal(t, fc.GraphID(), cc.GraphID())
}

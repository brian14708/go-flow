package flowutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContextChecker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := NewContextChecker(ctx)
	assert.True(t, cc.Valid())
	assert.Nil(t, cc.Err())
	cancel()
	for cc.Valid() {
		time.Sleep(time.Millisecond)
	}
	assert.False(t, cc.Valid())
	assert.Equal(t, cc.Err(), ctx.Err())
}

func BenchmarkContextChecker(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b.RunParallel(func(pb *testing.PB) {
		cc := NewContextChecker(ctx)
		for pb.Next() {
			_ = cc.Valid()
		}
	})
}

package token

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenQueue(t *testing.T) {
	var ret []int
	var wg sync.WaitGroup
	q := NewTokenQueue(0, true)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		n, err := q.Acquire(context.Background())
		assert.NoError(t, err)
		i := i
		go func() {
			defer wg.Done()

			time.Sleep(time.Duration(rand.Intn(25)) * time.Millisecond)
			assert.NoError(t, n.Wait(context.Background()))
			assert.NoError(t, n.Wait(context.Background()))
			assert.NoError(t, n.Wait(context.Background()))
			assert.NoError(t, n.Wait(context.Background()))
			ret = append(ret, i)
			n.Release()
		}()
	}
	wg.Wait()
	for i, r := range ret {
		assert.Equal(t, i, r)
	}
}

func TestTokenQueueContext(t *testing.T) {
	q := NewTokenQueue(2, true)
	ctx := context.Background()

	_, err := q.Acquire(ctx)
	assert.NoError(t, err)
	w, err := q.Acquire(ctx)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.Equal(t, ctx.Err(), w.Wait(ctx))
	_, err = q.Acquire(ctx)
	assert.Equal(t, ctx.Err(), err)
}

func TestTokenQueueError(t *testing.T) {
	q := NewTokenQueue(2, false)
	token, _ := q.Acquire(context.Background())
	token.Release()

	q = NewTokenQueue(2, true)
	token, _ = q.Acquire(context.Background())
	assert.Panics(t, func() {
		token.Release()
	})
}

func BenchmarkTokenQueue(b *testing.B) {
	q := NewTokenQueue(32, true)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		w, _ := q.Acquire(ctx)
		go func() {
			_ = w.Wait(ctx)
			w.Release()
		}()
	}
}

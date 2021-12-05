package token

import (
	"context"
	"sync"
)

type TokenQueue struct {
	parallel chan struct{}
	tail     chan struct{}
}

type Token struct {
	parallel <-chan struct{}
	curr     chan struct{}
	next     chan<- struct{}
}

var (
	chanPool = sync.Pool{
		New: func() interface{} {
			return make(chan struct{}, 1)
		},
	}
	tokenPool = sync.Pool{
		New: func() interface{} {
			return new(Token)
		},
	}
)

func NewTokenQueue(maxParallel int, serialize bool) *TokenQueue {
	var c chan struct{}
	if serialize {
		c = chanPool.Get().(chan struct{})
		c <- struct{}{}
	}

	var parallel chan struct{}
	if maxParallel > 0 {
		parallel = make(chan struct{}, maxParallel)
	}

	return &TokenQueue{
		parallel: parallel,
		tail:     c,
	}
}

func (q *TokenQueue) Acquire(ctx context.Context) (*Token, error) {
	if q.parallel != nil {
		select {
		case q.parallel <- struct{}{}:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	t := tokenPool.Get().(*Token)
	t.parallel = q.parallel
	if q.tail != nil {
		t.curr = q.tail
		q.tail = chanPool.Get().(chan struct{})
		t.next = q.tail
	}
	return t, nil
}

func (w *Token) Wait(ctx context.Context) error {
	if w.curr != nil {
		select {
		case <-w.curr:
			chanPool.Put(w.curr)
			w.curr = nil
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (w *Token) Release() {
	if w.curr != nil {
		panic("Token.Wait must be called before release")
	}
	if w.next != nil {
		w.next <- struct{}{}
		w.next = nil
	}
	if w.parallel != nil {
		<-w.parallel
		w.parallel = nil
	}
	tokenPool.Put(w)
}

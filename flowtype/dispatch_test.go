package flowtype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenericDispatch(t *testing.T) {
	type M struct {
		m int
	}

	t.Run("Send", func(t *testing.T) {
		ch := make(chan M, 3)
		done := make(chan struct{})
		s := GenericDispatch.ChanSender(ch)
		s(M{2}, nil, true)
		s(M{2}, nil, true)

		// nonblock & cancel
		assert.True(t, s(nil, done, true))
		assert.False(t, s(nil, nil, false))
		assert.False(t, s(nil, done, false))
		go close(done)
		assert.False(t, s(nil, done, true))

		assert.Panics(t, func() {
			s(1, nil, true)
		})
	})

	t.Run("Recv", func(t *testing.T) {
		ch := make(chan M, 1)
		done := make(chan struct{})
		r := GenericDispatch.ChanRecver(ch)

		ch <- M{2}
		m, ok := r(nil, true)
		assert.True(t, ok)
		assert.Equal(t, 2, m.(M).m)

		{
			ch <- M{2}
			m, ok = r(nil, false)
			assert.True(t, ok)
			assert.Equal(t, 2, m.(M).m)

			ch <- M{2}
			m, ok = r(done, false)
			assert.True(t, ok)
			assert.Equal(t, 2, m.(M).m)

			_, ok = r(done, false)
			assert.False(t, ok)
			go close(done)
			_, ok = r(done, true)
			assert.False(t, ok)

			ch <- M{2}
			for ok = false; !ok; {
				_, ok = r(done, true)
			}
			assert.True(t, ok)
		}

		close(ch)
		_, ok = r(nil, true)
		assert.False(t, ok)
	})

	t.Run("Error", func(t *testing.T) {
		ch := make(chan M, 3)
		assert.Panics(t, func() {
			GenericDispatch.ChanSender((<-chan M)(ch))
		})
		assert.Panics(t, func() {
			GenericDispatch.ChanRecver((chan<- M)(ch))
		})
	})

	t.Run("FuncCaller", func(t *testing.T) {
		fn := GenericDispatch.FuncCaller(func(m M) M {
			m.m++
			return m
		})
		m := fn([]interface{}{M{2}}, nil)[0]
		assert.Equal(t, 3, m.(M).m)
		m = fn([]interface{}{nil}, nil)[0]
		assert.Equal(t, 1, m.(M).m)

		assert.Panics(t, func() {
			GenericDispatch.FuncCaller(123)
		})
	})
}

func BenchmarkGenericDispatch(b *testing.B) {
	ch := make(chan int, 10)
	send := GenericDispatch.ChanSender(ch)
	recv := GenericDispatch.ChanRecver(ch)
	ch <- 1
	done := make(chan struct{})
	for i := 0; i < b.N; i++ {
		v, _ := recv(done, true)
		send(v, done, true)
	}
}

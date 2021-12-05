package channel

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChanAdaptive(t *testing.T) {
	i := reflect.TypeOf(0)
	ch, err := newAdaptiveChan(i, i, &options{
		size:         32,
		adaptiveGain: 1,
	})
	assert.NoError(t, err)
	assert.Equal(t, 32, ch.Cap())

	var snd, rcv chan int
	ch.AssignTo(reflect.SendDir, &snd)
	ch.AssignTo(reflect.RecvDir, &rcv)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		ch.Serve()
		wg.Done()
	}()

	go func() {
		for i := 0; i < 64; i++ {
			snd <- i
		}
		close(snd)
	}()
	for i := 0; i < 64; i++ {
		assert.Equal(t, i, <-rcv)
	}
	wg.Wait()
}

func BenchmarkChanAdaptive(b *testing.B) {
	i := reflect.TypeOf(0)
	ch, _ := newAdaptiveChan(i, i, &options{
		size:         32,
		adaptiveGain: 1,
	})

	var snd, rcv chan int
	ch.AssignTo(reflect.SendDir, &snd)
	ch.AssignTo(reflect.RecvDir, &rcv)
	go ch.Serve()
	for i := 0; i < b.N; i++ {
		snd <- 1
		<-rcv
	}
	close(snd)
}

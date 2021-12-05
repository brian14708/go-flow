package flowdebug

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
)

func TestChanMonitor(t *testing.T) {
	ch := make(chan int, 100)
	m := Metric{Int32: atomic.NewInt32(0)}
	mon := newChanMonitor(reflect.ValueOf(ch), m)
	for i := 0; i < cap(ch); i++ {
		ch <- i
		mon.Sample(1)
	}
	mon.Record()
	assert.Equal(t, 100, int(m.Load()))
}

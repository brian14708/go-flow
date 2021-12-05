package funcop

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/node"
)

func TestMap(t *testing.T) {
	assert.Panics(t, func() {
		Map(func() int { return 0 })
	})
	assert.Panics(t, func() {
		Map(func(int) int { return 0 }, WithParallel(-1, false))
	})

	m := Map(func(int) float32 {
		return 0
	}, WithParallel(2, true))

	in, out := m.Ports()
	assert.Equal(t, 1, len(in))
	assert.Equal(t, 1, len(out))
	assert.Contains(t, m.(interface {
		NodeType() string
	}).NodeType(), "Mapper")
	assert.Contains(t, m.(interface {
		Description() string
	}).Description(), "Ordered")
}

func TestMapRun(t *testing.T) {
	err := errors.New("error")
	testcase := [...]struct {
		success bool
		node    flow.Node
	}{
		{true, Map(func(int) {})},
		{true, Map(func(int) float32 {
			return 99
		})},
		{true, Map(func(context.Context, int) (float32, error) {
			return 99, nil
		})},
		{true, Map(func(context.Context, int) error {
			return nil
		})},
		{false, Map(func(context.Context, int) (float32, error) {
			return 0, err
		})},
		{false, Map(func(context.Context, int) error {
			return err
		})},
	}

	for _, test := range testcase {
		in, out := make(chan interface{}, 10), make(chan interface{}, 10)
		inPort, outPort := test.node.Ports()
		setPort(inPort["in"], in)
		if len(outPort) > 0 {
			setPort(outPort["out"], out)
		} else {
			close(out)
		}
		for i := 0; i < 10; i++ {
			in <- i
		}
		close(in)
		if test.success {
			assert.NoError(t, test.node.Run(context.Background()))
			if len(outPort) > 0 {
				for i := 0; i < 10; i++ {
					assert.Equal(t, float32(99), <-out)
				}
			}
		} else {
			assert.Equal(t, err, test.node.Run(context.Background()))
		}
	}
}

type dropper struct {
	atomic.Int32
}

func (d *dropper) DropMessage(interface{}) {
	d.Inc()
}

func TestMapError(t *testing.T) {
	var wg sync.WaitGroup
	m := Map(func(ctx context.Context, i int) (int, error) {
		wg.Done()
		wg.Wait()
		if i == 2 {
			return 0, errors.New("err")
		}
		return i, nil
	}, WithParallel(10, true))

	d := &dropper{}
	m.(*mapper).outChan = d

	in, out := make(chan interface{}, 10), make(chan interface{}, 10)
	for i := 0; i < 10; i++ {
		in <- i
		wg.Add(1)
	}

	inPort, outPort := m.Ports()
	setPort(inPort["in"], in)
	setPort(outPort["out"], out)

	assert.Error(t, m.Run(context.Background()))
	assert.Equal(t, int32(7), d.Load())
}

func TestMapGraph(t *testing.T) {
	g, _ := flow.NewGraph(nil)
	in, out := make(chan int, 8), make(chan int, 8)
	inN, _ := node.NewChanNode((<-chan int)(in))
	outN, _ := node.NewChanNode((chan<- int)(out))

	_ = g.AddNode("m", Map(func(i int) int {
		return i * 2
	}))
	_ = g.AddNode("in", inN)
	_ = g.AddNode("out", outN)
	_ = g.Connect([]string{"in:out"}, []string{"m:in"})
	_ = g.Connect([]string{"m:out"}, []string{"out:in"})

	in <- 10
	close(in)
	assert.NoError(t, g.Run(context.Background()))
	assert.Equal(t, 20, <-out)
}

func BenchmarkMap(b *testing.B) {
	for _, cnt := range []int{1, 4} {
		b.Run(fmt.Sprintf("%d", cnt), func(b *testing.B) {
			node := Map(func(i int) int {
				return i + 2
			}, WithParallel(cnt, false))
			in, out := make(chan interface{}, 10), make(chan interface{}, 10)
			inPort, outPort := node.Ports()
			setPort(inPort["in"], in)
			setPort(outPort["out"], out)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() { _ = node.Run(ctx) }()

			for i := 0; i < b.N; i++ {
				in <- 1
				<-out
			}
		})
	}
}

func setPort(p interface{}, val interface{}) {
	reflect.ValueOf(p.(*flow.Port).Ref).Elem().Set(reflect.ValueOf(val))
}

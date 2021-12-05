package funcop

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow"
)

func TestFlatMap(t *testing.T) {
	assert.Panics(t, func() {
		FlatMap(func() int { return 0 })
	})
	assert.Panics(t, func() {
		FlatMap(func(int) int { return 0 })
	})

	m := FlatMap(func(int) []float32 {
		return nil
	}, WithParallel(2, true))

	in, out := m.Ports()
	assert.Equal(t, 1, len(in))
	assert.Equal(t, 1, len(out))
	assert.Contains(t, m.(interface {
		NodeType() string
	}).NodeType(), "FlatMapper")
	assert.Contains(t, m.(interface {
		Description() string
	}).Description(), "Ordered")
}

func TestFlatMapRun(t *testing.T) {
	err := errors.New("error")
	testcase := [...]struct {
		success bool
		node    flow.Node
	}{
		{true, FlatMap(func(i int) []float32 {
			f := float32(i)
			return []float32{f, 2 * f}
		})},
		{true, FlatMap(func(_ context.Context, i int) ([]float32, error) {
			f := float32(i)
			return []float32{f, 2 * f}, nil
		})},
		{false, FlatMap(func(context.Context, int) ([]float32, error) {
			return nil, err
		})},
	}

	for _, test := range testcase {
		in, out := make(chan interface{}, 10), make(chan interface{}, 20)
		inPort, outPort := test.node.Ports()
		setPort(inPort["in"], in)
		setPort(outPort["out"], out)
		for i := 0; i < 10; i++ {
			in <- i
		}
		close(in)
		if test.success {
			assert.NoError(t, test.node.Run(context.Background()))
			for i := 0; i < 10; i++ {
				assert.Equal(t, float32(i), <-out)
				assert.Equal(t, 2*float32(i), <-out)
			}
		} else {
			assert.Error(t, test.node.Run(context.Background()))
		}
	}
}

func TestFlatMapError(t *testing.T) {
	var wg sync.WaitGroup
	f := FlatMap(func(ctx context.Context, i int) ([]int, error) {
		wg.Done()
		wg.Wait()
		if i == 2 {
			return nil, errors.New("err")
		}
		return []int{i}, nil
	}, WithParallel(10, true))

	d := &dropper{}
	f.(*flatMapper).outChan = d

	in, out := make(chan interface{}, 10), make(chan interface{}, 10)
	for i := 0; i < 10; i++ {
		in <- i
		wg.Add(1)
	}

	inPort, outPort := f.Ports()
	setPort(inPort["in"], in)
	setPort(outPort["out"], out)

	assert.Error(t, f.Run(context.Background()))
	assert.Equal(t, int32(7), d.Load())
}

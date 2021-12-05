package funcop

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow"
)

func TestFilter(t *testing.T) {
	assert.Panics(t, func() {
		Filter(func(int) int { return 0 })
	})

	f := Filter(func(int) bool {
		return true
	}, WithParallel(2, true))

	in, out := f.Ports()
	assert.Equal(t, 1, len(in))
	assert.Equal(t, 1, len(out))
	assert.Contains(t, f.(interface {
		NodeType() string
	}).NodeType(), "Filter")
	assert.Contains(t, f.(interface {
		Description() string
	}).Description(), "Ordered")
}

func TestFilterRun(t *testing.T) {
	err := errors.New("error")
	testcase := [...]struct {
		success bool
		node    flow.Node
	}{
		{true, Filter(func(i int) bool { return i%2 == 0 })},
		{true, Filter(func(_ context.Context, i int) (bool, error) {
			return i%2 == 0, nil
		})},
		{false, Filter(func(context.Context, int) (bool, error) {
			return false, err
		})},
	}
	for _, test := range testcase {
		in, out := make(chan interface{}, 10), make(chan interface{}, 10)
		inPort, outPort := test.node.Ports()
		setPort(inPort["in"], in)
		setPort(outPort["out"], out)

		for i := 0; i < 10; i++ {
			in <- i
		}
		close(in)

		if test.success {
			assert.NoError(t, test.node.Run(context.Background()))
			for i := 0; i < 5; i++ {
				assert.Equal(t, 2*i, <-out)
			}
		} else {
			assert.Equal(t, err, test.node.Run(context.Background()))
		}
	}
}

func TestFilterError(t *testing.T) {
	var wg sync.WaitGroup
	f := Filter(func(ctx context.Context, i int) (bool, error) {
		wg.Done()
		wg.Wait()
		if i == 2 {
			return false, errors.New("err")
		}
		return true, nil
	}, WithParallel(10, true))

	d := &dropper{}
	f.(*filter).outChan = d

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

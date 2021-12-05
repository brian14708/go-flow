package node

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeFuncNode(t *testing.T) {
	testcase := [...]struct {
		success bool
		in, out int
		fn      interface{}
	}{
		{false, 0, 0, "badType"},
		{false, 0, 0, func(context.Context) {}},
		{false, 0, 0, func(context.Context, chan<- int) int { return 0 }},
		{false, 0, 0, func(context.Context, chan<- int) (int, int) { return 0, 0 }},
		{false, 0, 0, func(int, chan<- int) error { return nil }},
		{false, 0, 0, func(context.Context, int) error { return nil }},
		{false, 0, 0, func(context.Context, ...int) error { return nil }},
		{false, 0, 0, func(context.Context, int, int, int) error { return nil }},

		{false, 0, 0, func(context.Context, chan int) error { return nil }},
		{false, 0, 0, func(context.Context, <-chan int, <-chan int) error { return nil }},
		{false, 0, 0, func(context.Context, chan<- int, chan<- int) error { return nil }},

		{false, 0, 0, func(context.Context, chan<- int) {}},
		{false, 0, 0, func(context.Context, <-chan int) {}},
		{true, 1, 0, func(<-chan int) {}},
		{true, 0, 1, func(context.Context, chan<- int) error { return nil }},
		{true, 1, 0, func(context.Context, <-chan int) error { return nil }},
		{true, 1, 1, func(context.Context, <-chan int, chan<- int) error { return nil }},
		{false, 0, 0, func(context.Context, chan<- int, <-chan int) error { return nil }},
	}
	for _, test := range testcase {
		n, err := NewFuncNode(test.fn)
		if !test.success {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		in, out := n.Ports()
		assert.Equal(t, test.in, len(in))
		assert.Equal(t, test.out, len(out))
		assert.Equal(t, reflect.TypeOf(test.fn).String(), n.(interface {
			NodeType() string
		}).NodeType())
	}
}

func TestFuncNodeRun(t *testing.T) {
	fnErr := errors.New("error")
	testcase := [...]struct {
		success bool
		fn      interface{}
	}{
		{true, func(context.Context, <-chan int, chan<- int) error { return nil }},
		{true, func(context.Context, <-chan int) error { return nil }},
		{true, func(context.Context, chan<- int) error { return nil }},
		{true, func(<-chan int) {}},
		{true, func(<-chan int, chan<- int) {}},
		{false, func(context.Context, <-chan int, chan<- int) error { return fnErr }},
		{false, func(context.Context, <-chan int) error { return fnErr }},
		{false, func(context.Context, chan<- int) error { return fnErr }},
	}
	for _, test := range testcase {
		fn, err := NewFuncNode(test.fn)
		assert.NoError(t, err)
		err = fn.Run(context.Background())
		if test.success {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, fnErr, err)
		}
	}
}

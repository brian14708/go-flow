package flowtype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fn struct{}

func (fn) Fn() {}

type converter struct {
	v interface{}
}

func (c converter) As(t interface{}) bool {
	return As(c.v, t)
}

func TestAs(t *testing.T) {
	assert.Panics(t, func() {
		As(fn{}, nil)
	})
	assert.Panics(t, func() {
		var i int
		As(fn{}, &i)
	})
	assert.Panics(t, func() {
		var n *fn = nil
		As(fn{}, n)
	})

	var fnVal interface{ Fn() }
	a := interface{}(fn{})
	assert.True(t, As(a, &fnVal))
	assert.True(t, As(converter{a}, &fnVal))

	var badVal interface{ Fn() int }
	assert.False(t, As(a, &badVal))
	assert.False(t, As(converter{a}, &badVal))
}

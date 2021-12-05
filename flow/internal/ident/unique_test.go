package ident

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueID(t *testing.T) {
	a := UniqueID()
	b := UniqueID()
	assert.NotEqual(t, a, b)
}

func BenchmarkUniqueID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		UniqueID()
	}
}

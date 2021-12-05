package ident

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	pass := [...]string{
		"a",
		"abc",
		"a.b.c",
		"0123456789_.",
		"abcdefghijklmnopqrstuvwxyz",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	}
	fail := [...]string{
		"",
		"@",
		"(abc)",
	}

	for _, p := range pass {
		assert.True(t, Check(p))
	}
	for _, f := range fail {
		assert.False(t, Check(f))
	}
}

func BenchmarkCheck(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Check("hello")
	}
}

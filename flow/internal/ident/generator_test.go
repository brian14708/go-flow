package ident

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	var g Generator
	assert.True(t, strings.HasPrefix(g.Generate(""), "Unnamed"))
	assert.True(t, strings.HasPrefix(g.Generate("@"), "Unnamed"))
	assert.True(t, strings.HasPrefix(g.Generate("test"), "test"))
	assert.NotEqual(t, g.Generate("a"), g.Generate("a"))
}

func BenchmarkGenerate(b *testing.B) {
	var g Generator
	for i := 0; i < b.N; i++ {
		g.Generate("a")
	}
}

package ident

import (
	"strconv"
)

type Generator struct {
	uniq map[string]int
}

func (g *Generator) Copy() Generator {
	if g.uniq == nil {
		g.uniq = make(map[string]int)
	}
	return *g
}

func (g *Generator) Generate(prefix string) string {
	if prefix == "" || !Check(prefix) {
		prefix = "Unnamed"
	}
	if g.uniq == nil {
		g.uniq = map[string]int{
			prefix: 0,
		}
	}
	g.uniq[prefix]++
	return prefix + "_" + strconv.Itoa(g.uniq[prefix])
}

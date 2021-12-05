package flow

import "reflect"

type GraphBuilder interface {
	BuildGraph() (g *Graph, in, out []string, err error)
}

type partial struct {
	g       *Graph
	in, out []string
}

func (v *partial) BuildGraph() (g *Graph, in, out []string, err error) {
	for _, i := range v.in {
		if _, err := v.g.getPort(i, reflect.RecvDir); err != nil {
			return nil, nil, nil, err
		}
	}
	for _, o := range v.out {
		if _, err := v.g.getPort(o, reflect.SendDir); err != nil {
			return nil, nil, nil, err
		}
	}
	return v.g, v.in, v.out, nil
}

func PartialGraph(g *Graph, in, out []string) GraphBuilder {
	return &partial{g, in, out}
}

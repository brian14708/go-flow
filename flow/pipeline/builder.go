package pipeline

import "github.com/brian14708/go-flow/flow"

var _ flow.GraphBuilder = (*Pipeline)(nil)

func (p *Pipeline) BuildGraph() (g *flow.Graph, in, out []string, err error) {
	p = p.initialize(nil)
	in, out = p.Ports()
	p.IgnoreInput()
	p.IgnoreOutput()
	return p.g, in, out, nil
}

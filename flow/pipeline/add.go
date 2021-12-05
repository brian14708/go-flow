package pipeline

import (
	"fmt"
	"reflect"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/node"
)

func (p *Pipeline) getName(name string, n interface{}) string {
	if name == "" {
		t := reflect.TypeOf(n)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		name = p.nameGen.Generate(t.Name())
	}
	return p.prefix + name
}

func (p *Pipeline) Add(name string, n interface{}, opts ...flow.ConnectOption) *Pipeline {
	name = p.getName(name, n)

	blk, err := p.addBlock(name, n)
	if err != nil {
		panic("fail to make stage `" + name + "': " + err.Error())
	}

	err = p.connectBlock(blk, opts)
	if err != nil {
		panic("fail to connect stage `" + name + "': " + err.Error())
	}

	if p.first == nil {
		p.first = blk
	}
	p.last = blk
	return p
}

func (p *Pipeline) Discard(s ...string) *Pipeline {
	p.SplitOutput(s).Add("", node.NewDiscardNode())
	return p
}

// connect block to end of the pipeline
func (p *Pipeline) connectBlock(
	curr *block,
	opts []flow.ConnectOption,
) error {
	addOpts, opts := extractAddOptions(opts)
	inMatcher := newPortMatcher(curr.in)

	// connect all side inputs
	for s, desc := range addOpts.sideInputs {
		var in []string
		for _, m := range desc.ports {
			var ok bool
			in, ok = inMatcher.appendMatch(in, m)
			if !ok {
				return fmt.Errorf("side input `%s' not found", m)
			}
		}
		if s == nil {
			continue
		}
		if s.g != p.g {
			return fmt.Errorf("side input can only be sub-pipeline")
		}
		err := p.g.Connect(s.last.out, in, desc.options...)
		if err != nil {
			return fmt.Errorf("connect failed %s -> %s: %s", s.last.out, in, err.Error())
		}
		s.IgnoreOutput()
	}

	in := inMatcher.remaining()
	var out []string
	if p.last != nil { // not first block
		out = p.last.out

		// check mismatch
		if len(out) != 0 && len(in) == 0 {
			return fmt.Errorf("dangling output ports %s", out)
		} else if len(out) == 0 && len(in) != 0 {
			return fmt.Errorf("dangling input ports %s", in)
		}
	}

	if len(out) == 0 || len(in) == 0 {
		if len(opts) != 0 {
			return fmt.Errorf("dangling connection options")
		}
		return nil
	}

	err := p.g.Connect(out, in, opts...)
	if err != nil {
		return fmt.Errorf("connect failed %s -> %s: %s", out, in, err.Error())
	}
	return nil
}

package pipeline

import (
	"context"
	"fmt"
	"reflect"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/internal/ident"
)

type Pipeline struct {
	g        *flow.Graph
	parent   *Pipeline
	children []*Pipeline
	first    *block
	last     *block

	prefix  string
	nameGen ident.Generator

	initialized bool
}

func New(o *flow.GraphOptions) *Pipeline {
	g, err := flow.NewGraph(o)
	if err != nil {
		panic("new graph failed: " + err.Error())
	}
	return &Pipeline{g: g}
}

func FromGraph(b flow.GraphBuilder) *Pipeline {
	g, in, out, err := b.BuildGraph()
	if err != nil {
		panic(fmt.Sprintf("fail to build graph: %s", err))
	}
	return &Pipeline{
		g:     g,
		first: newBlock(in, nil),
		last:  newBlock(nil, out),
	}
}

func (p *Pipeline) initialize(chk func(*Pipeline)) *Pipeline {
	for p.parent != nil {
		p = p.parent
	}
	if p.initialized {
		panic("pipeline already initialized")
	}
	if p.parent != nil {
		panic("only main pipeline can be initialized")
	}

	var (
		ppl   *Pipeline
		stack = []*Pipeline{p}
	)
	for len(stack) != 0 {
		ppl, stack = stack[len(stack)-1], stack[:len(stack)-1]

		ppl.initialized = true
		if chk != nil {
			chk(ppl)
		}
		stack = append(stack, ppl.children...)
		ppl.parent = nil
		ppl.children = nil
	}

	return p
}

func (p *Pipeline) Run(ctx context.Context) error {
	p = p.initialize(func(p *Pipeline) {
		in, out := p.Ports()
		if len(in) != 0 {
			panic(fmt.Sprintf("dangling pipeline input %s", in))
		}
		if len(out) != 0 {
			panic(fmt.Sprintf("dangling pipeline output %s", out))
		}
	})
	return p.g.Run(ctx)
}

func (p *Pipeline) Graph() *flow.Graph {
	return p.g
}

func (p *Pipeline) IgnoreInput(s ...string) *Pipeline {
	var blk *block
	if len(s) == 0 {
		blk = newBlock(nil, nil)
	} else {
		in, _ := p.Ports()
		matcher := newPortMatcher(in)
		discard := make([]string, 0, 1)
		for _, in := range s {
			var ok bool
			if discard, ok = matcher.appendMatch(discard[:0], in); !ok {
				panic("ignore input `" + in + "' not found")
			}
		}
		blk = newBlock(matcher.remaining(), nil)
	}
	if p.last != nil {
		p.first = blk
	}
	return p
}

func (p *Pipeline) IgnoreOutput(s ...string) *Pipeline {
	if len(s) == 0 {
		p.setOutputPorts(nil)
	} else {
		p.SplitOutput(s)
	}
	return p
}

// pipeline namespace for subpipelines, separated by `.'
func (p *Pipeline) Namespace() string {
	if p.prefix == "" {
		return p.prefix
	}
	return p.prefix[:len(p.prefix)-1]
}

func (p *Pipeline) SubPipeline(name string) *Pipeline {
	type SubPipeline struct{} // tag
	name = p.getName(name, SubPipeline{})
	ppl := &Pipeline{
		g:      p.g,
		parent: p,
		prefix: name + ".",
	}
	p.children = append(p.children, ppl)
	return ppl
}

func (p *Pipeline) SplitOutput(portArg interface{}) *Pipeline {
	_, out := p.Ports()
	outMatcher := newPortMatcher(out)

	var matches []string
	for _, out := range parsePortList(portArg) {
		var ok bool
		matches, ok = outMatcher.appendMatch(matches, out)
		if !ok {
			panic("match output `" + out + "' not found")
		}
	}
	p.setOutputPorts(outMatcher.remaining())

	ppl := &Pipeline{
		g:       p.g,
		parent:  p,
		prefix:  p.prefix,
		nameGen: p.nameGen.Copy(),
	}
	if len(matches) != 0 {
		ppl.first = newBlock(nil, nil)
		ppl.last = newBlock(nil, matches)
	}
	p.children = append(p.children, ppl)
	return ppl
}

func (p *Pipeline) Merge(srcs ...*Pipeline) *Pipeline {
	cnt := 0
	n := 0
	for _, src := range srcs {
		if src.g != p.g {
			panic("fail to merge pipeline: not sub-pipeline")
		}
		if src.first == nil || src == p {
			continue
		}
		srcs[n] = src
		n++
		i, o := src.Ports()
		cnt += len(i) + len(o)
	}
	srcs = srcs[:n]

	inPorts, outPorts := p.Ports()
	tmp := make([]string, 0, cnt+len(inPorts)+len(outPorts))

	tmp = append(tmp, inPorts...)
	for _, src := range srcs {
		srcInPorts, _ := src.Ports()
		tmp = append(tmp, srcInPorts...)
		src.IgnoreInput()
	}
	inCnt := len(tmp)
	tmp = append(tmp, outPorts...)
	for _, src := range srcs {
		_, srcOutPorts := src.Ports()
		tmp = append(tmp, srcOutPorts...)
		src.IgnoreOutput()
	}

	p.first = newBlock(tmp[:inCnt], nil)
	p.last = newBlock(nil, tmp[inCnt:])
	return p
}

func (p *Pipeline) outputType() reflect.Type {
	_, out := p.Ports()
	if len(out) == 0 {
		return nil
	}

	elemType, err := p.g.PortType(out[0])
	if err != nil {
		panic("cannot get output type for `" + out[0] + "': " + err.Error())
	}
	for _, o := range out {
		t, err := p.g.PortType(o)
		if err != nil {
			panic("cannot get output type for `" + o + "': " + err.Error())
		}
		if t != elemType {
			panic("multiple output types for pipeline")
		}
	}
	return elemType
}

func (p *Pipeline) setOutputPorts(s []string) {
	if p.first == nil {
		if len(s) == 0 {
			return
		} else {
			panic("set output on empty pipeline")
		}
	}
	p.last = newBlock(nil, s)
}

func (p *Pipeline) String() string {
	if p.first == nil {
		return "(nil)"
	}

	in, out := p.Ports()
	return fmt.Sprintf("input=%s output=%s", in, out)
}

// read-only slice
func (p *Pipeline) Ports() ([]string, []string) {
	if p.first == nil {
		return nil, nil
	}
	return p.first.in, p.last.out
}

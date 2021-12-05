package pipeline

import (
	"strconv"
	"strings"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/node"
)

func OrderedParallel(n []interface{}, opt ...flow.ConnectOption) func(*Pipeline) {
	if len(n) <= 0 {
		panic("parallel count must be larger than 0")
	}
	return func(p *Pipeline) {
		lb, err := node.NewBalancerNode(len(n))
		if err != nil {
			panic("fail to make stage BalancerNode: " + err.Error())
		}
		p.Add("dispatch", lb)

		name := p.Namespace()
		if idx := strings.LastIndexByte(name, '.'); idx >= 0 {
			name = name[idx+1:]
		}

		joinOpts := make([]flow.ConnectOption, len(n)+1)
		joinOpts[0] = flow.WithChanSize(len(n))

		ppls := make([]*Pipeline, len(n))
		for i := range ppls {
			iStr := strconv.Itoa(i)
			ppl := p.SplitOutput("out_"+iStr).
				Add(name+"_"+iStr, n[i], flow.WithChanSize(0))

			ppls[i] = ppl
			joinOpts[i+1] = SideInput(ppls[i], "in_"+iStr, flow.WithChanSize(0))
		}

		join, err := node.NewSelectNode(len(n))
		if err != nil {
			panic("fail to make SelectNode: " + err.Error())
		}
		p.Add("join", join, joinOpts...)
	}
}

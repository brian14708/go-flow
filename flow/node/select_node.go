package node

import (
	"context"
	"errors"
	"fmt"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/port"
	"github.com/brian14708/go-flow/flowutil"
)

type selectNode struct {
	idx  <-chan int
	srcs []<-chan flow.AnyMessage
	out  chan<- flow.AnyMessage
}

func (s *selectNode) Ports() (in, out flow.PortMap) {
	in = port.MakeMap("in_idx", &s.idx)
	for i := range s.srcs {
		in[fmt.Sprintf("in_%d", i)] = &s.srcs[i]
	}

	return in, port.MakeMap("out", &s.out)
}

func (s *selectNode) Run(ctx context.Context) error {
	chk := flowutil.NewContextChecker(ctx)
	for i := range s.idx {
		val, ok := <-s.srcs[i]
		if !ok {
			panic("SelectNode on closed channel")
		}
		s.out <- val

		if !chk.Valid() {
			return chk.Err()
		}
	}
	return nil
}

func NewSelectNode(n int) (flow.Node, error) {
	if n <= 0 {
		return nil, errors.New("select node must have input channel")
	}

	s := &selectNode{
		srcs: make([]<-chan flow.AnyMessage, n),
	}
	return s, nil
}

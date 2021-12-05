package node

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/port"
	"github.com/brian14708/go-flow/flowutil"
)

type balancerNode struct {
	ch    <-chan flow.AnyMessage
	idx   chan<- int
	dests []chan<- flow.AnyMessage
}

func (l *balancerNode) Ports() (in, out flow.PortMap) {
	out = port.MakeMap("out_idx", &l.idx)
	for i := range l.dests {
		out[fmt.Sprintf("out_%d", i)] = &l.dests[i]
	}
	return port.MakeMap("in", &l.ch), out
}

func (l *balancerNode) Run(ctx context.Context) error {
	ll := newLeastLoad(l.dests)

	chk := flowutil.NewContextChecker(ctx)
	sel := make([]reflect.SelectCase, 0, len(l.dests))
	for _, t := range l.dests {
		sel = append(sel, reflect.SelectCase{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(t),
		})
	}

	for {
		val, ok := <-l.ch
		if !ok {
			break
		}

		idx := -1
		if ll != nil {
			idx = ll.next()
			select {
			case l.dests[idx] <- val:
			default:
				idx = -1
			}
		}
		if idx == -1 {
			v := reflect.ValueOf(val)
			for i := range sel {
				sel[i].Send = v
			}
			idx, _, _ = reflect.Select(sel)
		}
		if l.idx != nil {
			l.idx <- idx
		}

		if !chk.Valid() {
			return chk.Err()
		}
	}
	return nil
}

func NewBalancerNode(n int) (flow.Node, error) {
	if n <= 0 {
		return nil, errors.New("load balancer must have output channel")
	}

	l := &balancerNode{
		dests: make([]chan<- flow.AnyMessage, n),
	}
	return l, nil
}

type leastLoad struct {
	chs      []chan<- flow.AnyMessage
	capacity []float32
}

func newLeastLoad(chs []chan<- flow.AnyMessage) *leastLoad {
	allZero := true
	for _, ch := range chs {
		if cap(ch) != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return nil
	}

	ll := &leastLoad{
		chs: chs,
	}
	for _, ch := range chs {
		c := 1e-2 + float32(cap(ch))
		ll.capacity = append(ll.capacity, c)
	}
	return ll
}

func (ll *leastLoad) next() int {
	var (
		minIdx         = 0
		min    float32 = 1.0
	)
	for i, ch := range ll.chs {
		util := float32(len(ch)) / ll.capacity[i]
		if util < min {
			minIdx, min = i, util
		}
	}
	return minIdx
}

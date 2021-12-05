package node

import (
	"context"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/port"
)

type discardNode struct {
	ch <-chan flow.AnyMessage
}

func (d *discardNode) Ports() (in, out flow.PortMap) {
	in = port.MakeMap()
	in["in"] = &d.ch
	return
}

func (d *discardNode) Run(ctx context.Context) error {
	var ch flow.Chan
	if _, fctx := flow.FromContext(ctx); fctx != nil {
		ch = fctx.GetChan("in")
	}

	for val := range d.ch {
		if ch != nil {
			ch.DropMessage(val)
		}
	}
	return nil
}

func NewDiscardNode() flow.Node {
	return new(discardNode)
}

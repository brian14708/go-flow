package flow

import (
	"context"

	"github.com/brian14708/go-flow/flowtype"
	_ "github.com/brian14708/go-flow/flowtype/builtin"
)

type Node interface {
	Run(context.Context) error
	Ports() (in, out PortMap)
}

// should implement this interface if a wrapper node is needed and want:
//   * graph runtime to treat it as if it were the original node
//   * forward all optional attributes
type NodeWrapper interface {
	Node
	// expose underlying node for deduplication
	NodeHash() interface{}
	// get access to underlying node attributes using flowtype.As
	As(interface{}) bool
}

func WrapNode(n Node) NodeWrapper {
	return nodeWrapper{n}
}

type nodeWrapper struct{ Node }

func (n nodeWrapper) NodeHash() interface{} {
	var hasher interface{ NodeHash() interface{} }
	if ok := flowtype.As(n.Node, &hasher); ok {
		return hasher.NodeHash()
	}
	return n.Node
}

func (n nodeWrapper) As(t interface{}) bool {
	return flowtype.As(n.Node, t)
}

package flow

import (
	"reflect"
	"sort"

	"github.com/brian14708/go-flow/flowdebug/types"
	"github.com/brian14708/go-flow/flowtype"
)

func (g *Graph) Topology() *types.Topology {
	topology := &types.Topology{
		ID: g.ID(),
	}

	for name, node := range g.nodes {
		nodeDesc := types.Node{
			Name: name,
		}

		// node info
		{
			var t interface{ NodeType() string }
			if ok := flowtype.As(node.node, &t); ok {
				nodeDesc.TypeName = t.NodeType()
			} else {
				nodeDesc.TypeName = reflect.TypeOf(node.nodeHash).String()
			}
		}
		{
			var t interface{ Description() string }
			if ok := flowtype.As(node.node, &t); ok {
				nodeDesc.Description = t.Description()
			}
		}
		// input port
		for _, val := range node.in {
			nodeDesc.InPorts = append(nodeDesc.InPorts, types.Port{
				Name:     val.name,
				TypeName: val.ElemType.String(),
			})
		}
		sort.Slice(nodeDesc.InPorts, func(a, b int) bool {
			return nodeDesc.InPorts[a].Name < nodeDesc.InPorts[b].Name
		})
		// output port
		for _, val := range node.out {
			nodeDesc.OutPorts = append(nodeDesc.OutPorts, types.Port{
				Name:     val.name,
				TypeName: val.ElemType.String(),
			})
		}
		sort.Slice(nodeDesc.OutPorts, func(a, b int) bool {
			return nodeDesc.OutPorts[a].Name < nodeDesc.OutPorts[b].Name
		})

		topology.Nodes = append(topology.Nodes, nodeDesc)
	}
	sort.Slice(topology.Nodes, func(a, b int) bool {
		return topology.Nodes[a].Name < topology.Nodes[b].Name
	})

	for _, conn := range g.conns {
		connDesc := types.Connection{
			ID:          conn.id,
			Source:      conn.src,
			Destination: conn.dst,
			Capacity:    conn.ch.Cap(),
		}
		topology.Connections = append(topology.Connections, connDesc)
	}
	sort.Slice(topology.Connections, func(a, b int) bool {
		return topology.Connections[a].ID < topology.Connections[b].ID
	})

	return topology
}

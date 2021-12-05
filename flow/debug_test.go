package flow

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MultiIONode struct {
	InA  <-chan int
	InB  <-chan int
	OutA chan<- int
	OutB chan<- int
}

func (m *MultiIONode) Ports() (PortMap, PortMap) {
	return PortMap{
			"in_a": &m.InA,
			"in_b": &m.InB,
		}, PortMap{
			"out_a": &m.OutA,
			"out_b": &m.OutB,
		}
}

func (m *MultiIONode) Run(ctx context.Context) error {
	return nil
}

func (m *MultiIONode) NodeType() string {
	return reflect.TypeOf(int(0)).String()
}

func (m *MultiIONode) Description() string {
	return "DESC"
}

func TestGraphTopology(t *testing.T) {
	g, _ := NewGraph(nil)
	buildGraph(g)
	assert.NoError(t, g.AddNode("a", new(MultiIONode)))

	topo := g.Topology()
	assert.Equal(t, 2, len(topo.Connections))
	assert.Equal(t, 6, len(topo.Nodes))
	assert.Equal(t, "a", topo.Nodes[0].Name) // sorted
	assert.Equal(t, 2, len(topo.Nodes[0].InPorts))
}

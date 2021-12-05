package flowdebug

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flowdebug/types"
)

func TestGraphviz(t *testing.T) {
	_, err := Graphviz(nil)
	assert.Error(t, err)
	_, _ = Graphviz(&types.Topology{
		Nodes: []types.Node{
			{
				Name: "na",
				OutPorts: []types.Port{
					{
						Name: "out",
					},
				},
			},
			{
				Name: "nb",
				InPorts: []types.Port{
					{
						Name: "in",
					},
				},
			},
		},
		Connections: []types.Connection{
			{
				ID:          "e",
				Source:      []string{"na:out"},
				Destination: []string{"nb:in"},
			},
		},
	})
}

package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPartialGraph(t *testing.T) {
	g, err := NewGraph(nil)
	assert.NoError(t, err)
	buildGraph(g)

	_, _, _, err = PartialGraph(g, nil, nil).BuildGraph()
	assert.NoError(t, err)

	_, _, _, err = PartialGraph(g, []string{"a:x"}, nil).BuildGraph()
	assert.Error(t, err)

	_, _, _, err = PartialGraph(g, nil, []string{"a:x"}).BuildGraph()
	assert.Error(t, err)
}

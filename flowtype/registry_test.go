package flowtype

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	tmp := NewRegistry()

	type T struct{}
	rtype := reflect.TypeOf(T{})
	assert.Error(t, tmp.RegisterDispatch(rtype, new(DispatchTable)))

	require.Equal(t, GenericDispatch, tmp.GetDispatchTable(rtype))

	dtable := &DispatchTable{
		Version: DispatchVersion,
	}
	_ = tmp.RegisterDispatch(rtype, dtable)
	assert.Equal(t, dtable, tmp.GetDispatchTable(rtype))
}

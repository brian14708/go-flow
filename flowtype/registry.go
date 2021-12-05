package flowtype

import (
	"fmt"
	"reflect"
	"sync"
)

var DefaultRegistry = NewRegistry()

type Registry struct {
	dispatch sync.Map
}

func NewRegistry() *Registry {
	return new(Registry)
}

func (r *Registry) RegisterDispatch(t reflect.Type, i *DispatchTable) error {
	if i.Version != DispatchVersion {
		return fmt.Errorf("invalid typeinfo version `%d', expected `%d'", i.Version, DispatchVersion)
	}

	r.dispatch.Store(t, i)
	return nil
}

func (r *Registry) GetDispatchTable(t reflect.Type) *DispatchTable {
	if d, ok := r.dispatch.Load(t); ok {
		return d.(*DispatchTable)
	}
	return GenericDispatch
}

func RegisterDispatch(t reflect.Type, i *DispatchTable) error {
	return DefaultRegistry.RegisterDispatch(t, i)
}

func MustRegisterDispatch(t reflect.Type, i *DispatchTable) {
	if err := RegisterDispatch(t, i); err != nil {
		panic(fmt.Sprintf("type register failed `%s': %v", t.String(), err))
	}
}

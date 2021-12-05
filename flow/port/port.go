package port

import (
	"reflect"
	"sync"
)

var portPool = sync.Pool{
	New: func() interface{} {
		return new(Port)
	},
}

type Port struct {
	// types that may differ between compile time and runtime
	ElemType        reflect.Type
	StorageElemType reflect.Type

	Ref interface{}
}

func TemplatePort(ch interface{}, ty reflect.Type) *Port {
	p := portPool.Get().(*Port)
	*p = Port{
		Ref:      ch,
		ElemType: ty,
	}
	return p
}

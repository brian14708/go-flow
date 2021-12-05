package channel

import (
	"reflect"
)

type runtimeChan struct {
	reflect.Value
	*options
}

func (r *runtimeChan) DropMessage(v interface{}) {
	if r.dropHandler != nil {
		r.dropHandler(v)
	}
}

func (r *runtimeChan) AssignTo(_ reflect.ChanDir, p interface{}) {
	reflect.ValueOf(p).Elem().Set(r.Value)
}

func (r *runtimeChan) Serve() {}

func (r *runtimeChan) NeedServe() bool {
	return false
}

func (r *runtimeChan) Drain() {
	drain(r.Value, r.drainRate, r.dropHandler)
}

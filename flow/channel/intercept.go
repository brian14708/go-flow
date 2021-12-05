package channel

import (
	"reflect"

	"github.com/brian14708/go-flow/flowtype"
)

type interceptChan struct {
	reflect.Value
	output reflect.Value
	*options
}

func (o *interceptChan) Drain() {
	drain(o.output, o.drainRate, o.dropHandler)
}

func (o *interceptChan) Serve() {
	// move elements from o.Value -> o.output

	recv := flowtype.ChanRecver(o.Value.Interface())
	send := flowtype.ChanSender(o.output.Interface())

	for {
		v, ok := recv(nil, true)
		if !ok {
			o.output.Close()
			return
		}

		for _, interceptor := range o.interceptor {
			v = interceptor(v)
		}

		send(v, nil, true)
	}
}

func (o *interceptChan) DropMessage(v interface{}) {
	if o.dropHandler != nil {
		o.dropHandler(v)
	}
}

func (o *interceptChan) AssignTo(d reflect.ChanDir, p interface{}) {
	el := reflect.ValueOf(p).Elem()
	switch d {
	case reflect.RecvDir:
		el.Set(o.output)
	case reflect.SendDir:
		el.Set(o.Value)
	default:
		panic("invalid chan direction")
	}
}

package channel

import (
	"reflect"
)

type Channel interface {
	Cap() int
	Len() int
	DropMessage(interface{})

	// get underlying channel
	AssignTo(reflect.ChanDir, interface{})
	Serve()

	Close()
	Drain()
}

func New(src, dst reflect.Type, os *Storage, opts ...Option) (Channel, error) {
	var opt *options

	if os != nil {
		opt = &os.options
		*opt = defaultOptions
	} else {
		tmp := defaultOptions
		opt = &tmp
	}

	for _, o := range opts {
		o.apply(opt)
	}

	if opt.adaptiveGain > 0 && opt.size > 4 {
		return newAdaptiveChan(src, dst, opt)
	}

	if len(opt.interceptor) == 0 && src == dst {
		ch := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, src), opt.size)
		return &runtimeChan{ch, opt}, nil
	}

	return &interceptChan{
		Value:   reflect.MakeChan(reflect.ChanOf(reflect.BothDir, src), opt.size),
		output:  reflect.MakeChan(reflect.ChanOf(reflect.BothDir, dst), 0),
		options: opt,
	}, nil
}

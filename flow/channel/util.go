package channel

import (
	"context"
	"reflect"

	"golang.org/x/time/rate"

	"github.com/brian14708/go-flow/flowtype"
)

func drain(ch reflect.Value, r rate.Limit, dropper func(interface{})) {
	val, ok := ch.Recv()
	if !ok {
		return
	}
	if dropper != nil {
		dropper(val.Interface())
	}

	recv := flowtype.ChanRecver(ch.Interface())
	l := ch.Cap()
	// non blocking drain all
	for i := 0; i < l; i++ {
		val, ok := recv(nil, false)
		if !ok {
			break
		}
		if dropper != nil {
			dropper(val)
		}
	}

	var limit *rate.Limiter
	for {
		val, ok := recv(nil, true)
		if !ok {
			return
		}
		if dropper != nil {
			dropper(val)
		}

		if limit == nil {
			limit = rate.NewLimiter(r, 0)
		}
		_ = limit.Wait(context.Background())
	}
}

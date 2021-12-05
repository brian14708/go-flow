package flowtype

import (
	"context"
	"reflect"
)

const DispatchVersion = 4

type (
	SendFunc func(val interface{}, cancel <-chan struct{}, block bool) bool
	RecvFunc func(cancel <-chan struct{}, block bool) (interface{}, bool)
	CallFunc func(arg, buf []interface{}) []interface{}
)

type DispatchTable struct {
	Version    int
	ChanSender func(ch interface{}) SendFunc
	ChanRecver func(ch interface{}) RecvFunc
	FuncCaller func(fn interface{}) CallFunc
}

var GenericDispatch = &DispatchTable{
	Version: DispatchVersion,
	ChanRecver: func(c interface{}) RecvFunc {
		ch := reflect.ValueOf(c)
		if (ch.Type().ChanDir() & reflect.RecvDir) == 0 {
			panic("invalid chan direction")
		}
		return func(cancel <-chan struct{}, block bool) (interface{}, bool) {
			if !block {
				if val, ok := ch.TryRecv(); ok {
					return val.Interface(), true
				}
				return nil, false
			}
			if cancel == nil {
				if val, ok := ch.Recv(); ok {
					return val.Interface(), true
				}
				return nil, false
			}

			idx, val, ok := reflect.Select([]reflect.SelectCase{
				{
					Dir:  reflect.SelectRecv,
					Chan: ch,
				},
				{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(cancel),
				},
			})
			if idx == 0 && ok {
				return val.Interface(), true
			}
			return nil, false
		}
	},
	ChanSender: func(c interface{}) SendFunc {
		ch := reflect.ValueOf(c)
		if (ch.Type().ChanDir() & reflect.SendDir) == 0 {
			panic("invalid chan direction")
		}
		nilElem := reflect.Zero(ch.Type().Elem())
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			el := nilElem
			if v != nil {
				el = reflect.ValueOf(v)
			}
			if !block {
				return ch.TrySend(el)
			}
			if cancel == nil {
				ch.Send(el)
				return true
			}

			idx, _, _ := reflect.Select([]reflect.SelectCase{
				{
					Dir:  reflect.SelectSend,
					Chan: ch,
					Send: el,
				},
				{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(cancel),
				},
			})
			return idx == 0
		}
	},
	FuncCaller: func(f interface{}) CallFunc {
		fn := reflect.ValueOf(f)

		t := fn.Type()
		if t.Kind() != reflect.Func {
			panic("invalid function")
		}

		nilArgs := make([]reflect.Value, t.NumIn())
		for i := range nilArgs {
			nilArgs[i] = reflect.Zero(t.In(i))
		}

		return func(in, out []interface{}) []interface{} {
			var buf [4]reflect.Value
			args := buf[:0]
			for i, v := range in {
				if v == nil {
					args = append(args, nilArgs[i])
				} else {
					args = append(args, reflect.ValueOf(v))
				}
			}

			ret := fn.Call(args)

			if len(ret) == 0 {
				return nil
			}
			if cap(out)-len(out) < len(ret) {
				out = make([]interface{}, 0, len(ret))
			}
			for _, r := range ret {
				out = append(out, r.Interface())
			}
			return out
		}
	},
}

func ChanSender(ch interface{}) SendFunc {
	typ := reflect.TypeOf(ch)
	if typ.Kind() == reflect.Chan {
		typ = typ.Elem()
	}
	tbl := DefaultRegistry.GetDispatchTable(typ)
	return tbl.ChanSender(ch)
}

func ChanRecver(ch interface{}) RecvFunc {
	typ := reflect.TypeOf(ch)
	if typ.Kind() == reflect.Chan {
		typ = typ.Elem()
	}
	tbl := DefaultRegistry.GetDispatchTable(typ)
	return tbl.ChanRecver(ch)
}

var commonTypes = map[reflect.Type]struct{}{
	reflect.TypeOf((*context.Context)(nil)).Elem(): {},
	reflect.TypeOf((*error)(nil)).Elem():           {},
}

func FuncCaller(fn interface{}) CallFunc {
	typ := reflect.TypeOf(fn)
	if typ.Kind() == reflect.Func {
		var foundType reflect.Type
		for i := 0; i < typ.NumIn() && foundType == nil; i++ {
			foundType = typ.In(i)
			if _, ok := commonTypes[foundType]; ok {
				foundType = nil
			}
		}
		for i := 0; i < typ.NumOut() && foundType == nil; i++ {
			foundType = typ.Out(i)
			if _, ok := commonTypes[foundType]; ok {
				foundType = nil
			}
		}
		if foundType != nil {
			typ = foundType
		}
	}
	tbl := DefaultRegistry.GetDispatchTable(typ)
	return tbl.FuncCaller(fn)
}

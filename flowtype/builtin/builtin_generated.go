//go:build !skipflowtype
// +build !skipflowtype

// Code generated by github.com/brian14708/go-flow/flowtype/codegen. DO NOT EDIT.
package builtin

import (
	flowtype "github.com/brian14708/go-flow/flowtype"
	"reflect"
)

var builtinBool = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan bool
		switch c := c.(type) {
		case <-chan bool:
			ch = c
		case chan bool:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- bool
		switch c := c.(type) {
		case chan<- bool:
			ch = c
		case chan bool:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el bool
			if v != nil {
				el = v.(bool)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(bool) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 bool
				if i := args[0]; i != nil {
					i0 = i.(bool)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinComplex64 = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan complex64
		switch c := c.(type) {
		case <-chan complex64:
			ch = c
		case chan complex64:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- complex64
		switch c := c.(type) {
		case chan<- complex64:
			ch = c
		case chan complex64:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el complex64
			if v != nil {
				el = v.(complex64)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(complex64) complex64:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 complex64
				if i := args[0]; i != nil {
					i0 = i.(complex64)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(complex64) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 complex64
				if i := args[0]; i != nil {
					i0 = i.(complex64)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinComplex128 = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan complex128
		switch c := c.(type) {
		case <-chan complex128:
			ch = c
		case chan complex128:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- complex128
		switch c := c.(type) {
		case chan<- complex128:
			ch = c
		case chan complex128:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el complex128
			if v != nil {
				el = v.(complex128)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(complex128) complex128:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 complex128
				if i := args[0]; i != nil {
					i0 = i.(complex128)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(complex128) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 complex128
				if i := args[0]; i != nil {
					i0 = i.(complex128)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinError = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan error
		switch c := c.(type) {
		case <-chan error:
			ch = c
		case chan error:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- error
		switch c := c.(type) {
		case chan<- error:
			ch = c
		case chan error:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el error
			if v != nil {
				el = v.(error)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(error) error:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 error
				if i := args[0]; i != nil {
					i0 = i.(error)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(error) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 error
				if i := args[0]; i != nil {
					i0 = i.(error)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinFloat32 = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan float32
		switch c := c.(type) {
		case <-chan float32:
			ch = c
		case chan float32:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- float32
		switch c := c.(type) {
		case chan<- float32:
			ch = c
		case chan float32:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el float32
			if v != nil {
				el = v.(float32)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(float32) float32:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 float32
				if i := args[0]; i != nil {
					i0 = i.(float32)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(float32) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 float32
				if i := args[0]; i != nil {
					i0 = i.(float32)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinFloat64 = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan float64
		switch c := c.(type) {
		case <-chan float64:
			ch = c
		case chan float64:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- float64
		switch c := c.(type) {
		case chan<- float64:
			ch = c
		case chan float64:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el float64
			if v != nil {
				el = v.(float64)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(float64) float64:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 float64
				if i := args[0]; i != nil {
					i0 = i.(float64)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(float64) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 float64
				if i := args[0]; i != nil {
					i0 = i.(float64)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinInt = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan int
		switch c := c.(type) {
		case <-chan int:
			ch = c
		case chan int:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- int
		switch c := c.(type) {
		case chan<- int:
			ch = c
		case chan int:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el int
			if v != nil {
				el = v.(int)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(int) int:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 int
				if i := args[0]; i != nil {
					i0 = i.(int)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(int) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 int
				if i := args[0]; i != nil {
					i0 = i.(int)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinInt8 = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan int8
		switch c := c.(type) {
		case <-chan int8:
			ch = c
		case chan int8:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- int8
		switch c := c.(type) {
		case chan<- int8:
			ch = c
		case chan int8:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el int8
			if v != nil {
				el = v.(int8)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(int8) int8:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 int8
				if i := args[0]; i != nil {
					i0 = i.(int8)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(int8) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 int8
				if i := args[0]; i != nil {
					i0 = i.(int8)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinInt16 = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan int16
		switch c := c.(type) {
		case <-chan int16:
			ch = c
		case chan int16:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- int16
		switch c := c.(type) {
		case chan<- int16:
			ch = c
		case chan int16:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el int16
			if v != nil {
				el = v.(int16)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(int16) int16:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 int16
				if i := args[0]; i != nil {
					i0 = i.(int16)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(int16) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 int16
				if i := args[0]; i != nil {
					i0 = i.(int16)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinInt32 = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan int32
		switch c := c.(type) {
		case <-chan int32:
			ch = c
		case chan int32:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- int32
		switch c := c.(type) {
		case chan<- int32:
			ch = c
		case chan int32:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el int32
			if v != nil {
				el = v.(int32)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(int32) int32:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 int32
				if i := args[0]; i != nil {
					i0 = i.(int32)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(int32) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 int32
				if i := args[0]; i != nil {
					i0 = i.(int32)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinInt64 = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan int64
		switch c := c.(type) {
		case <-chan int64:
			ch = c
		case chan int64:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- int64
		switch c := c.(type) {
		case chan<- int64:
			ch = c
		case chan int64:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el int64
			if v != nil {
				el = v.(int64)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(int64) int64:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 int64
				if i := args[0]; i != nil {
					i0 = i.(int64)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(int64) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 int64
				if i := args[0]; i != nil {
					i0 = i.(int64)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinString = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan string
		switch c := c.(type) {
		case <-chan string:
			ch = c
		case chan string:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- string
		switch c := c.(type) {
		case chan<- string:
			ch = c
		case chan string:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el string
			if v != nil {
				el = v.(string)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(string) string:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 string
				if i := args[0]; i != nil {
					i0 = i.(string)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(string) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 string
				if i := args[0]; i != nil {
					i0 = i.(string)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinUint = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan uint
		switch c := c.(type) {
		case <-chan uint:
			ch = c
		case chan uint:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- uint
		switch c := c.(type) {
		case chan<- uint:
			ch = c
		case chan uint:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el uint
			if v != nil {
				el = v.(uint)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(uint) uint:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 uint
				if i := args[0]; i != nil {
					i0 = i.(uint)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(uint) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 uint
				if i := args[0]; i != nil {
					i0 = i.(uint)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinUint8 = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan uint8
		switch c := c.(type) {
		case <-chan uint8:
			ch = c
		case chan uint8:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- uint8
		switch c := c.(type) {
		case chan<- uint8:
			ch = c
		case chan uint8:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el uint8
			if v != nil {
				el = v.(uint8)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(uint8) uint8:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 uint8
				if i := args[0]; i != nil {
					i0 = i.(uint8)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(uint8) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 uint8
				if i := args[0]; i != nil {
					i0 = i.(uint8)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinUint16 = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan uint16
		switch c := c.(type) {
		case <-chan uint16:
			ch = c
		case chan uint16:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- uint16
		switch c := c.(type) {
		case chan<- uint16:
			ch = c
		case chan uint16:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el uint16
			if v != nil {
				el = v.(uint16)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(uint16) uint16:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 uint16
				if i := args[0]; i != nil {
					i0 = i.(uint16)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(uint16) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 uint16
				if i := args[0]; i != nil {
					i0 = i.(uint16)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinUint32 = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan uint32
		switch c := c.(type) {
		case <-chan uint32:
			ch = c
		case chan uint32:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- uint32
		switch c := c.(type) {
		case chan<- uint32:
			ch = c
		case chan uint32:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el uint32
			if v != nil {
				el = v.(uint32)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(uint32) uint32:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 uint32
				if i := args[0]; i != nil {
					i0 = i.(uint32)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(uint32) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 uint32
				if i := args[0]; i != nil {
					i0 = i.(uint32)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinUint64 = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan uint64
		switch c := c.(type) {
		case <-chan uint64:
			ch = c
		case chan uint64:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- uint64
		switch c := c.(type) {
		case chan<- uint64:
			ch = c
		case chan uint64:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el uint64
			if v != nil {
				el = v.(uint64)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(uint64) uint64:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 uint64
				if i := args[0]; i != nil {
					i0 = i.(uint64)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(uint64) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 uint64
				if i := args[0]; i != nil {
					i0 = i.(uint64)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinUintptr = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan uintptr
		switch c := c.(type) {
		case <-chan uintptr:
			ch = c
		case chan uintptr:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- uintptr
		switch c := c.(type) {
		case chan<- uintptr:
			ch = c
		case chan uintptr:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el uintptr
			if v != nil {
				el = v.(uintptr)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(uintptr) uintptr:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 uintptr
				if i := args[0]; i != nil {
					i0 = i.(uintptr)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(uintptr) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 uintptr
				if i := args[0]; i != nil {
					i0 = i.(uintptr)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var builtinInterface = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan interface{}
		switch c := c.(type) {
		case <-chan interface{}:
			ch = c
		case chan interface{}:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- interface{}
		switch c := c.(type) {
		case chan<- interface{}:
			ch = c
		case chan interface{}:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el interface{}
			if v != nil {
				el = v.(interface{})
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(interface{}) interface{}:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 interface{}
				if i := args[0]; i != nil {
					i0 = i.(interface{})
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(interface{}) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 interface{}
				if i := args[0]; i != nil {
					i0 = i.(interface{})
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

var dispatchAnyMessage = &flowtype.DispatchTable{
	ChanRecver: func(c interface{}) flowtype.RecvFunc {
		var ch <-chan flowtype.AnyMessage
		switch c := c.(type) {
		case <-chan flowtype.AnyMessage:
			ch = c
		case chan flowtype.AnyMessage:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanRecver(c)
		}
		return func(cancel <-chan struct{}, block bool) (v interface{}, ok bool) {
			if !block {
				select {
				case v, ok = <-ch:
				default:
				}
			} else if cancel == nil {
				v, ok = <-ch
			} else {
				select {
				case v, ok = <-ch:
				case <-cancel:
				}
			}
			return
		}
	},
	ChanSender: func(c interface{}) flowtype.SendFunc {
		var ch chan<- flowtype.AnyMessage
		switch c := c.(type) {
		case chan<- flowtype.AnyMessage:
			ch = c
		case chan flowtype.AnyMessage:
			ch = c
		default:
			return flowtype.GenericDispatch.ChanSender(c)
		}
		return func(v interface{}, cancel <-chan struct{}, block bool) bool {
			var el flowtype.AnyMessage
			if v != nil {
				el = v.(flowtype.AnyMessage)
			}
			if !block {
				select {
				case ch <- el:
				default:
					return false
				}
			} else if cancel == nil {
				ch <- el
			} else {
				select {
				case ch <- el:
				case <-cancel:
					return false
				}
			}
			return true
		}
	},
	FuncCaller: func(fn interface{}) flowtype.CallFunc {
		switch fn := fn.(type) {
		case func(flowtype.AnyMessage) flowtype.AnyMessage:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 flowtype.AnyMessage
				if i := args[0]; i != nil {
					i0 = i.(flowtype.AnyMessage)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		case func(flowtype.AnyMessage) bool:
			return func(args, ret []interface{}) []interface{} {
				if 1 != len(args) {
					panic("wrong number of arguments")
				}
				var i0 flowtype.AnyMessage
				if i := args[0]; i != nil {
					i0 = i.(flowtype.AnyMessage)
				}
				o0 := fn(i0)
				return append(ret, o0)
			}
		}
		return flowtype.GenericDispatch.FuncCaller(fn)
	},
	Version: 4,
}

func init() {
	flowtype.MustRegisterDispatch(reflect.TypeOf((*bool)(nil)).Elem(), builtinBool)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*complex64)(nil)).Elem(), builtinComplex64)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*complex128)(nil)).Elem(), builtinComplex128)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*error)(nil)).Elem(), builtinError)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*float32)(nil)).Elem(), builtinFloat32)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*float64)(nil)).Elem(), builtinFloat64)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*int)(nil)).Elem(), builtinInt)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*int8)(nil)).Elem(), builtinInt8)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*int16)(nil)).Elem(), builtinInt16)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*int32)(nil)).Elem(), builtinInt32)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*int64)(nil)).Elem(), builtinInt64)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*string)(nil)).Elem(), builtinString)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*uint)(nil)).Elem(), builtinUint)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*uint8)(nil)).Elem(), builtinUint8)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*uint16)(nil)).Elem(), builtinUint16)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*uint32)(nil)).Elem(), builtinUint32)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*uint64)(nil)).Elem(), builtinUint64)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*uintptr)(nil)).Elem(), builtinUintptr)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*interface{})(nil)).Elem(), builtinInterface)
	flowtype.MustRegisterDispatch(reflect.TypeOf((*flowtype.AnyMessage)(nil)).Elem(), dispatchAnyMessage)
}

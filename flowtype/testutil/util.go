package testutil

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flowtype"
)

type T struct{}

func isTCompatible(rtype reflect.Type) bool {
	if rtype.Kind() == reflect.Interface && reflect.TypeOf(T{}).Implements(rtype) {
		return true
	}
	return false
}

func RunChanTest(t *testing.T, rtype reflect.Type) {
	tbl := flowtype.DefaultRegistry.GetDispatchTable(rtype)
	assert.NotNil(t, tbl)

	testChanRecver(t, tbl, rtype)
	testChanSender(t, tbl, rtype)
}

func testChanRecver(t *testing.T, tbl *flowtype.DispatchTable, rtype reflect.Type) {
	ch := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, rtype), 1)
	recv := tbl.ChanRecver(ch.Interface())
	assert.NotNil(t, recv)

	val := reflect.Zero(rtype)
	{
		ch.Send(val)
		got, ok := recv(nil, true)
		assert.True(t, ok)
		assert.Equal(t, got, val.Interface())
	}

	{
		ch.Send(val)
		got, ok := recv(nil, false)
		assert.True(t, ok)
		assert.Equal(t, got, val.Interface())
	}

	{
		ch.Send(val)
		done := make(chan struct{})
		got, ok := recv(done, true)
		assert.True(t, ok)
		assert.Equal(t, got, val.Interface())
		close(done)
		_, ok = recv(done, true)
		assert.False(t, ok)
	}

	{
		ch.Send(val)
		done := make(chan struct{})
		got, ok := recv(done, false)
		assert.True(t, ok)
		assert.Equal(t, got, val.Interface())
		_, ok = recv(done, false)
		assert.False(t, ok)
		close(done)
		_, ok = recv(done, true)
		assert.False(t, ok)
	}

	ch.Close()
	assert.Panics(t, func() {
		ch := ch.Convert(reflect.ChanOf(reflect.SendDir, rtype))
		tbl.ChanRecver(ch.Interface())
	})

	ch = ch.Convert(reflect.ChanOf(reflect.RecvDir, rtype))
	recv = tbl.ChanRecver(ch.Interface())
	assert.NotNil(t, recv)

	_, ok := recv(nil, true)
	assert.False(t, ok)
}

func testChanSender(t *testing.T, tbl *flowtype.DispatchTable, rtype reflect.Type) {
	ch := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, rtype), 1)
	send := tbl.ChanSender(ch.Interface())
	assert.NotNil(t, send)

	val := reflect.Zero(rtype)
	{
		send(val.Interface(), nil, true)
		_, ok := ch.Recv()
		assert.True(t, ok)
	}

	{
		assert.True(t, send(nil, nil, true))
		assert.False(t, send(nil, nil, false))
		_, ok := ch.Recv()
		assert.True(t, ok)
	}

	{
		assert.True(t, send(nil, nil, false))
		assert.False(t, send(nil, nil, false))
		_, ok := ch.Recv()
		assert.True(t, ok)
	}

	{
		done := make(chan struct{})
		assert.True(t, send(nil, done, true))
		assert.False(t, send(nil, done, false))
		_, ok := ch.Recv()
		assert.True(t, ok)
	}

	{
		done := make(chan struct{})
		assert.True(t, send(nil, done, false))
		assert.False(t, send(nil, done, false))
		go close(done)
		assert.False(t, send(nil, done, true))
		_, ok := ch.Recv()
		assert.True(t, ok)
	}

	if isTCompatible(rtype) {
		send(T{}, nil, true)
		ch.Recv()
	} else {
		assert.Panics(t, func() {
			send(T{}, nil, true)
		})
	}

	ch.Close()

	assert.Panics(t, func() {
		ch := ch.Convert(reflect.ChanOf(reflect.RecvDir, rtype))
		tbl.ChanSender(ch.Interface())
	})

	ch = ch.Convert(reflect.ChanOf(reflect.SendDir, rtype))
	send = tbl.ChanSender(ch.Interface())
	assert.NotNil(t, send)

	assert.Panics(t, func() {
		// send on close channel
		send(val.Interface(), nil, true)
	})
}

func RunFuncCallerTest(t *testing.T, rtype reflect.Type, cnt *int, fn interface{}, args interface{}) {
	tbl := flowtype.DefaultRegistry.GetDispatchTable(rtype)
	assert.NotNil(t, tbl)

	*cnt = 0
	for i := 1; i < 5; i++ {
		tt := reflect.TypeOf(fn)
		cArgs := make([]interface{}, tt.NumIn())
		for j := 0; j < tt.NumIn(); j++ {
			if tt.In(j) == rtype {
				cArgs[j] = args
			}
		}
		tbl.FuncCaller(fn)(cArgs, nil)
		assert.Equal(t, i, *cnt)
	}

	assert.Nil(t, tbl.FuncCaller(func() error { return nil })(nil, nil)[0])

	{
		ft := reflect.TypeOf(fn)
		panics := false
		var args []interface{}
		for i := 0; i < ft.NumIn(); i++ {
			switch t := ft.In(i); t {
			case reflect.TypeOf((*context.Context)(nil)).Elem():
				args = append(args, context.Background())
			default:
				if t.Kind() == reflect.Interface {
					args = append(args, T{})
					if !isTCompatible(t) {
						panics = true
					}
				} else {
					args = append(args, reflect.Zero(t).Interface())
				}
			}
		}
		if panics {
			assert.Panics(t, func() {
				tbl.FuncCaller(fn)(args, nil)
			})
		} else {
			tbl.FuncCaller(fn)(args, nil)
		}
	}

	assert.Panics(t, func() {
		tbl.FuncCaller(fn)([]interface{}{T{}, T{}, T{}, T{}}, nil)
	})
}

func TestWithNilRegistry(t *testing.T, f func(t *testing.T)) {
	t.Run("DefaultRegistry", f)

	tmp := flowtype.NewRegistry()
	old := flowtype.DefaultRegistry
	flowtype.DefaultRegistry = tmp
	defer func() {
		flowtype.DefaultRegistry = old
	}()
	t.Run("NilRegistry", f)
}

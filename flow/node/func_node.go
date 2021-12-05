package node

import (
	"context"
	"fmt"
	"reflect"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/channel"
	"github.com/brian14708/go-flow/flow/internal/funcx"
	"github.com/brian14708/go-flow/flow/port"
)

type funcNode struct {
	fn interface{}

	hasCtx bool
	in     interface{}
	out    interface{}
}

func (f *funcNode) Run(ctx context.Context) error {
	var args [3]reflect.Value
	idx := 0
	if f.hasCtx {
		args[idx] = reflect.ValueOf(ctx)
		idx++
	}
	if f.in != nil {
		args[idx] = reflect.ValueOf(f.in).Elem()
		idx++
	}
	if f.out != nil {
		args[idx] = reflect.ValueOf(f.out).Elem()
		idx++
	}

	ret := reflect.ValueOf(f.fn).Call(args[:idx])
	if len(ret) == 0 {
		return nil
	}
	if err := ret[0].Interface(); err != nil {
		return err.(error)
	}
	return nil
}

func (f *funcNode) Ports() (in, out flow.PortMap) {
	if f.in != nil {
		in = port.MakeMap()
		in["in"] = f.in
	}
	if f.out != nil {
		out = port.MakeMap()
		out["out"] = f.out
	}
	return
}

func (f *funcNode) NodeType() string {
	return reflect.TypeOf(f.fn).String()
}

var fnTemplate = funcx.MustNewTemplate(
	(func(context.Context, funcx.T0, funcx.T1) error)(nil),
	(func(context.Context, funcx.T0) error)(nil),
	(func(funcx.T0, funcx.T1))(nil),
	(func(funcx.T0))(nil),
)

func NewFuncNode(n interface{}) (flow.Node, error) {
	idx, types, err := fnTemplate.MatchValue(n)
	if err != nil {
		return nil, err
	}
	defer types.Free()

	fn := &funcNode{
		fn: n,
	}

	if idx <= 1 {
		fn.hasCtx = true
	}

	var in, out reflect.Type
	t0 := types.Get(funcx.T0{})
	if t1 := types.Get(funcx.T1{}); t1 != nil {
		_, err := channel.IsAssignable(reflect.RecvDir, nil, reflect.PtrTo(t0))
		if err != nil {
			return nil, fmt.Errorf("invalid recv port: %w", err)
		}
		_, err = channel.IsAssignable(reflect.SendDir, nil, reflect.PtrTo(t1))
		if err != nil {
			return nil, fmt.Errorf("invalid send port: %w", err)
		}
		in, out = t0, t1
	} else {
		dir, _ := channel.AssignableDir(nil, reflect.PtrTo(t0))
		switch dir {
		case reflect.BothDir:
			return nil, fmt.Errorf("bidirectional port unsupported")
		case reflect.SendDir:
			out = t0
		case reflect.RecvDir:
			in = t0
		default:
			return nil, fmt.Errorf("invalid port type `%s'", t0)
		}
	}

	if in != nil {
		fn.in = reflect.New(in).Interface()
	}
	if out != nil {
		fn.out = reflect.New(out).Interface()
	}
	return fn, nil
}

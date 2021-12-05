package flow

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"go.uber.org/atomic"

	"github.com/brian14708/go-flow/flow/channel"
	"github.com/brian14708/go-flow/flow/internal/ident"
	"github.com/brian14708/go-flow/flow/port"
)

type nodeContainer struct {
	name     string
	node     Node
	nodeHash interface{}
	in       []portRef
	out      []portRef

	numActiveOut atomic.Int32
	// run info
	ctx nodeContext
}

func newNodeContainer(name string, node Node) (*nodeContainer, error) {
	var hash interface{}
	if w, ok := node.(NodeWrapper); ok {
		hash = w.NodeHash()
	} else {
		hash = node
	}

	buf := &struct {
		nodeContainer
		buf [2]portRef // small port optimization
	}{
		nodeContainer: nodeContainer{
			name:     name,
			node:     node,
			nodeHash: hash,
		},
	}

	ports := buf.buf[:0]
	nc := &buf.nodeContainer

	in, out := node.Ports()
	for k, v := range in {
		if !ident.Check(k) {
			return nil, fmt.Errorf("invalid input port name `%s'", k)
		}
		if _, ok := out[k]; ok {
			return nil, fmt.Errorf("duplicate port name `%s'", k)
		}

		port, err := newPortRef(k, v, reflect.RecvDir)
		if err != nil {
			return nil, fmt.Errorf("get input type of `%s' failed: %w", k, err)
		}
		ports = append(ports, port)
	}

	for k, v := range out {
		if !ident.Check(k) {
			return nil, fmt.Errorf("invalid output port name `%s'", k)
		}

		port, err := newPortRef(k, v, reflect.SendDir)
		if err != nil {
			return nil, fmt.Errorf("get output type of `%s' failed: %w", k, err)
		}
		ports = append(ports, port)
	}

	nc.in = ports[:len(in)]
	nc.out = ports[len(in):]

	port.RecycleMap(in)
	port.RecycleMap(out)

	return nc, nil
}

func (nc *nodeContainer) getPort(name string, dir reflect.ChanDir) *portRef {
	if dir&reflect.RecvDir != 0 {
		for i := range nc.in {
			if n := &nc.in[i]; n.name == name {
				return n
			}
		}
	}
	if dir&reflect.SendDir != 0 {
		for i := range nc.out {
			if n := &nc.out[i]; n.name == name {
				return n
			}
		}
	}
	return nil
}

func (nc *nodeContainer) setContext(ctx context.Context) {
	nc.ctx = *newNodeContext(ctx, nc)
}

func (nc *nodeContainer) run() error {
	return nc.node.Run(&nc.ctx)
}

func (nc *nodeContainer) stop() {
	nc.ctx.Cancel()
}

type portRef struct {
	Port
	name string
	ch   Chan
}

func newPortRef(name string, port interface{}, dir reflect.ChanDir) (portRef, error) {
	ref := portRef{
		name: name,
	}
	if p, ok := port.(*Port); ok {
		ref.Port = *p
	} else {
		ref.Port = Port{
			Ref: port,
		}
	}

	storageElemType, err := channel.IsAssignable(dir, ref.ElemType, reflect.TypeOf(ref.Ref))
	if err != nil {
		return portRef{}, err
	}
	if !ref.get().IsNil() {
		return portRef{}, errors.New("port is not nil")
	}
	ref.StorageElemType = storageElemType
	if ref.ElemType == nil {
		ref.ElemType = storageElemType
	}
	return ref, nil
}

func (r *portRef) set(ch Chan, dir reflect.ChanDir) {
	r.ch = ch
	ch.AssignTo(dir, r.Ref)
}

func (r *portRef) get() reflect.Value {
	return reflect.ValueOf(r.Ref).Elem()
}

func (r *portRef) addr() uintptr {
	return r.get().Pointer()
}

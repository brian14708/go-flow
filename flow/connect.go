package flow

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"go.uber.org/atomic"

	"github.com/brian14708/go-flow/flow/channel"
	"github.com/brian14708/go-flow/flow/internal/ident"
)

var anyType = reflect.TypeOf((*AnyMessage)(nil)).Elem()

func (g *Graph) Connect(srcs, dsts []string, opts ...ConnectOption) error {
	if len(g.opt.DefaultConnectOptions) > 0 {
		opts = append(g.opt.DefaultConnectOptions, opts...)
	}
	_, err := g.interceptor.Connect(ident.UniqueID(), srcs, dsts, opts...)
	return err
}

func (g *Graph) connect(id string, srcs, dsts []string, opts ...ConnectOption) (Chan, error) {
	if len(srcs) == 0 || len(dsts) == 0 {
		return nil, errors.New("connections must have src and dest")
	}

	obj := &struct {
		conn   connection
		opt    connectOptions
		cstor  channel.Storage
		reader atomic.Int32
		writer atomic.Int32
		strs   [2]string
	}{}
	conn := &obj.conn
	o := &obj.opt
	for _, opt := range opts {
		if opt, ok := opt.(connectOption); ok {
			opt.apply(o)
		}
	}

	var allPortsBuf [2]*portRef
	allPorts, err := g.appendPorts(allPortsBuf[:0], srcs, reflect.SendDir)
	if err != nil {
		return nil, fmt.Errorf("fail to get source port: %w", err)
	}
	allPorts, err = g.appendPorts(allPorts, dsts, reflect.RecvDir)
	if err != nil {
		return nil, fmt.Errorf("fail to get dest port: %w", err)
	}
	srcPortPtrs := allPorts[:len(srcs)]
	dstPortPtrs := allPorts[len(srcs):]
	srcStorageType := srcPortPtrs[0].StorageElemType
	dstStorageType := dstPortPtrs[0].StorageElemType

	// type check
	{
		for _, s := range srcPortPtrs[1:] {
			if s.StorageElemType != srcStorageType {
				return nil, fmt.Errorf("port chan type mismatch `%s', expected `%s'",
					s.StorageElemType, srcStorageType)
			}
		}
		for _, d := range dstPortPtrs {
			if d.StorageElemType != dstStorageType {
				return nil, fmt.Errorf("port chan type mismatch `%s', expected `%s'",
					d.StorageElemType, dstStorageType)
			}

			for _, s := range srcPortPtrs {
				if !convertibleTo(s.ElemType, d.ElemType, o.allowInterfaceCast) {
					return nil, fmt.Errorf("port type mismatch `%s', expected `%s'", d.ElemType, s.ElemType)
				}
			}
		}
	}

	ch, err := channel.New(srcStorageType, dstStorageType, &obj.cstor, opts...)
	if err != nil {
		return nil, err
	}

	// set chan port
	{
		for _, srcPortPtr := range srcPortPtrs {
			srcPortPtr.set(ch, reflect.SendDir)
		}
		for _, dstPortPtr := range dstPortPtrs {
			dstPortPtr.set(ch, reflect.RecvDir)
		}
		if e, ok := ch.(interface {
			NeedServe() bool
		}); !ok || e.NeedServe() {
			g.background = append(g.background, func(context.Context) error {
				ch.Serve()
				return nil
			})
		}

		obj.reader.Store(int32(len(dstPortPtrs)))
		obj.writer.Store(int32(len(srcPortPtrs)))
		g.chanMap[chanKey{dstPortPtrs[0].addr(), reflect.RecvDir}] = &obj.reader
		g.chanMap[chanKey{srcPortPtrs[0].addr(), reflect.SendDir}] = &obj.writer
	}

	// create connection entry
	{
		buf := append(obj.strs[:0], make([]string, len(srcs)+len(dsts))...)
		*conn = connection{
			id:  id,
			src: buf[:len(srcs)],
			dst: buf[len(srcs):],
			ch:  ch,
		}
		copy(conn.src, srcs)
		copy(conn.dst, dsts)
		g.conns = append(g.conns, conn)
	}

	return ch, nil
}

func (g *Graph) getPort(name string, dir reflect.ChanDir) (
	port *portRef,
	err error,
) {
	part := strings.LastIndexByte(name, ':')
	if part == -1 {
		return nil, fmt.Errorf("node without port `%s'", name)
	}

	node, ok := g.nodes[name[:part]]
	if !ok {
		return nil, fmt.Errorf("node not found `%s'", name)
	}

	if port := node.getPort(name[part+1:], dir); port != nil {
		return port, nil
	}

	return nil, fmt.Errorf("port not found `%s'", name)
}

func (g *Graph) appendPorts(
	refs []*portRef,
	names []string,
	dir reflect.ChanDir,
) (
	[]*portRef,
	error,
) {
	for _, name := range names {
		port, err := g.getPort(name, dir)
		if err != nil {
			return nil, fmt.Errorf("fail to get port: %w", err)
		}
		if !port.get().IsNil() {
			return nil, fmt.Errorf("port `%s' is not nil", name)
		}
		refs = append(refs, port)
	}
	return refs, nil
}

func convertibleTo(from, to reflect.Type, upcast bool) bool {
	if from == to {
		return true
	}
	if from == anyType || to == anyType {
		return true
	}
	if to.Kind() == reflect.Interface && from.Implements(to) {
		return true
	}
	if upcast {
		if from.Kind() == reflect.Interface && to.Kind() == reflect.Interface {
			// interface to interface cast
			return true
		}
		if from.Kind() == reflect.Interface && to.Implements(from) {
			// upcast
			return true
		}
	}
	return false
}

func (g *Graph) PortType(name string) (reflect.Type, error) {
	t, err := g.getPort(name, reflect.BothDir)
	if err != nil {
		return nil, err
	}
	return t.ElemType, nil
}

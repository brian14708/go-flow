package node

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/port"
	"github.com/brian14708/go-flow/flowtype"
)

type chanNode struct {
	ch     interface{} // chan T
	target interface{} // *chan T
	dir    reflect.ChanDir
}

func (c *chanNode) Ports() (in, out flow.PortMap) {
	switch c.dir {
	case reflect.SendDir:
		in = port.MakeMap("in", c.target)
	case reflect.RecvDir:
		out = port.MakeMap("out", c.target)
	}
	return
}

func (c *chanNode) NodeType() string {
	return reflect.TypeOf(c.ch).String()
}

func (c *chanNode) Run(ctx context.Context) error {
	target := reflect.ValueOf(c.target).Elem()
	ch := reflect.ValueOf(c.ch)

	var from, to reflect.Value
	var fromCancel <-chan struct{}
	switch c.dir {
	case reflect.SendDir:
		from, to = target, ch
	case reflect.RecvDir:
		from, to = ch, target
		fromCancel = ctx.Done()
	}

	recv := flowtype.ChanRecver(from.Interface())
	send := flowtype.ChanSender(to.Interface())

	for {
		val, ok := recv(fromCancel, true)
		if !ok {
			break
		}
		ok = send(val, nil, true)
		if !ok {
			break
		}
	}
	if c.dir == reflect.SendDir {
		ch.Close()
	}
	return ctx.Err()
}

func NewChanNode(c interface{}) (flow.Node, error) {
	t := reflect.TypeOf(c)
	if t.Kind() != reflect.Chan {
		return nil, fmt.Errorf("type `%s' is not channel", c)
	}

	dir := t.ChanDir()
	if dir != reflect.SendDir && dir != reflect.RecvDir {
		return nil, errors.New("channel direction should be send/recv")
	}

	return &chanNode{
		ch:     c,
		target: newChanPtr(reflect.BothDir&^dir, t.Elem()),
		dir:    dir,
	}, nil
}

func newChanPtr(dir reflect.ChanDir, t reflect.Type) interface{} {
	return reflect.New(reflect.ChanOf(dir, t)).Interface()
}

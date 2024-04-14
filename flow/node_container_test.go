package flow

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brian14708/go-flow/flow/channel"
)

type PortRefTester struct {
	In     <-chan int
	InBoth chan int
	Out    chan<- int
	OutBad int
}

func (s *PortRefTester) Ports() (PortMap, PortMap) {
	return PortMap{
		"in":      &s.In,
		"in_both": &s.InBoth,
		"out":     &s.Out,

		"out_bad":  &s.OutBad,
		"out_bad2": s.Out,
	}, PortMap{}
}

func (s *PortRefTester) Run(context.Context) error { return nil }

func TestNodePortRefType(t *testing.T) {
	p := new(PortRefTester)
	in, _ := p.Ports()

	testcase := [...]struct {
		name    string
		success bool
		el      reflect.Type
		dir     reflect.ChanDir
	}{
		{"in", true, reflect.TypeOf(int(0)), reflect.RecvDir},
		{"out", true, reflect.TypeOf(int(0)), reflect.SendDir},
		{"in_both", false, nil, reflect.BothDir},
		{"out_bad", false, nil, 0},
		{"out_bad2", false, nil, 0},
	}

	for _, test := range testcase {
		p, err := newPortRef(test.name, in[test.name], test.dir)
		if !test.success {
			assert.Error(t, err)
			continue
		}
		assert.Equal(t, test.el, p.ElemType)
	}
}

func TestNodePortRef(t *testing.T) {
	p := new(PortRefTester)
	in, _ := p.Ports()

	ref, err := newPortRef("in", in["in"], reflect.RecvDir)
	assert.NoError(t, err)
	assert.Equal(t, uintptr(0), ref.addr())
	v, _ := channel.New(reflect.TypeOf(0), reflect.TypeOf(0), nil)
	ref.set(v, reflect.RecvDir)
	assert.NotEqual(t, uintptr(0), ref.addr())
	assert.NotNil(t, p.In)
	v.Close()
	<-p.In // closed
}

type BadNodeTester struct {
	In, Out PortMap
}

func (s *BadNodeTester) Ports() (PortMap, PortMap) {
	return s.In, s.Out
}

func (s *BadNodeTester) Run(context.Context) error { return nil }

func TestInvalidNode(t *testing.T) {
	_, err := newNodeContainer("", new(BadNodeTester))
	assert.NoError(t, err)

	var (
		i      int
		ch     chan int
		recv   <-chan int
		send   chan<- int
		notNil = make(chan int)
	)

	testcase := [...]struct {
		in, out PortMap
	}{
		{nil, PortMap{"@bad_name": &ch}},
		{nil, PortMap{"bad_type": &i}},
		{nil, PortMap{"bad_type": ch}},
		{nil, PortMap{"bad_type": &recv}},
		{nil, PortMap{"not_nil": &notNil}},
		{PortMap{"@bad_name": &ch}, nil},
		{PortMap{"bad_type": &i}, nil},
		{PortMap{"bad_type": ch}, nil},
		{PortMap{"not_nil": &notNil}, nil},
		{PortMap{"bad_type": &send}, nil},
		{PortMap{"duplicate": &ch}, PortMap{"duplicate": &ch}},
	}
	for _, test := range testcase {
		_, err := newNodeContainer("", &BadNodeTester{
			In:  test.in,
			Out: test.out,
		})
		assert.Error(t, err)
	}
}

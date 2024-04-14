package flow

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"

	"github.com/brian14708/go-flow/flow/channel"
	"github.com/brian14708/go-flow/flowtype/testutil"
)

func TestGraphConnect(t *testing.T) {
	out := make(chan int, 2)
	g, _ := NewGraph(nil)
	assert.NoError(t, g.AddNode("g1", GeneratorNode(0)))
	assert.NoError(t, g.AddNode("sum1", SumNode(out)))
	assert.NoError(t, g.AddNode("g2", GeneratorNode(0)))
	assert.NoError(t, g.AddNode("float", new(FloatNode)))

	testcase := [...]struct {
		srcs, dsts []string
	}{
		{nil, []string{"sum1:in"}},
		{[]string{"g1:out"}, nil},
		{[]string{"g1:out", "missing"}, []string{"sum1:in"}},
		{[]string{"missing", "g1:out"}, []string{"sum1:in"}},
		{[]string{"g1:out"}, []string{"sum1:in", "missing"}},
		{[]string{"g1:out"}, []string{"missing", "sum1:in"}},
		{[]string{"missing:X", "g1:out"}, []string{"sum1:in"}},
		{[]string{"g1:out"}, []string{"missing:", "sum1:in"}},
		{[]string{"g1:out", "float:out"}, []string{"sum1:in"}},
		{[]string{"g1:out"}, []string{"sum1:in", "float:in"}},
	}

	for _, test := range testcase {
		assert.Error(t, g.Connect(test.srcs, test.dsts))
	}

	assert.NoError(t, g.Connect([]string{"g1:out"}, []string{"sum1:in"}))
	assert.Error(t, g.Connect([]string{"g1:out"}, []string{"sum1:in"}))
	assert.Error(t, g.Connect([]string{"g2:out"}, []string{"float:in"}))
}

func TestGraphGetPort(t *testing.T) {
	g, _ := NewGraph(nil)
	assert.NoError(t, g.AddNode("g1", GeneratorNode(0)))

	testcase := [...]struct {
		success  bool
		portName string
		dir      reflect.ChanDir
	}{
		{true, "g1:in", reflect.RecvDir},
		{true, "g1:out", reflect.SendDir},
		{false, "g1:notfound", 0},
		{false, "notfound", 0},
	}

	for _, test := range testcase {
		var err error
		_, err = g.getPort(test.portName, reflect.BothDir)
		if test.success {
			assert.NoError(t, err)
			_, err = g.getPort(test.portName, test.dir)
			assert.NoError(t, err)
			_, err = g.getPort(test.portName, reflect.BothDir&^test.dir)
			assert.Error(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestGraphPortType(t *testing.T) {
	g, _ := NewGraph(nil)
	assert.NoError(t, g.AddNode("g1", GeneratorNode(1)))

	ty, err := g.PortType("g1:out")
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(0), ty)

	_, err = g.PortType("g1:notfound")
	assert.Error(t, err)
}

func TestGraphObserver(t *testing.T) {
	testutil.TestWithNilRegistry(t, func(t *testing.T) {
		out := make(chan int, 5)
		cnt := 0
		g, _ := NewGraph(new(GraphOptions))
		assert.NoError(t, g.AddNode("g1", GeneratorNode(1)))
		assert.NoError(t, g.AddNode("sum1", SumNode(out)))
		assert.NoError(t, g.Connect([]string{"g1:out"}, []string{"sum1:in"},
			channel.WithObserver(func(interface{}) {
				cnt++
			}),
		))
		assert.NoError(t, g.Run(context.Background()))
		assert.Equal(t, 5, cnt)
	})
}

func TestChanSize(t *testing.T) {
	g, _ := NewGraph(new(GraphOptions))
	var called atomic.Int32
	assert.NoError(t, g.AddNode("t1", FnNode(func(ctx context.Context, in <-chan int, out chan<- int) error {
		_, f := FromContext(ctx)
		assert.Equal(t, 99, f.GetChan("out").Cap())
		assert.Equal(t, 0, f.GetChan("out").Len())
		assert.Nil(t, f.GetChan("x"))
		called.Inc()
		return nil
	})))
	assert.NoError(t, g.AddNode("t2", FnNode(func(ctx context.Context, in <-chan int, out chan<- int) error {
		_, f := FromContext(ctx)
		assert.Equal(t, 99, f.GetChan("in").Cap())
		called.Inc()
		return nil
	})))
	assert.Panics(t, func() {
		_ = g.Connect([]string{"t1:out"}, []string{"t2:in"},
			WithChanSize(-99),
		)
	})
	assert.NoError(t, g.Connect([]string{"t1:out"}, []string{"t2:in"},
		WithChanSize(99),
		channel.WithObserver(func(interface{}) {}),
	))
	assert.NoError(t, g.Run(context.Background()))
	assert.Equal(t, int32(2), called.Load())
}

func TestConnectType(t *testing.T) {
	testutil.TestWithNilRegistry(t, func(t *testing.T) {
		var (
			reader io.Reader     = new(bytes.Buffer)
			rw     io.ReadWriter = new(bytes.Buffer)
			writer io.Writer     = new(bytes.Buffer)
			any    AnyMessage    = new(bytes.Buffer)
			s                    = struct{}{}
		)
		testcase := [...]struct {
			success bool
			o, i    Node
			cvt     bool
		}{
			// primitive types
			{true, NewTypeNode(2), NewTypeNode(1), false},
			{true, NewTypeNode(0.5), NewTypeNode(0.6), false},
			{false, NewTypeNode(5), NewTypeNode(0.5), false},
			{false, NewTypeNode(0.5), NewTypeNode(5), false},
			{false, NewTypeNode(0.5), NewTypeNode(5), true},

			// struct types
			{true, NewTypeNode(new(bytes.Buffer)), NewTypeNode(new(bytes.Buffer)), false},

			// any
			{true, NewTypeNode(new(bytes.Buffer)), NewInterfaceTypeNode(&any), false},
			{true, NewInterfaceTypeNode(&any), NewTypeNode(new(bytes.Buffer)), false},

			// interface to type
			{true, NewTypeNode(new(bytes.Buffer)), NewInterfaceTypeNode(&reader), false},
			{true, NewInterfaceTypeNode(&reader), NewTypeNode(new(bytes.Buffer)), true},
			{false, NewInterfaceTypeNode(&reader), NewTypeNode(new(bytes.Buffer)), false},
			{true, NewInterfaceTypeNode(&reader), NewInterfaceTypeNode(&reader), false},

			// interface to interface
			{false, NewInterfaceTypeNode(&reader), NewInterfaceTypeNode(&writer), false},
			{true, NewInterfaceTypeNode(&reader), NewInterfaceTypeNode(&writer), true},
			{false, NewInterfaceTypeNode(&reader), NewInterfaceTypeNode(&rw), false},
			{true, NewInterfaceTypeNode(&reader), NewInterfaceTypeNode(&rw), true},

			{false, NewInterfaceTypeNode(&reader), NewTypeNode(s), true},
		}

		for _, test := range testcase {
			g, _ := NewGraph(nil)
			assert.NoError(t, g.AddNode("g1", test.o))
			assert.NoError(t, g.AddNode("g2", test.i))
			var err error
			if test.cvt {
				err = g.Connect([]string{"g1:out"}, []string{"g2:in"}, WithInterfaceCast())
			} else {
				err = g.Connect([]string{"g1:out"}, []string{"g2:in"})
			}
			if test.success {
				assert.NoError(t, err)
				if err == nil {
					assert.NoError(t, g.Run(context.Background()))
				}
			} else {
				assert.Error(t, err)
			}
		}
	})
}

func BenchmarkGraphObserver(b *testing.B) {
	for _, cnt := range []int{0, 4, 16} {
		b.Run(strconv.Itoa(cnt), func(b *testing.B) {
			in := make(chan int, 1)
			out := make(chan int, 1)

			g, _ := NewGraph(nil)
			_ = g.AddNode("g1", SourceChan(in))
			_ = g.AddNode("s1", SinkChan(out))
			var opt []ConnectOption
			for i := 0; i < cnt; i++ {
				opt = append(opt, channel.WithObserver(func(interface{}) {}))
			}
			_ = g.Connect([]string{"g1:out"}, []string{"s1:in"}, opt...)
			go func() { _ = g.Run(context.Background()) }()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				in <- i
				<-out
			}
			close(in)
		})
	}
}

func BenchmarkAnyMessage(b *testing.B) {
	in := make(chan int, 1)
	out := make(chan int, 1)

	g, _ := NewGraph(nil)
	_ = g.AddNode("g1", SourceChan(in))
	_ = g.AddNode("w1", NoopNode())
	_ = g.AddNode("s1", SinkChan(out))
	_ = g.Connect([]string{"g1:out"}, []string{"w1:in"})
	_ = g.Connect([]string{"w1:out"}, []string{"s1:in"})
	go func() { _ = g.Run(context.Background()) }()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		in <- i
		<-out
	}
	close(in)
}

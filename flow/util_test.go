package flow

import (
	"context"
	"reflect"

	"github.com/brian14708/go-flow/flow/port"
)

func FnNode(f func(context.Context, <-chan int, chan<- int) error) Node {
	return &fnNode{fn: f}
}

type fnNode struct {
	in  <-chan int
	out chan<- int
	fn  func(context.Context, <-chan int, chan<- int) error
}

func (f *fnNode) Ports() (in, out PortMap) {
	return port.MakeMap("in", &f.in), port.MakeMap("out", &f.out)
}

func (f *fnNode) Run(ctx context.Context) error {
	return f.fn(ctx, f.in, f.out)
}

func NoopNode() Node {
	return FnNode(func(_ context.Context, in <-chan int, out chan<- int) error {
		for i := range in {
			out <- i
		}
		return nil
	})
}

func ReturnNode() Node {
	return FnNode(func(_ context.Context, in <-chan int, out chan<- int) error {
		return nil
	})
}

func WaitNode() Node {
	return FnNode(func(ctx context.Context, _ <-chan int, _ chan<- int) error {
		<-ctx.Done()
		return nil
	})
}

func GeneratorNode(val int) Node {
	return FnNode(func(_ context.Context, _ <-chan int, out chan<- int) error {
		for i := 0; i < 5; i++ {
			out <- val
		}
		return nil
	})
}

func SumNode(ch chan int) Node {
	return FnNode(func(_ context.Context, in <-chan int, _ chan<- int) error {
		total := 0
		for i := range in {
			total += i
		}
		ch <- total
		return nil
	})
}

func SourceChan(ch chan int) Node {
	return FnNode(func(_ context.Context, _ <-chan int, out chan<- int) error {
		for i := range ch {
			out <- i
		}
		return nil
	})
}

func SinkChan(ch chan int) Node {
	return FnNode(func(_ context.Context, in <-chan int, _ chan<- int) error {
		for i := range in {
			ch <- i
		}
		close(ch)
		return nil
	})
}

type FloatNode struct {
	In  <-chan float32
	Out chan<- float32
}

func (f *FloatNode) Ports() (PortMap, PortMap) {
	return PortMap{
			"in": &f.In,
		}, PortMap{
			"out": &f.Out,
		}
}

func (*FloatNode) Run(ctx context.Context) error {
	return nil
}

type TypeNode struct {
	t   reflect.Type
	val interface{}
	in  interface{}
	out interface{}
}

func NewTypeNode(i interface{}) *TypeNode {
	t := reflect.TypeOf(i)
	return &TypeNode{
		t:   t,
		val: i,
		in:  reflect.New(reflect.ChanOf(reflect.BothDir, t)).Interface(),
		out: reflect.New(reflect.ChanOf(reflect.BothDir, t)).Interface(),
	}
}

func NewInterfaceTypeNode(i interface{}) *TypeNode {
	t := reflect.TypeOf(i).Elem()
	return &TypeNode{
		t:   t,
		val: reflect.ValueOf(i).Elem().Interface(),
		in:  reflect.New(reflect.ChanOf(reflect.BothDir, t)).Interface(),
		out: reflect.New(reflect.ChanOf(reflect.BothDir, t)).Interface(),
	}
}

func (g *TypeNode) Ports() (PortMap, PortMap) {
	return PortMap{
			"in": g.in,
		}, PortMap{
			"out": g.out,
		}
}

func (g *TypeNode) Run(ctx context.Context) error {
	in := reflect.ValueOf(g.in).Elem()
	out := reflect.ValueOf(g.out).Elem()
	if in.IsNil() {
		out.Send(reflect.Zero(g.t))
		out.Send(reflect.ValueOf(g.val))
	} else {
		for {
			_, ok := in.Recv()
			if !ok {
				return nil
			}
		}
	}
	return nil
}

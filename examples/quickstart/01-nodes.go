package main

import (
	"context"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/pipeline"
)

func init() {
	initFunc = append(initFunc, nodeExample)
}

func nodeExample() {
	ppl := pipeline.New(attachProfiler(&flow.GraphOptions{
		ID: "01-nodes",
	}))

	in := make(chan int)
	ppl.Add("ChanNode", (<-chan int)(in))
	ppl.Add("FuncNode_InOut", func(context.Context, <-chan int, chan<- int) error { return nil })
	ppl.Add("NodeInterface", new(simpleNode))
	ppl.Add("StructTag", new(simpleTagNode))
	ppl.Add("FuncNode_InOnly", func(context.Context, <-chan int) error { return nil })
	close(in)

	err := ppl.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

type simpleNode struct {
	In  <-chan int
	Out chan<- int
}

func (s *simpleNode) Ports() (in, out flow.PortMap) {
	return flow.PortMap{
			"my_in": &s.In,
		}, flow.PortMap{
			"my_out": &s.Out,
		}
}

func (s *simpleNode) Run(context.Context) error {
	return nil
}

type simpleTagNode struct {
	In  <-chan int `pipeline:"my_in"`
	Out chan<- int `pipeline:"my_out"`
}

func (s *simpleTagNode) Run(context.Context) error {
	return nil
}

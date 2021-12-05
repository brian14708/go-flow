package main

import (
	"context"
	"math"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/pipeline"
)

func init() {
	initFunc = append(initFunc, multiOutputExample)
}

func multiOutputExample() {
	ppl := pipeline.New(attachProfiler(&flow.GraphOptions{
		ID: "05-multi-output",
	}))

	ppl.Add("Generator", func(ctx context.Context, out chan<- int) error {
		out <- 4
		out <- 3
		return nil
	})
	ppl.Add("Switch3", new(switchNode))
	ppl.Merge(
		ppl.SplitOutput("if_3").Add("Pow8", makePowPipeline(8)),
		ppl.SplitOutput("else").Add("Pow4", makePowPipeline(4)),
	)
	ppl.Add("Sum", func(ctx context.Context, in <-chan int) error {
		sum := 0
		for i := range in {
			sum += i
		}
		if sum != int(math.Pow(3, 8)+math.Pow(4, 4)) {
			panic("wrong answer")
		}
		return nil
	})

	err := ppl.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

type switchNode struct {
	In   <-chan int `pipeline:"in"`
	Out3 chan<- int `pipeline:"if_3"`
	Out  chan<- int `pipeline:"else"`
}

func (s *switchNode) Run(context.Context) error {
	for i := range s.In {
		if i == 3 {
			s.Out3 <- i
		} else {
			s.Out <- i
		}
	}
	return nil
}

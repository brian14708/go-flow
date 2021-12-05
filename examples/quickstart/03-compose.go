package main

import (
	"context"
	"fmt"
	"math"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/pipeline"
)

func init() {
	initFunc = append(initFunc, composeExample)
}

func makePowPipeline(n int) func(*pipeline.Pipeline) {
	if n&(n-1) != 0 {
		panic("must be power of 2")
	}
	return func(ppl *pipeline.Pipeline) {
		ppl.Add("Square", func(ctx context.Context, in <-chan int, out chan<- int) error {
			for i := range in {
				out <- i * i
			}
			return nil
		})
		n /= 2
		if n > 1 {
			ppl.Add(fmt.Sprintf("Pow%d", n), makePowPipeline(n))
		}
	}
}

func composeExample() {
	ppl := pipeline.New(attachProfiler(&flow.GraphOptions{
		ID: "03-compose",
	}))

	ppl.
		Add("Generator", func(ctx context.Context, out chan<- int) error {
			out <- 4
			out <- 3
			return nil
		}).
		Add("Pow8", makePowPipeline(8)).
		Add("Sum", func(ctx context.Context, in <-chan int) error {
			sum := 0
			for i := range in {
				sum += i
			}
			if sum != int(math.Pow(3, 8)+math.Pow(4, 8)) {
				panic("wrong answer")
			}
			return nil
		})

	err := ppl.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

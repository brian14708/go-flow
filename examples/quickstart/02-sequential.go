package main

import (
	"context"
	"fmt"
	"math"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/pipeline"
)

func init() {
	initFunc = append(initFunc, seqentialExample)
}

func seqentialExample() {
	ppl := pipeline.New(attachProfiler(&flow.GraphOptions{
		ID: "02-seqential",
	}))

	ppl.Add("Generator", func(ctx context.Context, out chan<- int) error {
		out <- 4
		out <- 3
		return nil
	})
	for i := 0; i < 3; i++ {
		ppl.Add(fmt.Sprintf("Square_%d", i), func(ctx context.Context, in <-chan int, out chan<- int) error {
			for i := range in {
				out <- i * i
			}
			return nil
		})
	}
	ppl.Add("Sum", func(ctx context.Context, in <-chan int) error {
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

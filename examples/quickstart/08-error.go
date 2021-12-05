package main

import (
	"context"
	"errors"
	"time"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/funcop"
	"github.com/brian14708/go-flow/flow/pipeline"
)

func init() {
	initFunc = append(initFunc, errorExample)
}

func errorExample() {
	in := make(chan string, 10)
	out := make(chan int)

	ppl := pipeline.New(attachProfiler(&flow.GraphOptions{
		ID: "08-error",
	}))
	ppl.Add("in", (<-chan string)(in))
	ppl.Add("count",
		funcop.Map(func(_ context.Context, s string) (int, error) {
			time.Sleep(1 * time.Millisecond)
			if len(s) == 0 {
				return 0, errors.New("Custom Error")
			}
			return len(s), nil
		}),
	)
	ppl.Add("delay",
		funcop.Map(func(i int) int {
			time.Sleep(100 * time.Millisecond)
			return i
		}),
	)
	ppl.Add("out", (chan<- int)(out))

	in <- "Hello world"
	in <- "hello world"
	in <- ""
	in <- "hello world"
	close(in)
	go ppl.Run(context.Background())
	for range out {
	}
}

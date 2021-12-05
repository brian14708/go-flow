package main

import (
	"context"
	"runtime"
	"strings"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/funcop"
	"github.com/brian14708/go-flow/flow/pipeline"
)

func init() {
	initFunc = append(initFunc, funcExample)
}

func funcExample() {
	in := make(chan string, 10)
	out := make(chan int)

	ppl := pipeline.New(attachProfiler(&flow.GraphOptions{
		ID: "07-func",
	}))
	ppl.Add("", (<-chan string)(in))
	ppl.Add("",
		funcop.FlatMap(
			(&stringSplitter{" "}).Do,
			funcop.WithParallel(runtime.NumCPU(), false),
		),
	)
	ppl.Add("",
		// remove all words with "or"
		funcop.Filter(func(s string) bool {
			return !strings.Contains(s, "or")
		}, funcop.WithParallel(runtime.NumCPU(), false)),
	)
	ppl.Add("",
		funcop.Map(func(s string) int {
			return len(s)
		}, funcop.WithParallel(runtime.NumCPU(), false)),
	)
	ppl.Add("", (chan<- int)(out))

	in <- "Hello world"
	in <- "hello world"
	close(in)
	go ppl.Run(context.Background())
	var sum int
	for o := range out {
		sum += o
	}
	// Hello hello
	if sum != 10 {
		panic("wrong answer")
	}
}

type stringSplitter struct{ s string }

func (s *stringSplitter) Do(i string) []string {
	return strings.Split(i, s.s)
}

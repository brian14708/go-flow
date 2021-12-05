package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/pipeline"
)

func init() {
	initFunc = append(initFunc, profilerExample)
}

type CustomMessage struct {
	value int
}

func profilerExample() {
	ppl := pipeline.New(attachProfiler(&flow.GraphOptions{
		ID: "06-profiler",
	}))

	ppl.Add("Generator", func(ctx context.Context, out chan<- CustomMessage) error {
		for {
			t := CustomMessage{
				value: 4,
			}
			out <- t
			t = CustomMessage{
				value: 3,
			}
			out <- t
		}
	})
	ppl.Add("Switch3", new(customSwitchNode))
	ppl.Merge(
		ppl.SplitOutput("if_3").Add("delay", func(ctx context.Context, in <-chan CustomMessage, out chan<- CustomMessage) error {
			ch := make(chan struct{}, 3)
			for i := range in {
				ch <- struct{}{}
				i := i
				time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
				<-ch
				out <- i
			}
			return nil
		}),
		ppl.SplitOutput("else").Add("Pow4", makeCustomPowPpl(4)),
	)
	ppl.Add("Sum", func(ctx context.Context, in <-chan CustomMessage) error {
		sum := 0
		for i := range in {
			sum += i.value
		}
		return nil
	})

	go func() {
		err := ppl.Run(context.Background())
		if err != nil {
			panic(err)
		}
	}()
}

type customSwitchNode struct {
	In   <-chan CustomMessage `pipeline:"in"`
	Out3 chan<- CustomMessage `pipeline:"if_3"`
	Out  chan<- CustomMessage `pipeline:"else"`
}

func (s *customSwitchNode) Run(context.Context) error {
	for i := range s.In {
		if i.value == 3 {
			s.Out3 <- i
		} else {
			s.Out <- i
		}
	}
	return nil
}

func makeCustomPowPpl(n int) func(*pipeline.Pipeline) {
	if n&(n-1) != 0 {
		panic("must be power of 2")
	}
	return func(ppl *pipeline.Pipeline) {
		ppl.Add("Square", func(ctx context.Context, in <-chan CustomMessage, out chan<- CustomMessage) error {
			for i := range in {
				i.value *= i.value
				out <- i
			}
			return nil
		})
		n /= 2
		if n > 1 {
			ppl.Add(fmt.Sprintf("Pow%d", n), makeCustomPowPpl(n))
		}
	}
}

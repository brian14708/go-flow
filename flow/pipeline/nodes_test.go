package pipeline

import (
	"context"
)

type NoopIntNode struct {
	In  <-chan int `pipeline:"in"`
	Out chan<- int `pipeline:"out"`
}

func (n *NoopIntNode) Run(context.Context) error {
	for i := range n.In {
		n.Out <- i
	}
	return nil
}

type SumIntNode struct {
	InA <-chan int `pipeline:"in1"`
	InB <-chan int `pipeline:"in2"`
	Out chan<- int `pipeline:"out"`
}

func (s *SumIntNode) Run(context.Context) error {
	for i := range s.InA {
		s.Out <- i + <-s.InB
	}
	return nil
}

type Split struct {
	In   <-chan int     `pipeline:"in"`
	OutA chan<- int     `pipeline:"out1"`
	OutB chan<- float32 `pipeline:"out2"`
	OutC chan<- int     `pipeline:"out3"`
}

func (s *Split) Run(ctx context.Context) error {
	for i := range s.In {
		s.OutA <- i
		s.OutB <- float32(i)
	}
	return nil
}

type Join struct {
	InA <-chan int     `pipeline:"in1"`
	InB <-chan float32 `pipeline:"in2"`
	Out chan<- int     `pipeline:"out1"`
}

func (s *Join) Run(ctx context.Context) error {
	for a := range s.InA {
		s.Out <- a
		s.Out <- int(<-s.InB)
	}
	return nil
}

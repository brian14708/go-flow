package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/funcop"
	"github.com/brian14708/go-flow/flow/pipeline"
	"github.com/brian14708/go-flow/flow/rrworker"
)

func init() {
	initFunc = append(initFunc, rrworkerExample)
}

type MyTask struct {
	rrworker.Task
	Value string
	tmp   int
}

// declare result type
func (m *MyTask) SetResult(v int) {
	m.SetResultAny(v)
}

func startWorker() *rrworker.RRWorker {
	ppl := pipeline.New(attachProfiler(&flow.GraphOptions{
		ID: "09-rrworker-worker",
	}))

	ppl.Add("count",
		funcop.Map(func(s *MyTask) *MyTask {
			time.Sleep(1 * time.Millisecond)
			s.tmp = len(s.Value)
			return s
		}),
	)
	ppl.Add("delay",
		funcop.Map(func(i *MyTask) *MyTask {
			time.Sleep(100 * time.Millisecond)
			return i
		}),
	)
	ppl.Add("output",
		funcop.Map(func(i *MyTask) {
			i.SetResult(i.tmp)
		}),
	)
	w, err := rrworker.New(ppl)
	if err != nil {
		panic(err)
	}
	go w.Run(context.Background())
	return w
}

func rrworkerExample() {
	w := startWorker()
	for i := 0; i < 2; i++ {
		i := i
		go func() {
			ppl := pipeline.New(attachProfiler(&flow.GraphOptions{
				ID: fmt.Sprintf("09-rrworker-graph%d", i),
			}))
			ppl.Add("in", func(_ context.Context, ch chan<- string) error {
				for {
					ch <- "abcd"
				}
			})
			ppl.Add("run_worker", rrworker.SubmitNode(w,
				rrworker.WithTaskPreparer(func(s string) *MyTask {
					return &MyTask{Value: s}
				}),
				rrworker.WithErrorHandler(func(err error) error {
					// handle worker error
					return err
				}),
				rrworker.WithParallel(runtime.NumCPU(), true),
			))
			ppl.Add("out", func(_ context.Context, i <-chan int) error {
				for range i {
				}
				return nil
			})
			ppl.Run(context.Background())
		}()
	}
}

package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/VividCortex/ewma"
	"github.com/sirupsen/logrus"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/channel"
	"github.com/brian14708/go-flow/flow/funcop"
	"github.com/brian14708/go-flow/flow/pipeline"
)

func init() {
	initFunc = append(initFunc, adaptiveExample)
}

func testLatency(name string, opts ...flow.ConnectOption) ewma.MovingAverage {
	ppl := pipeline.New(attachProfiler(&flow.GraphOptions{
		ID:                    name,
		DefaultConnectOptions: opts,
	}))
	const N = 1000
	ppl.Add("", func(out chan<- time.Time) {
		for i := 0; i < N; i++ {
			out <- time.Now()
		}
	})
	for i := 0; i < 10; i++ {
		ppl.Add("",
			funcop.Map(func(i time.Time) time.Time {
				time.Sleep(time.Duration(10+rand.Intn(5)) * time.Millisecond)
				return i
			}, funcop.WithParallel(4, false)),
		)
	}
	latency := ewma.NewMovingAverage(5)
	ppl.Add("", funcop.Map(func(i time.Time) {
		latency.Add(float64(time.Since(i)) / float64(time.Second))
	}))
	ppl.Run(context.Background())
	return latency
}

func adaptiveExample() {
	go func() {
		adap := testLatency("10-adaptive",
			channel.WithSize(1024),
			channel.WithAdaptiveGain(1),
		)
		fixed := testLatency("10-fixed",
			flow.WithChanSize(1024),
		)
		logrus.Infof("adaptive latency: %fs", adap.Value())
		logrus.Infof("fixed latency: %fs", fixed.Value())
	}()
}

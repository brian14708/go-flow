package main

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flowdebug"
)

var initFunc []func()

var defaultProfiler *flowdebug.Profiler

func attachProfiler(o *flow.GraphOptions) *flow.GraphOptions {
	o.Interceptors = append(o.Interceptors,
		flowdebug.GraphInterceptor(defaultProfiler),
	)
	return o
}

func main() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	defaultProfiler = flowdebug.NewProfiler(&flowdebug.ProfilerOptions{
		Logger: log,
	})

	for _, i := range initFunc {
		i()
	}

	http.Handle("/", defaultProfiler)
	log.Println(http.ListenAndServe("localhost:6060", nil))
}

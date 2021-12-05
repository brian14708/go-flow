package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/brian14708/go-flow/flow"
	"github.com/brian14708/go-flow/flow/pipeline"
	"github.com/brian14708/go-flow/flowdebug"
)

type CaptureMessage struct{}

func makeAlert(ppl *pipeline.Pipeline) {
	ppl.Add("searcher", func(context.Context, <-chan CaptureMessage, chan<- CaptureMessage) error {
		return nil
	})
	ppl.Add("limiter", func(context.Context, <-chan CaptureMessage, chan<- CaptureMessage) error {
		return nil
	})
	ppl.Add("storage", func(context.Context, <-chan CaptureMessage, chan<- CaptureMessage) error {
		return nil
	})
	ppl.Add("sink", func(context.Context, <-chan CaptureMessage) error {
		return nil
	})
}

func makeCapture(item, bind, frame *pipeline.Pipeline) {
	item.Add("captured.filter_quality", func(context.Context, <-chan CaptureMessage, chan<- CaptureMessage) error {
		return nil
	})
	item.Add("captured.storage", func(context.Context, <-chan CaptureMessage, chan<- CaptureMessage) error {
		return nil
	})
	item.Add("captured.traj", new(MergeTraj),
		pipeline.SideInput(frame, "in_frame"),
	)
	item.Add("captured.capture_pub", func(context.Context, <-chan CaptureMessage, chan<- CaptureMessage) error {
		return nil
	})
	item.Add("captured.binder", new(MergeBind),
		pipeline.SideInput(bind, "in_bind"),
	)
	item.Add("captured.link_pub", func(context.Context, <-chan CaptureMessage) error {
		return nil
	})
}

func makeDispatch(item, frame, bind *pipeline.Pipeline) {
	item.Add("dispatch", new(Dispatch),
		pipeline.SideInput(frame, "in_frame"),
		pipeline.SideInput(bind, "in_bind"),
	)

	item.SplitOutput("alert_item").Add("alertd", makeAlert)
	item.SplitOutput("preview_frame").Add("previewd.forward", func(context.Context, <-chan CaptureMessage) error {
		return nil
	})
	makeCapture(
		item.SplitOutput("capture_item"),
		item.SplitOutput("capture_bind"),
		item.SplitOutput("capture_frame"),
	)
}

func main() {
	prof := flowdebug.NewProfiler(nil)

	ppl := pipeline.New(&flow.GraphOptions{
		ID: "A",
		Interceptors: []flow.Interceptor{
			flowdebug.GraphInterceptor(prof),
		},
	})
	ppl.Add("connector", func(ctx context.Context, in chan<- CaptureMessage) error {
		for {
			in <- CaptureMessage{}
			time.Sleep(time.Second)
		}
	})
	ppl.Add("fetcher",
		pipeline.OrderedParallel([]interface{}{
			func(ctx context.Context, in <-chan CaptureMessage, out chan<- CaptureMessage) error {
				for range in {
				}
				return nil
			},
			func(ctx context.Context, in <-chan CaptureMessage, out chan<- CaptureMessage) error {
				for range in {
				}
				return nil
			},
			func(ctx context.Context, in <-chan CaptureMessage, out chan<- CaptureMessage) error {
				for range in {
				}
				return nil
			},
			func(ctx context.Context, in <-chan CaptureMessage, out chan<- CaptureMessage) error {
				for range in {
				}
				return nil
			},
		}),
	)
	ppl.Add("decoder", func(context.Context, <-chan CaptureMessage, chan<- CaptureMessage) error {
		return nil
	})

	ppl.Add("anaylzer", new(Split))
	engine := ppl.SplitOutput("has_feat").Add("engine", new(Engine))

	makeDispatch(
		engine.SplitOutput("item"),
		engine.SplitOutput("frame").Merge(
			ppl.SplitOutput("no_feat"),
		),
		engine.SplitOutput("bind"),
	)

	go func() {
		http.Handle("/", prof)
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	fmt.Println(flowdebug.Graphviz(ppl.Graph().Topology()))
	err := ppl.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

type Split struct {
	In         <-chan CaptureMessage `pipeline:"in"`
	HasFeature chan<- CaptureMessage `pipeline:"has_feat"`
	NoFeature  chan<- CaptureMessage `pipeline:"no_feat"`
}

func (s *Split) Run(ctx context.Context) error {
	return nil
}

type Engine struct {
	In <-chan CaptureMessage `pipeline:"in"`

	Item  chan<- CaptureMessage `pipeline:"item"`
	Frame chan<- CaptureMessage `pipeline:"frame"`
	Bind  chan<- CaptureMessage `pipeline:"bind"`
}

func (s *Engine) Run(ctx context.Context) error {
	return nil
}

type Dispatch struct {
	InItem  <-chan CaptureMessage `pipeline:"in_item"`
	InFrame <-chan CaptureMessage `pipeline:"in_frame"`
	InBind  <-chan CaptureMessage `pipeline:"in_bind"`

	ItemForAlert    chan<- CaptureMessage `pipeline:"alert_item"`
	FrameForPreview chan<- CaptureMessage `pipeline:"preview_frame"`

	Item  chan<- CaptureMessage `pipeline:"capture_item"`
	Frame chan<- CaptureMessage `pipeline:"capture_frame"`
	Bind  chan<- CaptureMessage `pipeline:"capture_bind"`
}

func (s *Dispatch) Run(ctx context.Context) error {
	return nil
}

type MergeTraj struct {
	InItem  <-chan CaptureMessage `pipeline:"in_item"`
	InFrame <-chan CaptureMessage `pipeline:"in_frame"`
	Item    chan<- CaptureMessage `pipeline:"item"`
}

func (s *MergeTraj) Run(ctx context.Context) error {
	return nil
}

type MergeBind struct {
	InItem <-chan CaptureMessage `pipeline:"in_item"`
	InBind <-chan CaptureMessage `pipeline:"in_bind"`
	Item   chan<- CaptureMessage `pipeline:"item"`
}

func (s *MergeBind) Run(ctx context.Context) error {
	return nil
}

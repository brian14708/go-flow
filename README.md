# go-flow

[![Go Reference](https://pkg.go.dev/badge/github.com/brian14708/go-flow.svg)](https://pkg.go.dev/github.com/brian14708/go-flow)

A Go library for building type-safe dataflow pipelines using graphs, nodes, and channels.

## Features

- Graph-based execution with typed connections
- Functional operations: Map, Filter, FlatMap
- Built-in profiling and debugging
- Concurrent by design

## Quick Start

```go
p, _ := pipeline.New()

p.Add("gen", flow.GeneratorFunc(func(ctx context.Context, out chan<- int) error {
    for i := 1; i <= 5; i++ {
        out <- i
    }
    return nil
}))

p.Add("double", funcop.Map(func(x int) int { return x * 2 }))

p.Add("print", flow.ConsumerFunc(func(ctx context.Context, in <-chan int) error {
    for val := range in {
        fmt.Printf("%d\n", val)
    }
    return nil
}))

p.Run(context.Background())
```

## Installation

```bash
go get github.com/brian14708/go-flow
```

## Core Concepts

**Graph**: Container for nodes and connections

```go
g := flow.NewGraph()
g.AddNode("node1", myNode)
g.Connect(flow.Src{"node1:out"}, flow.Dst{"node2:in"})
```

**Pipeline**: Fluent builder API

```go
p, _ := pipeline.New()
p.Add("source", sourceNode).Add("transform", transformNode)
```

**Functional Ops**: Map, Filter, FlatMap

```go
funcop.Map(func(x int) string { return fmt.Sprintf("%d", x) })
funcop.Filter(func(x int) bool { return x > 0 })
```

## Debugging

```go
profiler := flowdebug.NewProfiler()
g := flow.NewGraph(flowdebug.GraphInterceptor(profiler))
go profiler.ServeHTTP(":8080")  // Visit http://localhost:8080
g.Run(ctx)
```

## Documentation

See [examples/](examples/) for complete examples.

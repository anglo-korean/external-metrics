# externalmetrics

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/anglo-korean/external-metrics)
[![Go Report Card](https://goreportcard.com/badge/github.com/anglo-korean/external-metrics)](https://goreportcard.com/report/github.com/anglo-korean/external-metrics)

package externalmetrics provides a dirt simple, lightweight, easily exstensible external metrics
api server for Kubernetes.

Why?

The officialy supported method ([https://github.com/kubernetes-sigs/custom-metrics-apiserver](https://github.com/kubernetes-sigs/custom-metrics-apiserver)) is just f-ing inscruitable.

Not only that, but the documentation for what these servers need to do are awful. The best bet is a design doc on an old branch of some random kubernetes repo. The second bet is a medium article where someone describes how they discovered what one of these APIs needs to do, but not what that is.

This package is pretty simple; it has a New() function, an AddMetric() function, and a Serve() function.

No thrills, no tunables, bugger all excitement.

A complete form of the following example can be found in `[./example](./example)`, but the nuts and bolts look like:

```go
func main() {
    s := externalmetrics.New()
    s.AddMetric(context.Background(), externalmetrics.NamespaceAll, "my-metric", externalmetrics.Tick(time.Second), someIncrementingMetric)
    panic(s.Serve(listenAddr))
}
```

The function `someIncrementingMetric` might look like:

```go
func someIncrementingMetric(ctx context.Context, namespace, name string) (result externalmetrics.Value, err error) {
    v := externalmetrics.NewValue(counterA)
    v.AddSelector("some-key", "A", counterA)
    v.AddSelector("some-key", "B", counterB)
    v.AddSelector("some-key", "C", counterC)
    return v, nil
}
```

The metrics exposed by this service can then be exposed in an HPA

## Functions

### func [Tick](/external-metrics.go#L26)

`func Tick(d time.Duration) <-chan context.Context`

Tick essentially wraps time.Tick, but sends a Notification instead of
a time.Time

## Types

### type [MetricFunc](/external-metrics.go#L102)

`type MetricFunc func(ctx context.Context, namespace, name string) (result Value, err error)`

MetricFunc is a function which is run on events to update the latest value
of a metric

### type [Server](/external-metrics.go#L109)

`type Server struct { ... }`

Server provides an HTTP server, which provides metrics
to horizontal pod autoscalers in the way which they expect.

This server handles things like retry logic, caching results,
and handling namespacing of metrics, should they need it

#### func (*Server) [AddMetric](/external-metrics.go#L134)

`func (s *Server) AddMetric(ctx context.Context, namespace, name string, c <-chan context.Context, f MetricFunc)`

AddMetric registers a new metric function against the server, along with
and runs it every in a go func every time something triggers `c`

`c` could be anything from a time.Tick, to something listening out for external
events

#### func (Server) [Serve](/external-metrics.go#L142)

`func (s Server) Serve(addr string) (err error)`

Serve serves metrics over http on the

### type [Value](/external-metrics.go#L42)

`type Value struct { ... }`

Value is the value of a metric, either as a basic value, or as a set of
selectors for more granular selections

#### func (*Value) [AddSelector](/external-metrics.go#L74)

`func (v *Value) AddSelector(label, value string, i int64)`

AddSelector allows a value to contain a more granular value for HPAs that use labels

Why might you want this? Because each value creates a new go routine, with tickers, and
notifiers, and so on, they can be quite expensive to run. If each metric requires a call
to an external API, too, then the whole thing becomes a lot.

Not to mention the annoying thing of exposing a million dimensions like "/namespace/default/foo_A",
"/namespace/default/foo_B", and "/namespace/default/foo_C" where they all do the same thing.

In this case, your MetricFunc merely needs to do:

```go
v := NewValue(aValue)
v.AddSelector("dimension", "A", aValue)
v.AddSelector("dimension", "B", bValue)
v.AddSelector("dimension", "C", cValue)
```

Which exposes the metric "/namespace/default/foo?labelSelector=dimension=A" (etc.)

## Sub Packages

* [example](./example)

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)

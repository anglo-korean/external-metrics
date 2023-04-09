# Kubernetes External Metrics Server

Provides an easily exstensible external kubernetes metrics server.

## Why?

The officialy supported method [here](https://github.com/kubernetes-sigs/custom-metrics-apiserver) is just f-ing inscruitable.

Not only that, but the documentation for what these servers need to do are awful. The best bet is a design doc on an old branch of some random kubernetes repo. The second bet is a medium article where someone describes how they discovered what one of these APIs needs to do, but not what that is.

This package is pretty simple; it has a New() function, an AddMetric(name, namespace, function, notifier) function, and a Serve() function.

No thrills, no tunables, bugger all excitement.
# externalmetrics

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/anglo-korean/external-metrics)
[![Go Report Card](https://goreportcard.com/badge/github.com/anglo-korean/external-metrics)](https://goreportcard.com/report/github.com/anglo-korean/external-metrics)

## Types

### type [MetricFunc](/external-metrics.go#L22)

`type MetricFunc func(ctx context.Context, namespace, name string) (result int64, err error)`

MetricFunc is a function which is run on events to update the latest value
of a metric

### type [Server](/external-metrics.go#L29)

`type Server struct { ... }`

Server provides an HTTP server, which provides metrics
to horizontal pod autoscalers in the way which they expect.

This server handles things like retry logic, caching results,
and handling namespacing of metrics, should they need it

#### func (*Server) [AddMetric](/external-metrics.go#L54)

`func (s *Server) AddMetric(ctx context.Context, namespace, name string, c chan any, f MetricFunc)`

AddMetric registers a new metric function against the server, along with
and runs it every in a go func every time something triggers `c`

`c` could be anything from a time.Tick, to something listening out for external
events

#### func (Server) [Serve](/external-metrics.go#L62)

`func (s Server) Serve(addr string) (err error)`

Serve serves metrics over http on the

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)

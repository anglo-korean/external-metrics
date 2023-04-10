// package externalmetrics provides a dirt simple, lightweight, easily exstensible external metrics
// api server for Kubernetes.
//
// Why?
//
// The officially supported method (https://github.com/kubernetes-sigs/custom-metrics-apiserver) is just f-ing inscruitable.
//
// Not only that, but the documentation for what these servers need to do are awful. The best bet is a design doc on an old branch of some random kubernetes repo. The second bet is a medium article where someone describes how they discovered what one of these APIs needs to do, but not what that is.
//
// This package is pretty simple; it has a New() function, an AddMetric() function, and a Serve() function.
//
// No thrills, no tunables, bugger all excitement.
//
// A complete form of the following example can be found in `./example`, but the nuts and bolts look like:
//
//	func main() {
//	    s := externalmetrics.New()
//	    s.AddMetric(context.Background(), externalmetrics.NamespaceAll, "my-metric", externalmetrics.Tick(time.Second), someIncrementingMetric)
//	    panic(s.Serve(listenAddr))
//	}
//
// The function `someIncrementingMetric` might look like:
//
//	func someIncrementingMetric(ctx context.Context, namespace, name string) (result externalmetrics.Value, err error) {
//	    v := externalmetrics.NewValue(counterA)
//	    v.AddSelector("some-key", "A", counterA)
//	    v.AddSelector("some-key", "B", counterB)
//	    v.AddSelector("some-key", "C", counterC)
//	    return v, nil
//	}
//
// The metrics exposed by this service can then be exposed in an HPA
package externalmetrics

# Kubernetes External Metrics Server

Provides an easily exstensible external kubernetes metrics server.

## Why?

The officialy supported method [here](https://github.com/kubernetes-sigs/custom-metrics-apiserver) is just f-ing inscruitable.

Not only that, but the documentation for what these servers need to do are awful. The best bet is a design doc on an old branch of some random kubernetes repo. The second bet is a medium article where someone describes how they discovered what one of these APIs needs to do, but not what that is.

This package is pretty simple; it has a New() function, an AddMetric(name, namespace, function, notifier) function, and a Serve() function.

No thrills, no tunables, bugger all excitement.

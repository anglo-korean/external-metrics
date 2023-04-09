package externalmetrics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/metrics/pkg/apis/external_metrics"
)

const (
	urlPrefix = "/apis/external.metrics.k8s.io/v1beta1/namespaces"
)

// MetricFunc is a function which is run on events to update the latest value
// of a metric
type MetricFunc func(ctx context.Context, namespace, name string) (result int64, err error)

// Server provides an HTTP server, which provides metrics
// to horizontal pod autoscalers in the way which they expect.
//
// This server handles things like retry logic, caching results,
// and handling namespacing of metrics, should they need it
type Server struct {
	// metrics is a map of `s.mapping[namespace][metricName]` pointing
	// to the latest value of that metric
	metrics map[string]map[string]int64
}

// New creates an empty Server
func New() Server {
	return Server{
		metrics: make(map[string]map[string]int64),
	}
}

func (s *Server) createNamespaceIfNotExist(ns string) {
	_, ok := s.metrics[ns]
	if !ok {
		s.metrics[ns] = make(map[string]int64)
	}
}

// AddMetric registers a new metric function against the server, along with
// and runs it every in a go func every time something triggers `c`
//
// `c` could be anything from a time.Tick, to something listening out for external
// events
func (s *Server) AddMetric(ctx context.Context, namespace, name string, c chan any, f MetricFunc) {
	s.createNamespaceIfNotExist(namespace)
	s.metrics[namespace][name] = 0

	go s.runMetric(ctx, namespace, name, f)
}

// Serve serves metrics over http on the
func (s Server) Serve(addr string) (err error) {
	http.HandleFunc(urlPrefix, s.handle)

	return http.ListenAndServe(addr, nil)
}

func (s *Server) runMetricLoop(ctx context.Context, namespace, name string, c chan any, f MetricFunc) {
	for {
		select {
		case <-c:
			s.runMetric(ctx, namespace, name, f)

		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) runMetric(ctx context.Context, namespace, name string, f MetricFunc) {
	v, err := f(ctx, namespace, name)
	if err != nil {
		log.Print(err)
	}

	s.metrics[namespace][name] = v
}

func (s Server) handle(w http.ResponseWriter, req *http.Request) {
	var (
		output any
		err    error
	)

	urlParts := strings.Split(strings.Replace(req.URL.Path, "/", "", -1), "/")

	switch len(urlParts) {
	case 0:

	case 2:
		output, err = s.getMetric(urlParts[0], urlParts[1])

	default:
		err = fmt.Errorf("unable to parse url %s", req.URL)
	}

	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	out, err := json.Marshal(output)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(200)
	w.Write(out)
}

func (s Server) getMetric(namespace, name string) (l k8s.ExternalMetricValueList, err error) {
	var ok bool
	if _, ok = s.metrics[namespace]; !ok {
		return k8s.ExternalMetricValueList{}, fmt.Errorf("Namespace %s either does not exist, or has no metrics stored against it", namespace)
	}

	if _, ok = s.metrics[namespace][name]; !ok {
		return k8s.ExternalMetricValueList{}, fmt.Errorf("Mwtric %s either does not exist under namespace %s", name, namespace)
	}

	value := s.metrics[namespace][name]

	return k8s.ExternalMetricValueList{
		Items: []k8s.ExternalMetricValue{
			{
				MetricName:   name,
				MetricLabels: nil,
				Timestamp:    metav1.Now(),
				Value:        *resource.NewQuantity(value, resource.DecimalSI),
			},
		},
	}, nil
}

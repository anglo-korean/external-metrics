package externalmetrics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/metrics/pkg/apis/external_metrics"
)

const (
	urlPrefix = "/apis/external.metrics.k8s.io/v1beta1/namespaces/"

	NamespaceAll     = "*"
	NamespaceDefault = "default"
)

// Tick essentially wraps time.Tick, but sends a Notification instead of
// a time.Time
func Tick(d time.Duration) <-chan context.Context {
	c := make(chan context.Context)

	go func(c chan context.Context) {
		for range time.Tick(d) {
			ctx, _ := context.WithTimeout(context.Background(), d/2)

			c <- ctx
		}
	}(c)

	return c
}

// Value is the value of a metric, either as a basic value, or as a set of
// selectors for more granular selections
type Value struct {
	base      int64
	selectors map[string]map[string]int64
}

// NewValue creates a simple Value, with an int64; this value is returned when
// an HPA makes a call without a specific selector, and so can be thought of as
// the base value
func NewValue(v int64) Value {
	return Value{
		base:      v,
		selectors: make(map[string]map[string]int64),
	}
}

// AddSelector allows a value to contain a more granular value for HPAs that use labels
//
// Why might you want this? Because each value creates a new go routine, with tickers, and
// notifiers, and so on, they can be quite expensive to run. If each metric requires a call
// to an external API, too, then the whole thing becomes a lot.
//
// Not to mention the annoying thing of exposing a million dimensions like "/namespace/default/foo_A",
// "/namespace/default/foo_B", and "/namespace/default/foo_C" where they all do the same thing.
//
// In this case, your MetricFunc merely needs to do:
//
//	v := NewValue(aValue)
//	v.AddSelector("dimension", "A", aValue)
//	v.AddSelector("dimension", "B", bValue)
//	v.AddSelector("dimension", "C", cValue)
//
// Which exposes the metric "/namespace/default/foo?labelSelector=dimension=A" (etc.)
func (v *Value) AddSelector(label, value string, i int64) {
	if _, ok := v.selectors[label]; !ok {
		v.selectors[label] = make(map[string]int64)
	}

	v.selectors[label][value] = i
}

func (v Value) quantity(label, value string) *resource.Quantity {
	if label == "" || value == "" {
		return resource.NewQuantity(v.base, resource.DecimalSI)
	}

	l, ok := v.selectors[label]
	if !ok {
		return resource.NewQuantity(v.base, resource.DecimalSI)
	}

	i, ok := l[value]
	if !ok {
		return resource.NewQuantity(v.base, resource.DecimalSI)
	}

	return resource.NewQuantity(i, resource.DecimalSI)
}

// MetricFunc is a function which is run on events to update the latest value
// of a metric
type MetricFunc func(ctx context.Context, namespace, name string) (result Value, err error)

// Server provides an HTTP server, which provides metrics
// to horizontal pod autoscalers in the way which they expect.
//
// This server handles things like retry logic, caching results,
// and handling namespacing of metrics, should they need it
type Server struct {
	// metrics is a map of `s.mapping[namespace][metricName]` pointing
	// to the latest value of that metric
	metrics map[string]map[string]Value
}

// New creates an empty Server
func New() Server {
	return Server{
		metrics: make(map[string]map[string]Value),
	}
}

func (s *Server) createNamespaceIfNotExist(ns string) {
	_, ok := s.metrics[ns]
	if !ok {
		s.metrics[ns] = make(map[string]Value)
	}
}

// AddMetric registers a new metric function against the server, along with
// and runs it every in a go func every time something triggers `c`
//
// `c` could be anything from a time.Tick, to something listening out for external
// events
func (s *Server) AddMetric(ctx context.Context, namespace, name string, c <-chan context.Context, f MetricFunc) {
	s.createNamespaceIfNotExist(namespace)
	s.metrics[namespace][name] = NewValue(0)

	go s.runMetricLoop(ctx, namespace, name, c, f)
}

// Serve serves metrics over http on the
func (s Server) Serve(addr string) (err error) {
	http.HandleFunc(urlPrefix, s.handle)

	return http.ListenAndServe(addr, nil)
}

func (s *Server) runMetricLoop(ctx context.Context, namespace, name string, c <-chan context.Context, f MetricFunc) {
	for {
		select {
		case in := <-c:
			s.runMetric(in, namespace, name, f)

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

		selectorLabel, selectorValue string
	)

	urlParts := strings.Split(strings.Replace(req.URL.Path, urlPrefix, "", -1), "/")
	if len(req.URL.Query()) > 0 && req.URL.Query()["labelSelector"] != nil {
		labelSelectorParts := strings.Split(req.URL.Query()["labelSelector"][0], "=")
		if len(labelSelectorParts) == 2 {
			selectorLabel = labelSelectorParts[0]
			selectorValue = labelSelectorParts[1]
		}
	}

	switch len(urlParts) {
	case 0:

	case 2:
		output, err = s.getMetric(urlParts[0], urlParts[1], selectorLabel, selectorValue)

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

func (s Server) getMetric(namespace, name, label, value string) (l k8s.ExternalMetricValueList, err error) {
	var ok bool
	if _, ok = s.metrics[namespace]; !ok {
		return k8s.ExternalMetricValueList{}, fmt.Errorf("Namespace %s either does not exist, or has no metrics stored against it", namespace)
	}

	if _, ok = s.metrics[namespace][name]; !ok {
		return k8s.ExternalMetricValueList{}, fmt.Errorf("Metric %s either does not exist under namespace %s", name, namespace)
	}

	var metricsLabels map[string]string
	if label != "" && value != "" {
		metricsLabels = map[string]string{
			label: value,
		}
	}

	v := s.metrics[namespace][name]

	return k8s.ExternalMetricValueList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ExternalMetricValueList",
			APIVersion: "external.metrics.k8s.io/v1beta1",
		},
		ListMeta: metav1.ListMeta{
			ResourceVersion: fmt.Sprintf("%d", time.Now().Unix()),
		},
		Items: []k8s.ExternalMetricValue{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ExternalMetricValue",
					APIVersion: "external.metrics.k8s.io/v1beta1",
				},
				MetricName:   name,
				MetricLabels: metricsLabels,
				Timestamp:    metav1.Now(),
				Value:        *v.quantity(label, value),
			},
		},
	}, nil
}

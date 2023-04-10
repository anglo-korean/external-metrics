package main

import (
	"context"
	"log"
	"time"

	"github.com/anglo-korean/external-metrics"
)

var (
	listenAddr = ":8000"

	counterA int64 = 1
	counterB int64 = 1
	counterC int64 = -1
)

func main() {
	s := externalmetrics.New()
	s.AddMetric(context.Background(), externalmetrics.NamespaceAll, "incrementable", externalmetrics.Tick(time.Second), someIncrementingMetric)

	panic(s.Serve(listenAddr))

}

func someIncrementingMetric(ctx context.Context, namespace, name string) (result externalmetrics.Value, err error) {
	log.Print("calling metric incrementer")

	counterA++
	counterB += 5
	counterC--

	v := externalmetrics.NewValue(counterA)

	v.AddSelector("some-key", "A", counterA)
	v.AddSelector("some-key", "B", counterB)
	v.AddSelector("some-key", "C", counterC)

	return v, nil
}

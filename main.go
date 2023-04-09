package main

import (
	"os"

	"github.com/anglo-korean/logger-go"
	//types "github.com/anglo-korean/protobuf/types/go"
)

var (
	KafkaBootstrapServers = os.Getenv("KAFKA_BROKER")
	Version               = os.Getenv("VERSION")

	log logger.Logger
)

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	log, err = logger.New(hostname, "{{ .Name }}", Version)
	if err != nil {
		panic(err)
	}
}

func main() {
	log.Status("Starting!")

	log.Status("Connecting to kafka")
	_, err := NewKafka(KafkaBootstrapServers)
	if err != nil {
		panic(err)
	}

	// do a thing
}

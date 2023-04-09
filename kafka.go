package main

import (
	"fmt"
	"os"

	"github.com/anglo-korean/protobuf/types/go"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
)

var (
	incomingTopic = "{{ .Name }}-incoming"
	outgoingTopic = "{{ .Name }}-outgoing"
)

type consumer interface {
	Events() chan kafka.Event
	SubscribeTopics([]string, kafka.RebalanceCb) error
}

type producer interface {
	Events() chan kafka.Event
	Produce(*kafka.Message, chan kafka.Event) error
}

type Kafka struct {
	consumer consumer
	producer producer
}

func NewKafka(bootstrapServers string) (k Kafka, err error) {
	hostname, err := os.Hostname()
	if err != nil {
		return
	}

	k.consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               bootstrapServers,
		"group.id":                        "{{ .Name }}",
		"auto.offset.reset":               "earliest",
		"session.timeout.ms":              30000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": false,
		"enable.partition.eof":            true,
		"client.id":                       hostname,
	})

	if err != nil {
		return
	}

	k.producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	})

	if err != nil {
		return
	}

	err = k.consumer.SubscribeTopics([]string{incomingTopic}, nil)

	return
}

func (k Kafka) ConsumerLoop(c chan *types.Article) (err error) {
	for ev := range k.consumer.Events() {
		switch ev := ev.(type) {
		case *kafka.Message:
			msg := ev.Value

			i := &types.Article{}
			err := proto.Unmarshal(msg, i)
			if err != nil {
				log.Status(fmt.Sprintf("Error: %+v", err))

				continue
			}

			c <- i

		case kafka.Error:
			return ev

		default:
			log.Status(fmt.Sprintf("Kafka Loop: %+v, %T", ev, ev))
		}
	}
	return
}

func (k *Kafka) Write(a *types.Article) (err error) {
	b, err := proto.Marshal(a)
	if err != nil {
		return
	}

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &outgoingTopic, Partition: kafka.PartitionAny},
		Value:          b,
		Headers:        []kafka.Header{},
	}

	return k.producer.Produce(message, nil)
}

func (k Kafka) FollowWriteLogs() {
	for e := range k.producer.Events() {
		log.Status(e.String())
	}
}

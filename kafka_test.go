package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/anglo-korean/protobuf/types/go"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
)

type dummyKafka struct {
	err     bool
	message []byte
}

func (d dummyKafka) SubscribeTopics([]string, kafka.RebalanceCb) error {
	if d.err {
		return fmt.Errorf("en error")
	}

	return nil
}

func (d dummyKafka) Produce(*kafka.Message, chan kafka.Event) error {
	if d.err {
		return fmt.Errorf("en error")
	}

	return nil
}

func (d dummyKafka) Events() chan kafka.Event {
	c := make(chan kafka.Event, 1)

	c <- &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &incomingTopic, Partition: kafka.PartitionAny},
		Value:          d.message,
		Headers:        []kafka.Header{},
	}

	return c
}

func TestNewKafka(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()

	_, err := NewKafka("example.com")
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}

func TestKafka_ConsumerLoop(t *testing.T) {
	data, _ := proto.Marshal(&types.Article{Content: &types.Text{Text: "Hello, world!"}})

	for _, test := range []struct {
		name        string
		msg         []byte
		expect      string
		expectError bool
	}{
		{"Empty input", []byte(""), "", true},
		{"Valid data", data, "Hello, world!", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			k := Kafka{consumer: dummyKafka{message: test.msg}}

			c := make(chan *types.Article)
			go func() {
				err := k.ConsumerLoop(c)
				if err == nil && test.expectError {
					t.Errorf("expected error")
				}

				if err != nil && !test.expectError {
					t.Errorf("unexpected error: %+v", err)
				}
			}()

			if !test.expectError {
				i := <-c

				if test.expect != i.Content.Text {
					t.Errorf("expected %q, received %q", test.expect, i.Content)
				}
			}
		})
	}
}

func TestKafka_Write(t *testing.T) {
	for _, test := range []struct {
		name        string
		client      dummyKafka
		expectError bool
	}{
		{"Happy path", dummyKafka{}, false},
		{"Erroring Kafka", dummyKafka{err: true}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			k := Kafka{producer: test.client}

			err := k.Write(&types.Article{})
			if err == nil && test.expectError {
				t.Errorf("expected error")
			}

			if err != nil && !test.expectError {
				t.Errorf("unexpected error: %+v", err)
			}
		})
	}
}

func TestKafka_FollowWriteLogs(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()

	go Kafka{producer: dummyKafka{}}.FollowWriteLogs()

	time.Sleep(100 * time.Millisecond)
}

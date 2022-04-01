package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"golang.org/x/sync/semaphore"
)

const TOPIC = "example-topic"

func printUsageAndExit() {
	fmt.Println("Usage:")
	fmt.Println("go run . producer")
	fmt.Println("  or")
	fmt.Println("go run . consumer <id>")
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 && len(os.Args) != 3 {
		printUsageAndExit()
	} else if os.Args[1] == "producer" {
		runProducer()
	} else if os.Args[1] == "consumer" && len(os.Args) == 3 {
		runConsumer(os.Args[2])
	} else {
		printUsageAndExit()
	}
}

func runConsumer(id string) {
	// A pool is used instead of a semaphore because there is no semaphore
	// in stdlib. This pool is used to know if a worker is available.
	// Channels don't provide any information about if a receiver is
	// currently listening at the other end, so need another sync
	// mechanism.
	//
	// EDIT: pool does not work properly.
	// pool := sync.Pool{}

	const N_WORKERS = 4
	sem := semaphore.NewWeighted(N_WORKERS)

	// set up workers
	workCh := make(chan string, 0)
	for i := 0; i < N_WORKERS; i++ {
		go consumerWorker(i, id, workCh, sem)
	}

	c := newConsumer()
	p, dCh := newProducer()

	for {
		msg := poll(c, 100)
		if msg == nil {
			continue
		}
		log.Printf("consumer %s got work: %s", id, string(msg.Value))

		// Check if a worker is available

		if !sem.TryAcquire(1) {
			// If no worker is available, sleep a bit then check
			// again. This check is not necessary for the example,
			// but something similar might be desirable in real
			// code to prevent constant message re-produces
			time.Sleep(100 * time.Millisecond)
			// if still not available, re-produce the work
			if !sem.TryAcquire(1) {
				log.Printf("consumer %s got unlucky, re-producing work: %s", id, string(msg.Value))
				produce(p, msg.Value, dCh)
				continue
			}
		}

		sem.Release(1)
		workCh <- string(msg.Value)
	}
}

func consumerWorker(wid int, cid string, workCh chan string, sem *semaphore.Weighted) {
	for work := range workCh {
		sem.Acquire(nil, 1)
		workDur, err := time.ParseDuration(work)
		if err != nil {
			log.Fatalf("could not parse work into duration: %v", workDur)
		}
		time.Sleep(workDur)
		sem.Release(1)
	}
}

func runProducer() {
	// run.sh will create 4 consumers which have 4 workers each. Produce
	// work that will on average be below load, but has randomness.
	p, dCh := newProducer()

	for {
		workMs := 1000 + rand.Intn(2000)
		work := fmt.Sprintf("%dms", workMs)
		produce(p, []byte(work), dCh)
		time.Sleep(200 * time.Millisecond)
	}
}

func newConsumer() *kafka.Consumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "example-group",
	})
	if err != nil {
		log.Fatalf("Could not create consumer: %v", err)
	}
	err = c.SubscribeTopics([]string{TOPIC}, nil)
	if err != nil {
		c.Close()
		log.Fatalf("Could not subscribe to topics: %v", err)
	}
	return c
}

func poll(c *kafka.Consumer, timeout int) *kafka.Message {
	ev := c.Poll(timeout)
	switch e := ev.(type) {
	case *kafka.Message:
		return e
	case kafka.Error:
		return nil
	case nil:
		return nil
	default:
		return nil
	}
}

func newProducer() (*kafka.Producer, chan kafka.Event) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		log.Fatalf("\n")
	}

	dCh := make(chan kafka.Event, 16)
	runDeliveryChecks(dCh)
	return p, dCh
}

func runDeliveryChecks(deliveryChan chan kafka.Event) {
	go func() {
		for range deliveryChan {
			e := <-deliveryChan
			if e == nil {
				continue
			}
			m := e.(*kafka.Message)

			if m.TopicPartition.Error != nil {
				log.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
			}
		}
	}()
}

func produce(p *kafka.Producer, value []byte, dCh chan kafka.Event) error {
	topic := TOPIC
	return p.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value: value,
		},
		dCh,
	)
}

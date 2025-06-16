package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func fail(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// 1) Connect & JetStream
	nc, err := nats.Connect(nats.DefaultURL)
	fail(err)
	defer nc.Drain()

	js, err := nc.JetStream()
	fail(err)

	// 2) Create stream FIVE with MaxMsgsPerSubject=5
	js.DeleteStream("FIVE")
	_, err = js.AddStream(&nats.StreamConfig{
		Name:               "FIVE",
		Subjects:           []string{"five.>"},
		Storage:            nats.FileStorage,
		MaxMsgsPerSubject:  5,
	})
	fail(err)

	// 3) Publish 6 messages on each of two subjects
	for _, subj := range []string{"five.1", "five.2"} {
		for i := 1; i <= 6; i++ {
			_, err := js.Publish(subj, []byte(fmt.Sprintf("msg-%d", i)))
			fail(err)
		}
	}

	// 4) Create a pull‐mode consumer C1 with LastPerSubject + Explicit ACK + MaxAckPending=1
	js.DeleteConsumer("FIVE", "C1")
	_, err = js.AddConsumer("FIVE", &nats.ConsumerConfig{
		Durable:        "C1",
		FilterSubject:  "five.>",
		DeliverPolicy:  nats.DeliverLastPerSubjectPolicy,
		AckPolicy:      nats.AckExplicitPolicy,
		AckWait:        30 * time.Second,
		MaxAckPending:  1,
		ReplayPolicy:   nats.ReplayInstantPolicy,
	})
	fail(err)

	sub, err := js.PullSubscribe(
		"five.>",
		"C1",
		nats.BindStream("FIVE"),
		nats.MaxAckPending(1),
		nats.DeliverLastPerSubject(),
	)
	fail(err)

	// 5) Pull one message at a time and ACK it
	for {
		// msgs, err := sub.Fetch(1, nats.Context(context.Background()))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		msgs, err := sub.Fetch(1, nats.Context(ctx))
		cancel()
		
		if err != nil {
			log.Printf("Fetch error: %v", err)
			break
		}
		if len(msgs) == 0 {
			log.Println("no more messages, exiting")
			break
		}
		m := msgs[0]
		log.Printf("Received %s, acking…", m.Subject)
		m.Ack()

		// short pause so logs are readable
		time.Sleep(500 * time.Millisecond)
	}
}

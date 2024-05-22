package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"

	kafka "github.com/segmentio/kafka-go"
)

type Consumer struct {
	ctx     context.Context
	reader  *kafka.Reader
	eventCh chan<- casino.Event
}

func NewConsumer(ctx context.Context, reader *kafka.Reader, eventCh chan<- casino.Event) *Consumer {
	return &Consumer{
		ctx:     ctx,
		reader:  reader,
		eventCh: eventCh,
	}
}

func (c *Consumer) Consume() {
	for {
		select {
		case <-c.ctx.Done():
			log.Println("Consumer context cancelled")
			return
		default:
			err := c.consumeMessage()
			if err != nil {
				log.Printf("error consuming message: %v", err)
			}
		}
	}
}

func (c *Consumer) consumeMessage() error {
	m, err := c.reader.ReadMessage(c.ctx)
	if err != nil {
		err := c.ctx.Err()
		if err != nil {
			return fmt.Errorf("context cancelled: %v", err)
		}

		return fmt.Errorf("read kafka message: %v", err)
	}

	var event casino.Event
	err = json.Unmarshal(m.Value, &event)
	if err != nil {
		return fmt.Errorf("unmarshal kafka message: %v", err)
	}

	c.eventCh <- event
	return nil
}

package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"

	kafka "github.com/segmentio/kafka-go"
)

type Publisher struct {
	ctx     context.Context
	writer  *kafka.Writer
	eventCh <-chan casino.Event
}

func NewPublisher(ctx context.Context, writer *kafka.Writer, eventCh <-chan casino.Event) *Publisher {
	return &Publisher{
		ctx:     ctx,
		writer:  writer,
		eventCh: eventCh,
	}
}

func (p *Publisher) Publish() {
	for event := range p.eventCh {
		err := p.publishEvent(event)
		if err != nil {
			log.Printf("error publishing event: %v", err)
		}
	}

	err := p.writer.Close()
	if err != nil {
		log.Printf("failed to close writer: %v", err)
	}
}

func (p *Publisher) publishEvent(event casino.Event) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling event: %v", err)
	}

	msg := kafka.Message{
		Key:   []byte("Key-" + strconv.Itoa(event.ID)),
		Value: jsonData,
	}

	err = p.writer.WriteMessages(p.ctx, msg)
	if err != nil {
		return fmt.Errorf("error writing message: %v", err)
	}

	return nil
}

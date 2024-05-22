package main

import (
	"log"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/config"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/generator"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/messaging"

	kafka "github.com/segmentio/kafka-go"
	"golang.org/x/net/context"
)

func main() {
	log.Println("Generator service starting...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := config.Initialize(
		config.WithKafkaBrokers(),
		config.WithOutputTopic(),
	)

	log.Println("Creating Kafka writer")
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.KafkaBrokers},
		Topic:   cfg.OutputTopic,
		Async:   true,
	})
	defer writer.Close()

	log.Println("Starting event generator")
	eventCh := generator.Generate(ctx)

	log.Println("Starting message publisher")
	publisher := messaging.NewPublisher(ctx, writer, eventCh)
	go publisher.Publish()

	for event := range eventCh {
		log.Printf("%#v\n", event)
	}

	log.Println("Finished")
}

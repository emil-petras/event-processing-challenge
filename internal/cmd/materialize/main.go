package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/config"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/messaging"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/process"

	"github.com/segmentio/kafka-go"
)

func main() {
	log.Println("Materialization service starting...")
	ctx := context.Background()

	cfg := config.Initialize(
		config.WithKafkaBrokers(),
		config.WithInputTopic(),
	)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBrokers},
		Topic:   cfg.InputTopic,
		GroupID: "materialization-service-group",
	})
	defer reader.Close()

	consumeCh := make(chan casino.Event)
	logCh := make(chan casino.Event)
	defer close(consumeCh)
	defer close(logCh)

	log.Println("Starting message consumer")
	consumer := messaging.NewConsumer(ctx, reader, consumeCh)
	go consumer.Consume()

	log.Println("Starting metrics processor")
	metrics := process.NewMetrics()
	go metrics.Process(consumeCh, logCh)

	log.Println("Starting logging")
	logger := process.NewLogger()
	go logger.Process(logCh)

	log.Println("Starting HTTP server")
	http.HandleFunc("/materialized", metrics.GetMetrics)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/config"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/messaging"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/process"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

func main() {
	log.Println("Player service starting...")
	ctx, cancel := context.WithCancel(context.Background())

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to handle OS signals
	go func() {
		<-signalChan
		log.Println("Received shutdown signal, stopping service...")
		cancel()
	}()

	cfg := config.Initialize(
		config.WithKafkaBrokers(),
		config.WithInputTopic(),
		config.WithOutputTopic(),
		config.WithDbConn(),
	)

	db, err := sql.Open("postgres", cfg.DbConnStr)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBrokers},
		GroupID: "player-enrichment-group",
		Topic:   cfg.InputTopic,
	})
	defer reader.Close()

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.KafkaBrokers},
		Topic:   cfg.OutputTopic,
	})
	defer writer.Close()

	consumeCh := make(chan casino.Event)
	publishCh := make(chan casino.Event)
	defer close(consumeCh)
	defer close(publishCh)

	log.Println("Starting message consumer")
	consumer := messaging.NewConsumer(ctx, reader, consumeCh)
	go consumer.Consume()

	log.Println("Starting description processor")
	player := process.NewPlayerData(ctx, db)
	go player.Process(consumeCh, publishCh)

	log.Println("Starting message publisher")
	publisher := messaging.NewPublisher(ctx, writer, publishCh)
	go publisher.Publish()

	// Wait for context cancellation (graceful shutdown)
	<-ctx.Done()
	time.Sleep(2 * time.Second)
	log.Println("Player service stopped")
}

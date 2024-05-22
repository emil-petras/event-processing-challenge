package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/config"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/messaging"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/process"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

func main() {
	log.Println("Currency service starting...")
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
		config.WithExchangeRateAPIURL(),
		config.WithExchangeRateCacheDuration(),
		config.WithExchangeRateAPIKey(),
		config.WithRedisAddr(),
		config.WithRedisPassword(),
		config.WithRedisDB(),
	)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBrokers},
		Topic:   cfg.InputTopic,
		GroupID: "currency-enricher-group",
	})
	defer reader.Close()

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.KafkaBrokers},
		Topic:   cfg.OutputTopic,
		Async:   true,
	})
	defer writer.Close()

	consumeCh := make(chan casino.Event)
	publishCh := make(chan casino.Event)
	defer close(consumeCh)
	defer close(publishCh)

	log.Println("Starting message consumer")
	consumer := messaging.NewConsumer(ctx, reader, consumeCh)
	go consumer.Consume()

	log.Println("Starting currency processor")
	converter := process.NewConverter(ctx, redisClient, httpClient)
	go converter.Process(cfg.Exchange(), consumeCh, publishCh)

	log.Println("Starting message publisher")
	publisher := messaging.NewPublisher(ctx, writer, publishCh)
	go publisher.Publish()

	// Wait for context cancellation (graceful shutdown)
	<-ctx.Done()
	time.Sleep(2 * time.Second)
	log.Println("Currency service stopped")
}

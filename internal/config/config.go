package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	KafkaBrokers              string
	InputTopic                string
	OutputTopic               string
	ExchangeRateAPIURL        string
	ExchangeRateCacheDuration int
	ExchangeRateAPIKey        string
	RedisAddr                 string
	RedisPassword             string
	RedisDB                   int
	DbConnStr                 string
}

type Option func(*Config)

func WithKafkaBrokers() Option {
	return func(cfg *Config) {
		cfg.KafkaBrokers = os.Getenv("KAFKA_BROKERS")
		if cfg.KafkaBrokers == "" {
			log.Fatal("KAFKA_BROKERS environment variable is not set")
		}
	}
}

func WithInputTopic() Option {
	return func(cfg *Config) {
		cfg.InputTopic = os.Getenv("KAFKA_INPUT_TOPIC")
		if cfg.InputTopic == "" {
			log.Fatal("KAFKA_INPUT_TOPIC environment variable is not set")
		}
	}
}

func WithOutputTopic() Option {
	return func(cfg *Config) {
		cfg.OutputTopic = os.Getenv("KAFKA_OUTPUT_TOPIC")
		if cfg.OutputTopic == "" {
			log.Fatal("KAFKA_OUTPUT_TOPIC environment variable is not set")
		}
	}
}

func WithExchangeRateAPIURL() Option {
	return func(cfg *Config) {
		cfg.ExchangeRateAPIURL = os.Getenv("EXCHANGE_RATE_API_URL")
		if cfg.ExchangeRateAPIURL == "" {
			log.Fatal("EXCHANGE_RATE_API_URL environment variable is not set")
		}
	}
}

func WithExchangeRateCacheDuration() Option {
	return func(cfg *Config) {
		cacheDurationStr := os.Getenv("EXCHANGE_RATE_CACHE_DURATION")
		if cacheDurationStr == "" {
			log.Fatal("EXCHANGE_RATE_CACHE_DURATION environment variable is not set")
		}
		cacheDuration, err := strconv.Atoi(cacheDurationStr)
		if err != nil {
			log.Fatal("Invalid value for EXCHANGE_RATE_CACHE_DURATION")
		}
		cfg.ExchangeRateCacheDuration = cacheDuration
	}
}

func WithExchangeRateAPIKey() Option {
	return func(cfg *Config) {
		cfg.ExchangeRateAPIKey = os.Getenv("EXCHANGE_RATE_API_KEY")
		if cfg.ExchangeRateAPIKey == "" {
			log.Fatal("EXCHANGE_RATE_API_KEY environment variable is not set")
		}
	}
}

func WithRedisAddr() Option {
	return func(cfg *Config) {
		cfg.RedisAddr = os.Getenv("REDIS_ADDR")
		if cfg.RedisAddr == "" {
			log.Fatal("REDIS_ADDR environment variable is not set")
		}
	}
}

func WithRedisPassword() Option {
	return func(cfg *Config) {
		cfg.RedisPassword = os.Getenv("REDIS_PASSWORD")
		if cfg.RedisPassword == "" {
			log.Fatal("REDIS_PASSWORD environment variable is not set")
		}
	}
}

func WithRedisDB() Option {
	return func(cfg *Config) {
		redisDBStr := os.Getenv("REDIS_DB")
		if redisDBStr == "" {
			log.Fatal("REDIS_DB environment variable is not set")
		}
		redisDB, err := strconv.Atoi(redisDBStr)
		if err != nil {
			log.Fatal("Invalid value for REDIS_DB")
		}
		cfg.RedisDB = redisDB
	}
}

func WithDbConn() Option {
	return func(cfg *Config) {
		DbConnStr := os.Getenv("PG_CONN_STR")
		if DbConnStr == "" {
			log.Fatal("PG_CONN_STR environment variable is not set")
		}
		cfg.DbConnStr = DbConnStr
	}
}

func Initialize(options ...Option) *Config {
	cfg := &Config{}
	for _, option := range options {
		option(cfg)
	}
	return cfg
}

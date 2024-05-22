package config

type Kafka struct {
	Brokers     string
	InputTopic  string
	OutputTopic string
}

type Exchange struct {
	APIURL        string
	CacheDuration int
	APIKey        string
}

type Redis struct {
	Addr     string
	Password string
	DB       int
}

func (cfg *Config) Kafka() Kafka {
	return Kafka{
		Brokers:     cfg.KafkaBrokers,
		InputTopic:  cfg.InputTopic,
		OutputTopic: cfg.OutputTopic,
	}
}

func (cfg *Config) Exchange() Exchange {
	return Exchange{
		APIURL:        cfg.ExchangeRateAPIURL,
		CacheDuration: cfg.ExchangeRateCacheDuration,
		APIKey:        cfg.ExchangeRateAPIKey,
	}
}

func (cfg *Config) Redis() Redis {
	return Redis{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
}

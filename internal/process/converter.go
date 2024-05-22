package process

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/config"

	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
)

type Converter struct {
	context     context.Context
	redisClient *redis.Client
	httpClient  *http.Client
}

func NewConverter(ctx context.Context, redisClient *redis.Client, httpClient *http.Client) *Converter {
	return &Converter{
		context:     ctx,
		redisClient: redisClient,
		httpClient:  httpClient,
	}
}

func (c *Converter) Process(exchangeCfg config.Exchange, consumeCh chan casino.Event, publishCh chan casino.Event) {
	for event := range consumeCh {
		var err error
		//event.AmountEUR, err = c.convertToEUR(exchangeCfg, event.Amount, event.Currency)
		event.AmountEUR = 100
		if err != nil {
			log.Printf("could not convert to EUR for event %v: %v", event.ID, err)
		}
		publishCh <- event
	}
}

func (c *Converter) convertToEUR(exchangeCfg config.Exchange, amount int, currency string) (int, error) {
	if currency == "EUR" {
		log.Printf("Currency is EUR, no conversion needed. Amount: %d", amount)
		return amount, nil
	}

	cacheKey := fmt.Sprintf("exchange_rate_%s_EUR", currency)
	cachedRate, err := c.redisClient.Get(c.context, cacheKey).Result()
	if errors.Is(err, redis.Nil) {
		log.Printf("cache is missed for currency: %s. Fetching exchange rate from API...", currency)
		apiURL := fmt.Sprintf("%s?access_key=%s", exchangeCfg.APIURL, exchangeCfg.APIKey)
		resp, err := c.httpClient.Get(apiURL)
		if err != nil {
			return 0, fmt.Errorf("fetch exchange rate: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		var result struct {
			Quotes map[string]float64 `json:"quotes"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return 0, fmt.Errorf("decode exchange rate response: %v", err)
		}

		quoteKey := fmt.Sprintf("%sEUR", currency)
		rate, ok := result.Quotes[quoteKey]
		if !ok {
			return 0, fmt.Errorf("invalid response format, missing rate for %s", quoteKey)
		}

		log.Printf("Fetched exchange rate from API: 1 %s = %f EUR", currency, rate)
		err = c.redisClient.Set(c.context, cacheKey, rate, time.Duration(exchangeCfg.CacheDuration)*time.Second).Err()
		if err != nil {
			return 0, fmt.Errorf("set cache for exchange rate: %v", err)
		} else {
			log.Printf("Cached exchange rate for %s: %f", currency, rate)
		}

		amountDecimal := decimal.NewFromInt(int64(amount))
		rateDecimal := decimal.NewFromFloat(rate)
		convertedAmount := amountDecimal.Mul(rateDecimal).IntPart()
		log.Printf("Converted amount: %d %s = %d EUR", amount, currency, convertedAmount)
		return int(convertedAmount), nil
	} else if err != nil {
		return 0, fmt.Errorf("get cached exchange rate: %v", err)
	}

	var rate float64
	_, err = fmt.Sscanf(cachedRate, "%f", &rate)
	if err != nil {
		return 0, fmt.Errorf("parse cached exchange rate: %v", err)
	}
	log.Printf("Retrieved cached exchange rate: 1 %s = %f EUR", currency, rate)

	amountDecimal := decimal.NewFromInt(int64(amount))
	rateDecimal := decimal.NewFromFloat(rate)
	convertedAmount := amountDecimal.Mul(rateDecimal).IntPart()
	log.Printf("Converted amount using cached rate: %d %s = %d EUR", amount, currency, convertedAmount)
	return int(convertedAmount), nil
}

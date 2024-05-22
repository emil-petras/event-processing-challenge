package generator

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

func Generate(ctx context.Context) <-chan casino.Event {
	eventCh := make(chan casino.Event)
	var id int64

	go func() {
		defer close(eventCh)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				eventID := atomic.AddInt64(&id, 1)
				eventCh <- generate(int(eventID))
			}

			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		}
	}()

	return eventCh
}

func generate(id int) casino.Event {
	amount, currency := randomAmountCurrency()

	return casino.Event{
		ID:        id,
		PlayerID:  10 + rand.Intn(10),
		GameID:    100 + rand.Intn(10),
		Type:      randomType(),
		Amount:    amount,
		Currency:  currency,
		HasWon:    randomHasWon(),
		CreatedAt: time.Now(),
	}
}

func randomType() string {
	return casino.EventTypes[rand.Intn(len(casino.EventTypes))]
}

func randomAmountCurrency() (amount int, currency string) {
	currency = casino.Currencies[rand.Intn(len(casino.Currencies))]

	switch currency {
	case "BTC":
		amount = rand.Intn(1e5)
	default:
		amount = rand.Intn(2000)
	}

	return
}

func randomHasWon() bool {
	return rand.Intn(100) < 5
}

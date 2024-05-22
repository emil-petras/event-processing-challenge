package process

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"

	"github.com/shopspring/decimal"
)

type Metrics struct {
	TotalEvents              int             `json:"events_total"`
	EventsPerMinute          decimal.Decimal `json:"events_per_minute"`
	EventsPerSecondMovingAvg decimal.Decimal `json:"events_per_second_moving_average"`
	TopPlayerBets            PlayerMetric    `json:"top_player_bets"`
	TopPlayerWins            PlayerMetric    `json:"top_player_wins"`
	TopPlayerDeposits        DepositMetric   `json:"top_player_deposits"`
	mu                       sync.RWMutex
	eventTimestamps          []int64
	playerBets               map[int]int
	playerWins               map[int]int
	playerDeposits           map[int]int
	playerGameStarts         map[int]int
	playerGameStops          map[int]int
}

func NewMetrics() *Metrics {
	return &Metrics{
		EventsPerMinute:          decimal.Zero,
		EventsPerSecondMovingAvg: decimal.Zero,
		playerBets:               make(map[int]int),
		playerWins:               make(map[int]int),
		playerDeposits:           make(map[int]int),
		playerGameStarts:         make(map[int]int),
		playerGameStops:          make(map[int]int),
	}
}

type PlayerMetric struct {
	ID    int `json:"id"`
	Count int `json:"count"`
}

type DepositMetric struct {
	PlayerMetric
	AmountEUR int `json:"amount_eur"`
}

func (m *Metrics) Process(consumeCh <-chan casino.Event, resultCh chan<- casino.Event) {
	for event := range consumeCh {
		m.mu.Lock()
		m.TotalEvents++

		// Insert timestamp in sorted order
		insertIdx := sort.Search(len(m.eventTimestamps), func(i int) bool { return m.eventTimestamps[i] > event.CreatedAt.Unix() })
		m.eventTimestamps = append(m.eventTimestamps[:insertIdx], append([]int64{event.CreatedAt.Unix()}, m.eventTimestamps[insertIdx:]...)...)

		// Clean up old timestamps using binary search
		oneMinuteAgo := time.Now().Add(-1 * time.Minute).Unix()
		oldIdx := sort.Search(len(m.eventTimestamps), func(i int) bool { return m.eventTimestamps[i] >= oneMinuteAgo })
		m.eventTimestamps = m.eventTimestamps[oldIdx:]

		// Update player metrics
		switch event.Type {
		case "bet":
			m.playerBets[event.PlayerID]++
			if event.HasWon {
				m.playerWins[event.PlayerID]++
			}
		case "deposit":
			m.playerDeposits[event.PlayerID] += event.AmountEUR
		case "game_start":
			m.playerGameStarts[event.PlayerID]++
		case "game_stop":
			m.playerGameStops[event.PlayerID]++
		}

		// Calculate events per minute
		durationCovered := time.Duration(len(m.eventTimestamps)) * time.Second
		if durationCovered > 0 {
			m.EventsPerMinute = decimal.NewFromInt(int64(len(m.eventTimestamps))).Div(decimal.NewFromFloat(durationCovered.Minutes()))
		}

		// Calculate events per second moving average in the last minute
		if len(m.eventTimestamps) > 0 {
			durationCovered := decimal.NewFromInt(time.Now().Unix() - m.eventTimestamps[0])
			m.EventsPerSecondMovingAvg = decimal.NewFromInt(int64(len(m.eventTimestamps))).Div(durationCovered)
		}

		// Determine top players
		m.TopPlayerBets = m.getMax(m.playerBets)
		m.TopPlayerWins = m.getMax(m.playerWins)
		m.TopPlayerDeposits = DepositMetric{
			PlayerMetric: m.getMax(m.playerDeposits),
			AmountEUR:    m.playerDeposits[m.getMax(m.playerDeposits).ID],
		}

		m.mu.Unlock()

		resultCh <- event
	}
}

func (m *Metrics) getMax(playerMetrics map[int]int) PlayerMetric {
	var topPlayerID int
	var maxCount int
	for playerID, count := range playerMetrics {
		if count > maxCount {
			topPlayerID = playerID
			maxCount = count
		}
	}
	return PlayerMetric{ID: topPlayerID, Count: maxCount}
}

func (m *Metrics) GetMetrics(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := m.ToJSON()
	if err != nil {
		log.Printf("Failed to encode metrics: %v", err)
		http.Error(w, "metrics error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (m *Metrics) ToJSON() ([]byte, error) {
	response := struct {
		TotalEvents              int             `json:"events_total"`
		EventsPerMinute          decimal.Decimal `json:"events_per_minute"`
		EventsPerSecondMovingAvg decimal.Decimal `json:"events_per_second_moving_average"`
		TopPlayerBets            PlayerMetric    `json:"top_player_bets"`
		TopPlayerWins            PlayerMetric    `json:"top_player_wins"`
		TopPlayerDeposits        struct {
			PlayerMetric
			AmountEUR decimal.Decimal `json:"amount_eur"`
		} `json:"top_player_deposits"`
	}{
		TotalEvents:              m.TotalEvents,
		EventsPerMinute:          m.EventsPerMinute,
		EventsPerSecondMovingAvg: m.EventsPerSecondMovingAvg,
		TopPlayerBets:            m.TopPlayerBets,
		TopPlayerWins:            m.TopPlayerWins,
		TopPlayerDeposits: struct {
			PlayerMetric
			AmountEUR decimal.Decimal `json:"amount_eur"`
		}{
			PlayerMetric: m.TopPlayerDeposits.PlayerMetric,
			AmountEUR:    decimal.NewFromInt(int64(m.TopPlayerDeposits.AmountEUR)).Div(decimal.NewFromInt(100)),
		},
	}

	return json.Marshal(&response)
}

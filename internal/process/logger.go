package process

import (
	"encoding/json"
	"log"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Process(logCh <-chan casino.Event) {
	for event := range logCh {
		logJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("error marshaling event to JSON: %v", err)
			continue
		}
		log.Println(string(logJSON))
	}
}

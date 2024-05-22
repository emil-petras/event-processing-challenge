package process

import (
	"context"
	"fmt"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"

	"github.com/shopspring/decimal"
)

const unknownGame = "Unknown Game"
const createdAtFormat = "January 2nd, 2006 at 15:04 UTC"

type Descriptor struct {
	ctx context.Context
}

func NewDescriptor(ctx context.Context) *Descriptor {
	return &Descriptor{
		ctx: ctx,
	}
}

func (d *Descriptor) Process(consumeCh <-chan casino.Event, publishCh chan<- casino.Event) {
	for event := range consumeCh {
		description := d.createDescription(event)
		event.Description = description

		publishCh <- event
	}
}

func (d *Descriptor) createDescription(event casino.Event) string {
	gameTitle := getGameTitle(event.GameID)
	createdAt := event.CreatedAt.Format(createdAtFormat)
	description := ""

	amount := decimal.NewFromInt(int64(event.Amount)).Div(decimal.NewFromInt(100))
	amountEUR := decimal.NewFromInt(int64(event.AmountEUR)).Div(decimal.NewFromInt(100))

	switch event.Type {
	case "game_start":
		description = fmt.Sprintf("Player #%d started playing a game \"%s\" on %s.", event.PlayerID, gameTitle, createdAt)
	case "bet":
		description = fmt.Sprintf("Player #%d (%s) placed a bet of %s %s (%s EUR) on a game \"%s\" on %s.",
			event.PlayerID, event.Player.Email, amount.StringFixed(2), event.Currency, amountEUR.StringFixed(2), gameTitle, createdAt)
		if event.HasWon {
			description += " The bet was won."
		} else {
			description += " The bet was lost."
		}
	case "deposit":
		description = fmt.Sprintf("Player #%d made a deposit of %s %s on %s.",
			event.PlayerID, amount.StringFixed(2), event.Currency, createdAt)
	case "game_stop":
		description = fmt.Sprintf("Player #%d stopped playing a game \"%s\" on %s.", event.PlayerID, gameTitle, createdAt)
	default:
		description = fmt.Sprintf("Event ID #%d of type %s occurred on %s.", event.ID, event.Type, createdAt)
	}

	return description
}

func getGameTitle(gameID int) string {
	game, ok := casino.Games[gameID]
	if !ok {
		return unknownGame
	}

	return game.Title
}

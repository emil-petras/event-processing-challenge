package process

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

type PlayerData struct {
	ctx context.Context
	db  *sql.DB
}

func NewPlayerData(ctx context.Context, db *sql.DB) *PlayerData {
	return &PlayerData{
		ctx: ctx,
		db:  db,
	}
}

func (p *PlayerData) Process(consumeCh chan casino.Event, publishCh chan casino.Event) {
	for event := range consumeCh {
		player, err := p.getPlayerData(event.PlayerID)
		if err != nil {
			log.Printf("error fetching player data: %v", err)
		} else {
			if player == nil || player.IsZero() {
				log.Printf("Player data missing ID: %v", event.PlayerID)
			} else {
				event.Player = *player
			}
		}

		publishCh <- event
	}
}

func (p *PlayerData) getPlayerData(playerID int) (*casino.Player, error) {
	query := `SELECT email, last_signed_in_at FROM players WHERE id = $1`
	row := p.db.QueryRowContext(p.ctx, query, playerID)

	var player casino.Player
	err := row.Scan(&player.Email, &player.LastSignedInAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("No player data found for player ID %d", playerID)
			return nil, nil
		}
		return nil, fmt.Errorf("scanning player data: %v", err)
	}

	return &player, nil
}

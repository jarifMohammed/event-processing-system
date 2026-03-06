package repository

import (
	"context"
	"strings"

	"event-processing-system/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	DB *pgxpool.Pool
}

func (r *EventRepository) Insert(event model.OrderEvent) error {
	_, err := r.DB.Exec(context.Background(),
		"INSERT INTO processed_events (event_id, user_id, amount) VALUES ($1, $2, $3)",
		event.EventID, event.UserID, event.Amount,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil
		}
		return err
	}

	return nil
}

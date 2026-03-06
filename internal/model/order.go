package model

import "time"

type OrderEvent struct {
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

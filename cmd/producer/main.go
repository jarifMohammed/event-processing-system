package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"event-processing-system/internal/model"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func main() {
	// Initialize structured logging with JSON handler
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "orders.created",
	})
	defer writer.Close()

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			slog.Warn("method not allowed", "method", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			UserID string  `json:"user_id"`
			Amount float64 `json:"amount"`
		}

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			slog.Error("invalid json request", "error", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if body.UserID == "" || body.Amount <= 0 {
			slog.Warn("invalid input data", "user_id", body.UserID, "amount", body.Amount)
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		event := model.OrderEvent{
			EventID:   uuid.New().String(),
			UserID:    body.UserID,
			Amount:    body.Amount,
			CreatedAt: time.Now(),
		}

		data, err := json.Marshal(event)
		if err != nil {
			slog.Error("failed to serialize event", "event_id", event.EventID, "error", err)
			http.Error(w, "Failed to serialize event", 500)
			return
		}

		err = writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(event.EventID),
				Value: data,
			},
		)

		if err != nil {
			slog.Error("failed to publish event to kafka", "event_id", event.EventID, "error", err)
			http.Error(w, "Failed to publish event", 500)
			return
		}

		slog.Info("event published", "event_id", event.EventID, "user_id", event.UserID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "success",
			"event_id": event.EventID,
		})
	})

	server := &http.Server{
		Addr: ":8081",
	}

	go func() {
		slog.Info("Producer API running", "addr", ":8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down producer...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
	}
	slog.Info("producer exited cleanly")
}

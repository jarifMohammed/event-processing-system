package main

import (
	"context"
	"encoding/json"
	"log"
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "orders.created",
	})
	defer writer.Close()

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			UserID string  `json:"user_id"`
			Amount float64 `json:"amount"`
		}

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if body.UserID == "" || body.Amount <= 0 {
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
			http.Error(w, "Failed to publish event", 500)
			return
		}

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
		log.Println("Producer running on :8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down producer...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)
	log.Println("Producer exited cleanly")
}

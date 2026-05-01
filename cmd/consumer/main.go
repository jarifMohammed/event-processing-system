package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"

	"event-processing-system/internal/processor"
	"event-processing-system/internal/repository"
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbpool, err := pgxpool.New(context.Background(),
		"postgres://postgres:pass@localhost:5433/events")
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	repo := &repository.EventRepository{DB: dbpool}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "orders.created",
		GroupID: "order-processor-group",
	})
	defer reader.Close()

	slog.Info("Consumer service started", "topic", "orders.created", "group", "order-processor-group")

	jobs := make(chan []byte, 100)

	// Start worker pool
	for i := 0; i < 5; i++ {
		go processor.Worker(repo, jobs, i)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		slog.Info("shutting down consumer...")
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			slog.Info("consumer stopped gracefully")
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				slog.Error("kafka read error", "error", err)
				continue
			}
			jobs <- msg.Value
		}
	}
}

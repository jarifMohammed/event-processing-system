package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"

	"event-processing-system/internal/processor"
	"event-processing-system/internal/repository"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dbpool, err := pgxpool.New(context.Background(),
		"postgres://postgres:pass@localhost:5433/events")
	if err != nil {
		log.Fatal(err)
	}
	defer dbpool.Close()

	repo := &repository.EventRepository{DB: dbpool}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "orders.created",
		GroupID: "order-processor-group",
	})
	defer reader.Close()

	log.Println("Consumer started...")

	jobs := make(chan []byte, 100)

	for i := 0; i < 5; i++ {
		go processor.Worker(repo, jobs, i)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down consumer...")
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer stopped gracefully")
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Kafka read error: %v", err)
				continue
			}
			jobs <- msg.Value
		}
	}
}

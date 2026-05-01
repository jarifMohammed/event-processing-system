package processor

import (
	"encoding/json"
	"log/slog"
	"math"
	"time"

	"event-processing-system/internal/model"
	"event-processing-system/internal/repository"
)

func Worker(repo *repository.EventRepository, jobs <-chan []byte, id int) {
	for job := range jobs {
		slog.Info("worker processing event", "worker_id", id)
		processWithRetry(repo, job)
	}
}

func processWithRetry(repo *repository.EventRepository, data []byte) {
	var event model.OrderEvent

	err := json.Unmarshal(data, &event)
	if err != nil {
		slog.Error("failed to unmarshal event", "error", err)
		return
	}

	for i := 0; i < 3; i++ {
		err = repo.Insert(event)
		if err == nil {
			slog.Info("event processed successfully", "event_id", event.EventID, "user_id", event.UserID)
			return
		}

		// Exponential Backoff: 2^i * 1 second (1s, 2s, 4s...)
		backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
		slog.Warn("retry processing event", 
			"retry_count", i+1, 
			"event_id", event.EventID, 
			"backoff_seconds", backoff.Seconds(),
			"error", err,
		)
		time.Sleep(backoff)
	}

	slog.Error("failed to process event after max retries", "event_id", event.EventID)
}

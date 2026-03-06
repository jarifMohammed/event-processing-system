package processor

import (
	"encoding/json"
	"log"
	"time"

	"event-processing-system/internal/model"
	"event-processing-system/internal/repository"
)

func Worker(repo *repository.EventRepository, jobs <-chan []byte, id int) {
	for job := range jobs {
		log.Printf("Worker %d processing event", id)
		processWithRetry(repo, job)
	}
}

func processWithRetry(repo *repository.EventRepository, data []byte) {
	var event model.OrderEvent

	err := json.Unmarshal(data, &event)
	if err != nil {
		log.Printf("JSON error: %v", err)
		return
	}

	for i := 0; i < 3; i++ {
		err = repo.Insert(event)
		if err == nil {
			log.Printf("Event %s processed", event.EventID)
			return
		}

		log.Printf("Retry %d for event %s", i+1, event.EventID)
		time.Sleep(2 * time.Second)
	}

	log.Printf("Failed permanently: %s", event.EventID)
}

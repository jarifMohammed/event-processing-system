# Event Processing System

Kafka-based event-driven processing system built with Go and PostgreSQL.

---

## Features

- Producer API publishes order events to Kafka
- Consumer group processes events asynchronously
- Worker pool for controlled concurrency
- Retry logic for fault tolerance
- Idempotent database inserts (unique event_id)
- Graceful shutdown support

---

## Architecture

```
Client  
↓  
Producer (Go API)  
↓  
Kafka Topic (orders.created)  
↓  
Consumer Group (Worker Pool)  
↓  
PostgreSQL
```

---

## Tech Stack

- Go
- Kafka
- PostgreSQL
- Docker
- pgxpool
- kafka-go

---

## Setup

### Start Infrastructure

```bash
docker compose up -d
```

### Create Topic

```bash
docker exec -it kafka kafka-topics \
  --create \
  --topic orders.created \
  --bootstrap-server localhost:9092 \
  --partitions 1 \
  --replication-factor 1
```

### Run Producer

```bash
go run cmd/producer/main.go
```

### Run Consumer

```bash
go run cmd/consumer/main.go
```

---

## API Endpoint

```
POST /orders
```

```json
{
  "user_id": "user1",
  "amount": 100
}
```

---

## Design Highlights

- Decoupled producer and consumer via Kafka
- Bounded worker pool prevents uncontrolled goroutines
- Retry mechanism ensures reliability
- Unique constraint ensures idempotency
- Graceful shutdown prevents message loss

---


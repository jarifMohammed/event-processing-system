# Event Processing System

Kafka-based event-driven processing system built with Go and PostgreSQL. Designed as a deep dive into **System Design patterns** for high-scale, resilient architectures.

---

## Features

- **Producer API**: Publishes order events to Kafka with structured logging.
- **Consumer Group**: Scalable event processing using Kafka Consumer Groups.
- **Bounded Worker Pool**: Controlled concurrency using fixed-size goroutine pools.
- **Exponential Backoff**: Resilient retry logic for database operations.
- **Idempotent Inserts**: Ensures data integrity using unique event constraints.
- **Structured Logging**: JSON-formatted logs for production observability.
- **Graceful Shutdown**: Prevents data loss during system termination.

---

## Architecture

```
Client  
↓  
Producer (Go API + JSON Logging)  
↓  
Kafka Topic (orders.created)  
↓  
Consumer Group (Worker Pool + Backoff)  
↓  
PostgreSQL (Idempotent Storage)
```

---

## System Design Patterns Implemented

- **Event-Driven Architecture (EDA)**: Decoupling services via Kafka to improve responsiveness and scalability.
- **Backpressure Management**: Using bounded channels and worker pools to prevent service exhaustion.
- **Exactly-Once Semantics (at Storage Layer)**: Implementing idempotency via Postgres unique constraints to handle "at-least-once" delivery.
- **Resiliency**: Using Exponential Backoff (1s, 2s, 4s...) to handle transient failures gracefully.
- **Observability**: Implementing Structured Logging (JSON) to enable centralized log analysis and monitoring.

---

## Tech Stack

- **Language**: Go (Golang)
- **Message Broker**: Apache Kafka
- **Database**: PostgreSQL
- **Infrastructure**: Docker & Docker Compose
- **Libraries**: `pgxpool`, `kafka-go`, `slog` (Standard Library)

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

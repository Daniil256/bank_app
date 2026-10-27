package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"bank_app/shared/roles"
)

// StatusUpdater — минимальный интерфейс, чтобы не тащить весь репозиторий
type StatusUpdater interface {
	UpdateStatus(ctx context.Context, customerID, status string) error
}

// Event — конверт Kafka-события
type Event struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}

// KycPayload — payload событий kyc.*
type KycPayload struct {
	ApplicationID string `json:"application_id"`
	CustomerID    string `json:"customer_id"`
	Reason        string `json:"reason,omitempty"`
}

// Consumer — заглушка консьюмера Kafka.
// В реальном проекте: kafka-go / segmentio/kafka-go reader,
// дедупликация по event_id (Redis/БД), retry, DLQ.
type Consumer struct {
	updater StatusUpdater
	brokers string
	topics  []string
}

func NewConsumer(brokers string, updater StatusUpdater) *Consumer {
	return &Consumer{
		updater: updater,
		brokers: brokers,
		topics:  []string{"kyc.events"},
	}
}

// Run — заглушка: не подключается к Kafka, только логирует.
// Реальная реализация должна читать сообщения из топика kyc.events
// и вызывать handle(event).
func (c *Consumer) Run(ctx context.Context) error {
	log.Printf("[kafka.consumer] stub started (brokers=%s, topics=%v)", c.brokers, c.topics)
	<-ctx.Done()
	log.Println("[kafka.consumer] stopped")
	return nil
}

// handle — обработка одного события (используется и в тестах)
func (c *Consumer) handle(ctx context.Context, e Event) error {
	var p KycPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		log.Printf("[kafka.consumer] bad payload event_id=%s: %v", e.EventID, err)
		return err
	}

	switch e.EventType {
	case "kyc.application_approved":
		log.Printf("[kafka.consumer] kyc.application_approved customer_id=%s", p.CustomerID)
		return c.updater.UpdateStatus(ctx, p.CustomerID, roles.StatusActive)

	case "kyc.application_rejected":
		log.Printf("[kafka.consumer] kyc.application_rejected customer_id=%s reason=%s", p.CustomerID, p.Reason)
		return c.updater.UpdateStatus(ctx, p.CustomerID, roles.StatusRejected)

	default:
		log.Printf("[kafka.consumer] ignore event_type=%s", e.EventType)
		return nil
	}
}

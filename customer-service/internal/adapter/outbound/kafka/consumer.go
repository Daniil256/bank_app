package kafka

import (
	"context"
	"customer-service/internal/domain"
	"encoding/json"
	"log"
	"time"

	"bank_app/shared/roles"

	"github.com/segmentio/kafka-go"
)

const (
	TopicAuthEvents = "auth.events"
	TopicKycEvents  = "kyc.events"
	GroupID         = "customer-service-group"
)

type StatusUpdater interface {
	UpdateStatus(ctx context.Context, userID, status string) error
}

// CustomerCreator — интерфейс для создания клиента (реализуется репозиторием)
type CustomerCreator interface {
	Create(ctx context.Context, customer *domain.Customer) error
}

type Event struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	OccurredAt string          `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}

type AuthPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

type KycPayload struct {
	ApplicationID string `json:"application_id"`
	UserID        string `json:"user_id"`
	Reason        string `json:"reason,omitempty"`
}

type Consumer struct {
	updater    StatusUpdater
	creator    CustomerCreator
	authReader *kafka.Reader
	kycReader  *kafka.Reader
}

func NewConsumer(broker string, updater StatusUpdater, creator CustomerCreator) *Consumer {
	// Reader для auth.events (создание клиентов)
	authReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    TopicAuthEvents,
		GroupID:  GroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  500 * time.Millisecond,
	})

	// Reader для kyc.events (обновление статуса)
	kycReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    TopicKycEvents,
		GroupID:  GroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
		MaxWait:  500 * time.Millisecond,
	})

	return &Consumer{
		updater:    updater,
		creator:    creator,
		authReader: authReader,
		kycReader:  kycReader,
	}
}

func (c *Consumer) Close() error {
	if err := c.authReader.Close(); err != nil {
		return err
	}
	return c.kycReader.Close()
}

// Run — запускает два горутин-цикла для чтения из двух топиков
func (c *Consumer) Run(ctx context.Context) error {
	go c.consumeAuthEvents(ctx)
	go c.consumeKycEvents(ctx)
	<-ctx.Done()
	return nil
}

func (c *Consumer) consumeAuthEvents(ctx context.Context) {
	log.Println("[kafka.consumer] auth.events consumer started")
	for {
		msg, err := c.authReader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // context cancelled
			}
			log.Printf("[kafka.consumer] auth.events fetch error: %v", err)
			continue
		}

		var event Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("[kafka.consumer] auth.events bad json: %v", err)
			c.commit(ctx, c.authReader, msg) // пропускаем битое сообщение
			continue
		}

		if event.EventType == "auth.user_registered" {
			if err := c.handleUserRegistered(ctx, event.Payload); err != nil {
				log.Printf("[kafka.consumer] auth.events handle error: %v", err)
				// не коммитим — попробуем снова
				continue
			}
		}

		if err := c.authReader.CommitMessages(ctx, msg); err != nil {
			log.Printf("[kafka.consumer] auth.events commit error: %v", err)
		}
	}
}

func (c *Consumer) consumeKycEvents(ctx context.Context) {
	log.Println("[kafka.consumer] kyc.events consumer started")
	for {
		msg, err := c.kycReader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[kafka.consumer] kyc.events fetch error: %v", err)
			continue
		}

		var event Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("[kafka.consumer] kyc.events bad json: %v", err)
			c.commit(ctx, c.kycReader, msg)
			continue
		}

		var p KycPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			log.Printf("[kafka.consumer] kyc.events bad payload: %v", err)
			c.commit(ctx, c.kycReader, msg)
			continue
		}

		switch event.EventType {
		case "kyc.application_approved":
			log.Printf("[kafka.consumer] kyc.application_approved customer_id=%s", p.UserID)
			if err := c.updater.UpdateStatus(ctx, p.UserID, roles.StatusActive); err != nil {
				log.Printf("[kafka.consumer] update status error: %v", err)
				continue // не коммитим
			}
		case "kyc.application_rejected":
			log.Printf("[kafka.consumer] kyc.application_rejected customer_id=%s", p.UserID)
			if err := c.updater.UpdateStatus(ctx, p.UserID, roles.StatusRejected); err != nil {
				log.Printf("[kafka.consumer] update status error: %v", err)
				continue
			}
		default:
			log.Printf("[kafka.consumer] ignore event_type=%s", event.EventType)
		}

		if err := c.kycReader.CommitMessages(ctx, msg); err != nil {
			log.Printf("[kafka.consumer] kyc.events commit error: %v", err)
		}
	}
}

func (c *Consumer) handleUserRegistered(ctx context.Context, payload json.RawMessage) error {
	var p AuthPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}

	// Создаём клиента со статусом NEW
	customer := &domain.Customer{
		UserID: p.UserID,
		Status: roles.StatusNew,
	}
	if err := c.creator.Create(ctx, customer); err != nil {
		// Если клиент уже существует — это ок (идемпотентность)
		log.Printf("[kafka.consumer] create customer (may already exist): %v", err)
	}
	log.Printf("[kafka.consumer] auth.user_registered handled: user_id=%s", p.UserID)
	return nil
}

func (c *Consumer) commit(ctx context.Context, r *kafka.Reader, msg kafka.Message) {
	log.Println("[kafka.commit] msg", msg)
	if err := r.CommitMessages(ctx, msg); err != nil {
		log.Printf("[kafka.commit] commit error: %v", err)
	}
}

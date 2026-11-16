package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

const TopicAuthEvents = "auth.events"

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker string) *Producer {
	// Writer — потокобезопасен, можно переиспользовать
	w := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        TopicAuthEvents,
		Balancer:     &kafka.LeastBytes{}, // равномерное распределение по партициям
		RequiredAcks: kafka.RequireOne,    // хотя бы лидер подтвердил запись
		Async:        false,               // синхронная отправка для простоты
	}
	return &Producer{writer: w}
}

// Close — закрывает writer при завершении приложения
func (p *Producer) Close() error {
	return p.writer.Close()
}

// PublishUserRegistered — отправляет событие в Kafka
func (p *Producer) PublishUserRegistered(userID, email string) error {
	event := map[string]interface{}{
		"event_id":    fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		"event_type":  "auth.user_registered",
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
		"payload": map[string]string{
			"user_id": userID,
			"email":   email,
		},
	}

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(userID), // ключ = user_id, чтобы события одного юзера шли в одну партицию
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	log.Printf("[kafka] auth.user_registered sent: user_id=%s", userID)
	return nil
}

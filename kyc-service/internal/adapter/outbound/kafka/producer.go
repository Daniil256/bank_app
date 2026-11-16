package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Producer struct {
	broker string
}

func NewProducer(broker string) *Producer {
	return &Producer{broker: broker}
}

// PublishApplicationSubmitted — отправляет событие в Kafka (заглушка)
func (p *Producer) PublishApplicationSubmitted(applicationID, userID string) error {
	event := map[string]interface{}{
		"event_id":    fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		"event_type":  "kyc.application_submitted",
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
		"payload": map[string]string{
			"application_id": applicationID,
			"user_id":        userID,
		},
	}
	log.Printf("[kafka] send kyc.application_submitted: application_id=%s", applicationID)
	_, _ = json.Marshal(event)
	return nil
}
func (p *Producer) PublishApplicationApproved(applicationID, userID string) error {
	event := map[string]interface{}{
		"event_id":    fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		"event_type":  "kyc.application_approved",
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
		"payload": map[string]string{
			"application_id": applicationID,
			"user_id":        userID,
		},
	}
	log.Printf("[kafka] send kyc.application_approved: application_id=%s customer_id=%s", applicationID, userID)
	_, _ = json.Marshal(event)
	return nil
}

func (p *Producer) PublishApplicationRejected(applicationID, customerID, reason string) error {
	event := map[string]interface{}{
		"event_id":    fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		"event_type":  "kyc.application_rejected",
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
		"payload": map[string]string{
			"application_id": applicationID,
			"customer_id":    customerID,
			"reason":         reason,
		},
	}
	log.Printf("[kafka] send kyc.application_rejected: application_id=%s reason=%s", applicationID, reason)
	_, _ = json.Marshal(event)
	return nil
}

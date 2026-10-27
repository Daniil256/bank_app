package kafka

import "log"

type Producer struct {
	brokers string
}

func NewProducer(brokers string) *Producer {
	return &Producer{brokers: brokers}
}

// PublishProfileFilled — заготовка. В реальном проекте — Kafka + outbox.
func (p *Producer) PublishProfileFilled(customerID, userID string) error {
	log.Printf("[kafka] customer.profile_filled: customer_id=%s user_id=%s", customerID, userID)
	return nil
}

package kafka

import "log"

type Producer struct {
	broker string
}

func NewProducer(broker string) *Producer {
	return &Producer{broker: broker}
}

// PublishUserRegistered — заглушка. В реальном проекте — запись в Kafka.
func (p *Producer) PublishUserRegistered(userID, email string) error {
	log.Printf("[kafka] auth.user_registered: user_id=%s email=%s", userID, email)
	return nil
}

package services

import (
	"github.com/Shopify/sarama"
)

type KafkaOperations interface {
	SendMessage()
}

// KafkaService is a ..
type KafkaService struct {
	ap sarama.AsyncProducer
}

type KafkaMessage struct {
	Name  string
	Value string
}

func NewKafkaService(ap sarama.AsyncProducer) *KafkaService {
	return &KafkaService{
		ap: ap,
	}
}

func (k *KafkaService) SendMessage() {

}

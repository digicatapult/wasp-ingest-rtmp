package services

import (
	"github.com/Shopify/sarama"
)

// Payload defines the data contained in
type Payload struct{}

type KafkaOperations interface {
	SendMessage()
	PayloadQueue() chan<- *Payload
}

// KafkaService is a ..
type KafkaService struct {
	ap sarama.AsyncProducer

	payloads chan *Payload
}

type KafkaMessage struct {
	Name  string
	Value string
}

func NewKafkaService(ap sarama.AsyncProducer) *KafkaService {
	return &KafkaService{
		ap: ap,

		payloads: make(chan *Payload),
	}
}

func (k *KafkaService) SendMessage() {
}

// PayloadQueue provides access to load a payload object into the queue for sending
func (k *KafkaService) PayloadQueue() chan<- *Payload {
	return k.payloads
}

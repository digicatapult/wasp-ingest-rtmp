package services

import (
	"github.com/Shopify/sarama"
	"log"
	"os"
	"time"
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

func (k *KafkaService) ProducerMessage(km KafkaMessage, signals chan os.Signal) {
	log.Println("Producer message km.Name", km.Name)
	log.Println("Producer message km.Value", km.Value)

	// Now, we set the Partition field of the ProducerMessage struct.
	msg := &sarama.ProducerMessage{
		Topic:     "raw-payloads",
		Partition: 1,
		Key:       sarama.StringEncoder(km.Name),
		Value:     sarama.StringEncoder(km.Value),
	}
	log.Println("Producer message", msg)

	var enqueued, errors int

ProducerLoop:
	for {
		time.Sleep(time.Second)

		select {
		case k.ap.Input() <- msg:
			enqueued++
			log.Println("New Message produced")
		case err := <-k.ap.Errors():
			log.Println("Failed to produce message", err)
			errors++
		case <-signals:
			break ProducerLoop
		}
	}

	log.Printf("Enqueued: %d; errors: %d\n", enqueued, errors)
}

package services

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/Shopify/sarama"

	"github.com/digicatapult/wasp-ingest-rtmp/util"
)

// Payload defines the data contained in
type Payload struct {
	ID      string
	FrameNo int
	Data    []byte
}

// KafkaOperations defines operations for kafka messaging
type KafkaOperations interface {
	SendMessage()
	PayloadQueue() chan<- *Payload
}

// KafkaService implements kafka message functionality
type KafkaService struct {
	ap sarama.AsyncProducer

	payloads chan *Payload
}

// KafkaMessage defines the message structure
type KafkaMessage struct {
	Ingest    string `json:"ingest"`
	IngestID  string `json:"ingestId"`
	Timestamp string `json:"timestamp"`
	Payload   string `json:"payload"`
	Metadata  string `json:"metadata"`
}

// NewKafkaService will instantiate an instance using the producer provided
func NewKafkaService(ap sarama.AsyncProducer) *KafkaService {
	return &KafkaService{
		ap: ap,

		payloads: make(chan *Payload),
	}
}

// SendMessage can send a message to the
func (k *KafkaService) SendMessage(mKey string, mValue KafkaMessage, signals chan os.Signal) {
	mValueMarshal := &mValue

	mValueMarshalled, errJSONMarshal := json.Marshal(mValueMarshal)
	if errJSONMarshal != nil {
		log.Fatalln(errJSONMarshal)

		return
	}

	msg := &sarama.ProducerMessage{
		Topic: util.GetEnv(util.KafkaTopicEnv, "raw-payloads"),
		Key:   sarama.StringEncoder(mKey),
		Value: sarama.StringEncoder(mValueMarshalled),
	}

	var enqueued, errors int

ProducerLoop:
	for {
		time.Sleep(time.Second)

		select {
		case k.ap.Input() <- msg:
			enqueued++
			log.Println("New Message produced")
		case err := <-k.ap.Errors():
			log.Fatalln("Failed to produce message", err)
			errors++
		case <-signals:
			break ProducerLoop
		}
	}

	log.Printf("Enqueued: %d; errors: %d\n", enqueued, errors)
}

// PayloadQueue provides access to load a payload object into the queue for sending
func (k *KafkaService) PayloadQueue() chan<- *Payload {
	return k.payloads
}

package services

import (
	"encoding/json"
	"log"

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
	SendMessage(mKey string, mValue KafkaMessage)
	PayloadQueue() chan *Payload
}

// KafkaService implements kafka message functionality
type KafkaService struct {
	sp sarama.SyncProducer

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
func NewKafkaService(sp sarama.SyncProducer) *KafkaService {
	return &KafkaService{
		sp: sp,

		payloads: make(chan *Payload),
	}
}

// SendMessage can send a message to the
func (k *KafkaService) SendMessage(mKey string, mValue KafkaMessage) {
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

	partition, offset, err := k.sp.SendMessage(msg)
	if err != nil {
		log.Printf("error sending msg %s - %s (%d, %d\n", msg.Key, err, partition, offset)
	}

	log.Printf("Message sent to partition %d, offset %d\n", partition, offset)
}

// PayloadQueue provides access to load a payload object into the queue for sending
func (k *KafkaService) PayloadQueue() chan *Payload {
	return k.payloads
}

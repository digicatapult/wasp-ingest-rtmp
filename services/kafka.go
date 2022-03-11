package services

import (
	"encoding/base64"
	"encoding/json"
	"sync"

	"github.com/Shopify/sarama"
	"go.uber.org/zap"

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
	StartBackgroundSend(*sync.WaitGroup, chan bool)
	PayloadQueue() chan<- *Payload
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
		zap.S().Fatal(errJSONMarshal)

		return
	}

	msg := &sarama.ProducerMessage{
		Topic: util.GetEnv(util.KafkaTopicEnv, "raw-payloads"),
		Key:   sarama.StringEncoder(mKey),
		Value: sarama.StringEncoder(mValueMarshalled),
	}

	partition, offset, err := k.sp.SendMessage(msg)
	if err != nil {
		zap.S().Infof("error sending msg %s - %s (%d, %d", msg.Key, err, partition, offset)
	}

	zap.S().Infof("Message sent to partition %d, offset %d", partition, offset)
}

// PayloadQueue provides access to load a payload object into the queue for sending
func (k *KafkaService) PayloadQueue() chan<- *Payload {
	return k.payloads
}

// StartBackgroundSend will start the background sender
func (k *KafkaService) StartBackgroundSend(sendWaitGroup *sync.WaitGroup, shutdown chan bool) {
	for {
		select {
		case payload := <-k.payloads:
			zap.S().Infof("Received video chunk: %d - %d", payload.FrameNo, len(payload.Data))

			messageKey := "01000000-0000-4000-8883-c7df300514ed"
			messageValue := KafkaMessage{
				Ingest:    "rtmp",
				IngestID:  "4883C7DF300514ED",
				Timestamp: "2021-08-31T14:51:36.507Z",
				Payload:   base64.StdEncoding.EncodeToString(payload.Data),
				Metadata:  "{}",
			}
			// zap.S().Infof("Encoded data: %s", messageValue.Payload)

			k.SendMessage(messageKey, messageValue)
			sendWaitGroup.Done()
		case <-shutdown:
			zap.S().Info("closing the background send")

			return
		}
	}
}

package services

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/Shopify/sarama"

	"github.com/digicatapult/wasp-ingest-rtmp/util"
)

type KafkaOperations interface {
	SendMessage()
	PayloadQueue() chan<- interface{}
}

type KafkaService struct {
	ap sarama.AsyncProducer
}

type KafkaMessage struct {
	Ingest    string `json:"ingest"`
	IngestID  string `json:"ingestId"`
	Timestamp string `json:"timestamp"`
	Payload   string `json:"payload"`
	Metadata  string `json:"metadata"`
}

func NewKafkaService(ap sarama.AsyncProducer) *KafkaService {
	return &KafkaService{
		ap: ap,
	}
}

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

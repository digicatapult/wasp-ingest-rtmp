package services

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/digicatapult/wasp-ingest-rtmp/util"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	KafkaTopicEnv     = "KAFKA_TOPIC"
	KafkaPartitionEnv = "KAFKA_PARTITION"
)

type KafkaOperations interface {
	SendMessage()
}

type KafkaService struct {
	ap sarama.AsyncProducer
}

type KafkaMessage struct {
	Ingest    string `json:"ingest"`
	IngestId  string `json:"ingestId"`
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
	mValueMarshalled, err := json.Marshal(mValueMarshal)
	if err != nil {
		fmt.Println(err)
		return
	}

	paritionStr := util.GetEnv(KafkaPartitionEnv, "1")
	partitionInt, err := strconv.ParseInt(paritionStr, 10, 32)
	if err != nil {
		fmt.Println(err)
		return
	}

	msg := &sarama.ProducerMessage{
		Topic:     util.GetEnv(KafkaTopicEnv, "raw-payloads"),
		Partition: int32(partitionInt),
		Key:       sarama.StringEncoder(mKey),
		Value:     sarama.StringEncoder(mValueMarshalled),
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
			log.Println("Failed to produce message", err)
			errors++
		case <-signals:
			break ProducerLoop
		}
	}

	log.Printf("Enqueued: %d; errors: %d\n", enqueued, errors)
}

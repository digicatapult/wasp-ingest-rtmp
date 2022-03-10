package main

import (
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"

	"github.com/digicatapult/wasp-ingest-rtmp/services"
	"github.com/digicatapult/wasp-ingest-rtmp/util"
)

func setupProducer(kafkaBrokers []string) (sarama.AsyncProducer, error) {
	var (
		producer sarama.AsyncProducer
		err      error
	)

	if producer, err = sarama.NewAsyncProducer(kafkaBrokers, nil); err != nil {
		return producer, errors.Wrap(err, "problem initiating producer")
	}

	return producer, nil
}

func main() {
	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)

	kafkaBrokers := util.GetEnv(util.KafkaBrokersEnv, "localhost:9092")

	producer, errProducer := setupProducer(strings.Split(kafkaBrokers, ","))
	if errProducer != nil {
		panic(errProducer)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	kafka := services.NewKafkaService(producer)

	messageKey := "01000000-0000-4000-8883-c7df300514ed"
	messageValue := services.KafkaMessage{
		Ingest:    "rtmp",
		IngestID:  "4883C7DF300514ED",
		Timestamp: "2021-08-31T14:51:36.507Z",
		Payload:   "",
		Metadata:  "{}",
	}

<<<<<<< HEAD
=======
	kafka.SendMessage(messageKey, messageValue, signals)

	// videoIngest := services.NewVideoIngestService()
>>>>>>> main
	// videoIngest.IngestVideo()
	kafka.SendMessage(messageKey, messageValue, signals)
}

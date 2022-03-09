package main

import (
	"github.com/Shopify/sarama"
	"github.com/digicatapult/wasp-ingest-rtmp/services"
	"log"
	"os"
	"os/signal"
	"strings"
)

const (
	KafkaBrokersEnv = "KAFKA_BROKERS"
)

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func setupProducer(kafkaBrokers []string) (sarama.AsyncProducer, error) {
	return sarama.NewAsyncProducer(kafkaBrokers, nil)
}

func main() {
	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
	kafkaBrokers := getEnv(KafkaBrokersEnv, "localhost:9092")

	producer, err := setupProducer(strings.Split(kafkaBrokers, ","))
	if err != nil {
		panic(err)
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
		IngestId:  "4883C7DF300514ED",
		Timestamp: "2021-08-31T14:51:36.507Z",
		Payload:   "",
		Metadata:  "{}",
	}

	kafka.SendMessage(messageKey, messageValue, signals)

	//videoIngest.IngestVideo()
}

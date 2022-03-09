package main

import (
	"log"
	"os"
	"strings"

	"github.com/Shopify/sarama"

	"github.com/digicatapult/wasp-ingest-rtmp/services"
)

const (
	kafkaEnvVar = "SOME_VARNAME"
)

func main() {
	kafkaBrokers := os.Getenv(kafkaEnvVar)

	producer, err := setupProducer(strings.Split(kafkaBrokers, ","))
	if err != nil {
		log.Fatal(err)
	}

	kafka := services.NewKafkaService(producer)
	videoIngest := services.NewVideoIngestService(kafka)

	kafka.SendMessage()
	videoIngest.IngestVideo()
}

// setupProducer will create a AsyncProducer and returns it.
func setupProducer(kafkaBrokers []string) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()

	return sarama.NewAsyncProducer(kafkaBrokers, config)
}

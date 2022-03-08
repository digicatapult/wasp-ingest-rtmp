package main

import (
	"github.com/Shopify/sarama"
	"github.com/digicatapult/wasp-ingest-rtmp/services"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
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

// setupProducer will create a AsyncProducer and returns it.
func setupProducer(kafkaBrokers []string) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Flush.Messages = 1
	//config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	//config.ChannelBufferSize = ka
	//config.Producer.RequiredAcks = sarama.WaitForAll
	//config.Producer.Return.Successes = true
	//config.Producer.Partitioner = sarama.NewManualPartitioner

	return sarama.NewAsyncProducer(kafkaBrokers, config)
}

func main() {
	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
	kafkaBrokers := getEnv(KafkaBrokersEnv, "localhost:9092")
	log.Println("kafkaBrokers", kafkaBrokers)

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
	//videoIngest := services.NewVideoIngestService(kafka)

	kafkaMessage := services.KafkaMessage{Name: "First", Value: "Message"}

	kafka.ProducerMessage(kafkaMessage, signals)
	//videoIngest.IngestVideo()
}

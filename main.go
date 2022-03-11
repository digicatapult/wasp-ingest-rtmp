package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"

	"github.com/digicatapult/wasp-ingest-rtmp/services"
	"github.com/digicatapult/wasp-ingest-rtmp/util"
)

func setupProducer(kafkaBrokers []string) (sarama.SyncProducer, error) {
	var (
		producer sarama.SyncProducer
		err      error
	)

	if producer, err = sarama.NewSyncProducer(kafkaBrokers, nil); err != nil {
		return producer, errors.Wrap(err, "problem initiating producer")
	}

	return producer, nil
}

var rtmpURL string

func main() {
	flag.StringVar(&rtmpURL, "rtmp", "default", "The url of the rtmp stream to ingest")
	flag.Parse()

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

	kafka := services.NewKafkaService(producer)

	videoIngest := services.NewVideoIngestService(kafka)
	videoIngest.IngestVideo()
}

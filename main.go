package main

import (
	"log"
	"os"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"go.uber.org/zap"

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

func main() {
	cfg := zap.NewDevelopmentConfig()
	if os.Getenv("ENV") == "production" {
		cfg = zap.NewProductionConfig()
		lvl, err := zap.ParseAtomicLevel(os.Getenv("LOG_LEVEL"))
		if err != nil {
			panic("invalid log level")
		}
		log.Printf("setting level: %s\n", lvl.String())
		cfg.Level = lvl
	}
	logger, err := cfg.Build()
	if err != nil {
		panic("error initialising the logger")
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	sarama.Logger = util.SaramaZapLogger{}

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

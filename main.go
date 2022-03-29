package main

import (
	"flag"
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
	var rtmpURL, outputDir string

	flag.StringVar(&rtmpURL, "rtmp", "rtmp://localhost:1935/stream/test", "The url of the rtmp stream to ingest")
	flag.StringVar(&outputDir, "outputDir", "/tmp/wasp-videos", "The directory to output videos to")
	flag.Parse()

	cfg := zap.NewDevelopmentConfig()
	if os.Getenv("ENV") == "production" {
		cfg = zap.NewProductionConfig()

		lvl, err := zap.ParseAtomicLevel(os.Getenv("LOG_LEVEL"))
		if err != nil {
			panic("invalid log level")
		}

		log.Printf("setting level: %s", lvl.String())

		cfg.Level = lvl
	}

	logger, err := cfg.Build()
	if err != nil {
		panic("error initializing the logger")
	}

	defer func() {
		err := logger.Sync()
		if err != nil {
			log.Printf("error whilst syncing zap: %s\n", err)
		}
	}()

	zap.ReplaceGlobals(logger)

	sarama.Logger = util.SaramaZapLogger{}

	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)

	kafkaBrokers := util.GetEnv(util.KafkaBrokersEnv, "localhost:9092")

	producer, errProducer := setupProducer(strings.Split(kafkaBrokers, ","))
	if errProducer != nil {
		panic(errProducer)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			zap.S().Fatal(err)
		}
	}()

	kafka := services.NewKafkaService(producer)

	err = util.CheckAndCreate(outputDir)
	if err != nil {
		zap.S().Fatalf("unable to create video directory '%s': %s", outputDir, err)
	}

	videoIngest := services.NewVideoIngestService(outputDir, kafka)
	videoIngest.IngestVideo(rtmpURL)
}

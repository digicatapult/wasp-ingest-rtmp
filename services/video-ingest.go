package services

import (
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"sync"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go.uber.org/zap"
)

// VideoIngestService is a video ingest service
type VideoIngestService struct {
	ks KafkaOperations
}

// NewVideoIngestService instantiates a new instance
func NewVideoIngestService(ks KafkaOperations) *VideoIngestService {
	return &VideoIngestService{
		ks: ks,
	}
}

// IngestVideo will ingest a video
func (vs *VideoIngestService) IngestVideo(rtmpURL string) {
	videoSendWaitGroup := &sync.WaitGroup{}

	shutdown := make(chan bool)

	go vs.ks.StartBackgroundSend(videoSendWaitGroup, shutdown)

	ingestID := getIngestIDFromURL(rtmpURL)
	if ingestID == "" {
		zap.S().Warn("ingestID is empty, not consuming video")

		return
	}

	go vs.consumeVideo(ingestID, videoSendWaitGroup)

	done := make(chan error)

	go func() {
		// segmented files that are also fragmented
		err := ffmpeg.Input(rtmpURL).
			Output("./bin/my_frag_bunny_%03d.mp4", ffmpeg.KwArgs{
				"f":                      "segment",
				"segment_format_options": "movflags=frag_keyframe+empty_moov+faststart",
			}).
			Run()
		if err != nil {
			zap.S().Fatalf("problem with ffmpeg: %v", err)
		}
		done <- err
	}()

	err := <-done
	zap.S().Infof("Done (waiting for completion of send): %s", err)
	videoSendWaitGroup.Wait()
	shutdown <- true
}

// payloadQueueSender send message via payload channel
func (vs *VideoIngestService) payloadQueueSender(ingestID string, filePath string, videoSendWaitGroup *sync.WaitGroup) {
	fileByteArray, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	zap.S().Debugf("Video file byte array length: %d", len(fileByteArray))

	payload := &Payload{
		ID:      ingestID,
		FrameNo: 0,
		Data:    fileByteArray,
	}

	zap.S().Debugf("Video chunk: %d - %d", payload.FrameNo, len(payload.Data))
	videoSendWaitGroup.Add(1)
	vs.ks.PayloadQueue() <- payload
}

// ignore the first file as this can be too large sometimes...
var previousVideoFilePath = "bin/my_frag_bunny_000.mp4"

func (vs *VideoIngestService) consumeVideo(ingestID string, videoSendWaitGroup *sync.WaitGroup) {
	zap.S().Debugf("consumeVideo called...")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
	MONITOR:
		for {
			select {
			case event := <-watcher.Events:
				switch event.Op {
				case fsnotify.Write:
					if event.Op&fsnotify.Write == fsnotify.Write {
						zap.S().Debugf("modified file: %s", event.Name)
						zap.S().Debugf("previousVideoFilePath: %s", previousVideoFilePath)

						if event.Name != "bin/.DS_Store" && event.Name != previousVideoFilePath {
							vs.payloadQueueSender(ingestID, event.Name, videoSendWaitGroup)
						}
					}

					continue MONITOR
				}
			case err := <-watcher.Errors:
				log.Println("Error:", err)
			}
		}
	}()

	err = watcher.Add("./bin")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Closing Monitor...")
	<-done
}

func getIngestIDFromURL(rtmpURL string) string {
	parsed, err := url.Parse(rtmpURL)
	if err != nil {
		zap.S().Errorf("unable to parse rtmp url to create ingest id: %s", err)

		return ""
	}

	ingestID := parsed.Path

	if strings.HasPrefix(ingestID, "/") {
		ingestID = strings.TrimLeft(ingestID, "/")
	}

	ingestID = strings.Replace(ingestID, "/", "-", -1)

	return ingestID
}

package services

import (
	"io"
	"io/ioutil"
	"net/url"
	"os"
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
	pipeReader, _ := io.Pipe()
	videoSendWaitGroup := &sync.WaitGroup{}

	shutdown := make(chan bool)

	go vs.ks.StartBackgroundSend(videoSendWaitGroup, shutdown)

	ingestID := getIngestIDFromURL(rtmpURL)
	if ingestID == "" {
		zap.S().Warn("ingestID is empty, not consuming video")

		return
	}

	go vs.consumeVideo(ingestID, pipeReader, videoSendWaitGroup)

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

// TODO file watcher needed on ./bin directory...
func (vs *VideoIngestService) consumeVideo(ingestID string, reader io.Reader, videoSendWaitGroup *sync.WaitGroup) {
	zap.S().Debugf("consumeVideo called...")

	frameCount := 0

	// ignore the first file as this can be too large sometimes...
	lastAccessedVideoFilename := "my_frag_bunny_000.mp4"
	lastAccessedVideoCount := 0

	items, _ := ioutil.ReadDir("./bin")
	fileCounter := 0

	for _, item := range items {
		zap.S().Debugf("file path: %s", item.Name())
		zap.S().Debugf("fileCounter: %d", fileCounter)
		zap.S().Debugf("!strings.HasPrefix %s", !strings.HasPrefix(item.Name(), "."))
		zap.S().Debugf("item.Name() != lastAccessedVideoFilename %s", item.Name() != lastAccessedVideoFilename)
		zap.S().Debugf("fileCounter != lastAccessedVideoCount %s", fileCounter != lastAccessedVideoCount)

		if !item.IsDir() {
			if !strings.HasPrefix(item.Name(), ".") && item.Name() != lastAccessedVideoFilename && fileCounter == lastAccessedVideoCount {
				lastAccessedVideoFilename = item.Name()
				lastAccessedVideoCount += 1

				zap.S().Debugf("lastAccessedVideoFilename: %s lastAccessedVideoCount: %d", lastAccessedVideoFilename, lastAccessedVideoCount)

				dat, err := os.ReadFile("./bin/" + item.Name())
				if err == io.EOF {
					err = nil
				}

				payload := &Payload{
					ID:      ingestID,
					FrameNo: frameCount, // * frameSize,
					Data:    dat,
				}

				zap.S().Debugf("Video chunk: %d - %d", payload.FrameNo, len(payload.Data))
				videoSendWaitGroup.Add(1)
				vs.ks.PayloadQueue() <- payload

				fileCounter++
			}
		}
	}
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

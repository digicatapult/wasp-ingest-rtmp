package services

import (
	"errors"
	"io"
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
	pipeReader, pipeWriter := io.Pipe()
	videoSendWaitGroup := &sync.WaitGroup{}

	shutdown := make(chan bool)

	go vs.ks.StartBackgroundSend(videoSendWaitGroup, shutdown)

	go vs.consumeVideo(pipeReader, videoSendWaitGroup)

	done := make(chan error)

	go func() {
		err := ffmpeg.Input(rtmpURL).
			Output("pipe:", ffmpeg.KwArgs{"f": "h264"}).
			WithOutput(pipeWriter).
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

func (vs *VideoIngestService) consumeVideo(reader io.Reader, videoSendWaitGroup *sync.WaitGroup) {
	frameSize := 1000
	frameCount := 0
	buf := make([]byte, frameSize)

	for {
		count, err := io.ReadFull(reader, buf)
		frameCount++

		switch {
		case count == 0 || errors.Is(err, io.EOF):
			zap.S().Debug("end of stream reached")

			return
		case count != frameSize:
			zap.S().Infof("end of stream reached, sending short chunk: %d, %s", count, err)
		case err != nil:
			zap.S().Errorf("read error: %d, %s", count, err)

			if count == 0 {
				return
			}
		}

		bufCopy := make([]byte, frameSize)
		copy(bufCopy, buf)

		payload := &Payload{
			ID:      "the-stream-identifier",
			FrameNo: frameCount * frameSize,
			Data:    bufCopy,
		}

		zap.S().Debugf("Video chunk: %d - %d", payload.FrameNo, len(payload.Data))
		videoSendWaitGroup.Add(1)
		vs.ks.PayloadQueue() <- payload
	}
}

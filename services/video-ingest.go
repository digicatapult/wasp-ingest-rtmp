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
func (vs *VideoIngestService) IngestVideo() {
	pipeReader, pipeWriter := io.Pipe()
	videoSendWaitGroup := &sync.WaitGroup{}

	shutdown := make(chan bool)

	go vs.ks.StartBackgroundSend(videoSendWaitGroup, shutdown)

	go func() {
		frameSize := 1000
		frameCount := 0
		buf := make([]byte, frameSize)

		for {
			count, err := io.ReadFull(pipeReader, buf)
			frameCount++

			switch {
			case count == 0 || errors.Is(err, io.EOF):
				zap.S().Info("nothing found")

				return
			case count != frameSize:
				zap.S().Infof("end of stream: %d, %s", count, err)
			case err != nil:
				zap.S().Infof("read error: %d, %s", count, err)
			}

			bufCopy := make([]byte, frameSize)
			copy(bufCopy, buf)

			payload := &Payload{
				ID:      "the-stream-identifier",
				FrameNo: frameCount * frameSize,
				Data:    bufCopy,
			}

			zap.S().Infof("Video chunk: %d - %d", payload.FrameNo, len(payload.Data))
			videoSendWaitGroup.Add(1)
			vs.ks.PayloadQueue() <- payload
		}
	}()

	done := make(chan error)

	go func() {
		err := ffmpeg.Input("rtmp://192.168.68.119:1935/live/rfBd56ti2SMtYvSgD5xAV0YU99zampta7Z7S575KLkIZ9PYk").
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

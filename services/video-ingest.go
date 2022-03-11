package services

import (
	"errors"
	"io"
	"log"
	"sync"

	ffmpeg "github.com/u2takey/ffmpeg-go"
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

	go func() {
		frameSize := 1000
		frameCount := 0
		buf := make([]byte, frameSize)

		for {
			count, err := io.ReadFull(pipeReader, buf)
			frameCount++

			switch {
			case count == 0 || errors.Is(err, io.EOF):
				log.Println("nothing found")

				return
			case count != frameSize:
				log.Printf("end of stream: %d, %s\n", count, err)
			case err != nil:
				log.Printf("read error: %d, %s\n", count, err)
			}

			bufCopy := make([]byte, frameSize)
			copy(bufCopy, buf)

			payload := &Payload{
				ID:      "the-stream-identifier",
				FrameNo: frameCount * frameSize,
				Data:    bufCopy,
			}

			log.Printf("Video chunk: %d - %d", payload.FrameNo, len(payload.Data))
			videoSendWaitGroup.Add(1)
			vs.ks.PayloadQueue() <- payload
		}
	}()

	done := make(chan error)

	go func() {
		err := ffmpeg.Input(rtmpURL).
			Output("pipe:", ffmpeg.KwArgs{"f": "h264"}).
			WithOutput(pipeWriter).
			Run()
		if err != nil {
			log.Fatalf("problem with ffmpeg: %v\n", err)
		}
		done <- err
	}()

	err := <-done
	log.Printf("Done (waiting for completion of send): %s\n", err)
	videoSendWaitGroup.Wait()
	shutdown <- true
}

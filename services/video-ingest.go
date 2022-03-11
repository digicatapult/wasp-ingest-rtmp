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
func (vs *VideoIngestService) IngestVideo() {
	pipeReader, pipeWriter := io.Pipe()
	wg := &sync.WaitGroup{}

	shutdown := make(chan bool)

	go vs.ks.StartBackgroundSend(wg, shutdown)

	go func() {
		frameSize := 1000
		frameCount := 0
		buf := make([]byte, frameSize)
		for {
			n, err := io.ReadFull(pipeReader, buf)
			frameCount++

			switch {
			case n == 0 || errors.Is(err, io.EOF):
				log.Println("nothing found")
				return
			case n != frameSize:
				log.Printf("end of stream: %d, %s\n", n, err)
			case err != nil:
				log.Printf("read error: %d, %s\n", n, err)
			}

			new := make([]byte, frameSize)
			copy(new, buf)

			payload := &Payload{
				ID:      "the-stream-identifier",
				FrameNo: frameCount * frameSize,
				Data:    new,
			}

			log.Printf("Video chunk: %d - %d", payload.FrameNo, len(payload.Data))
			wg.Add(1)
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
			log.Fatalf("problem with ffmpeg: %v\n", err)
		}
		done <- err
	}()

	err := <-done
	log.Printf("Done (waiting for completion of send): %s\n", err)
	wg.Wait()
	shutdown <- true
}

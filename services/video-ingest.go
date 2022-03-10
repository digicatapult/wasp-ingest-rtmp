package services

import (
	"errors"
	"io"
	"log"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// VideoIngestService is a video ingest service
type VideoIngestService struct{}

// NewVideoIngestService instantiates a new instance
func NewVideoIngestService() *VideoIngestService {
	return &VideoIngestService{}
}

// IngestVideo will ingest a video
func (vs *VideoIngestService) IngestVideo() {
	pipeReader, pipeWriter := io.Pipe()

	go func() {
		frameSize := 1000
		buf := make([]byte, frameSize)
		for {
			n, err := io.ReadFull(pipeReader, buf)
			switch {
			case n == 0 || errors.Is(err, io.EOF):
				log.Println("nothing found")
				return
			case n != frameSize:
				log.Printf("end of stream: %d, %s\n", n, err)
			case err != nil:
				log.Printf("read error: %d, %s\n", n, err)
			}

			payload := struct {
				ID   string
				Data []byte
			}{
				ID:   "the-stream-identifier",
				Data: buf,
			}
			log.Println(payload)
			// vs.ks.PayloadQueue() <- &payload
		}
	}()

	done := make(chan error)

	go func() {
		err := ffmpeg.Input("rtmp://192.168.68.119:1935/live/rfBd56ti2SMtYvSgD5xAV0YU99zampta7Z7S575KLkIZ9PYk").
			Output("pipe:", ffmpeg.KwArgs{"f": "rawvideo"}).
			WithOutput(pipeWriter).
			Run()
		if err != nil {
			log.Fatalf("problem with ffmpeg: %v\n", err)
			done <- err
		}
	}()

	err := <-done
	log.Printf("Done: %s\n", err)
}

package services

import (
	"errors"
	"io"
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
		// for i := 1; i <= 5; i++ {
		// 	filename := "video" + strconv.Itoa(i) + ".mp4"
		// 	ffmpeg.Input(rtmpURL).
		// 		Output(filename, ffmpeg.KwArgs{"f": "h264", "t": 1}).
		// 		Run()
		// }
		//codecs="avc1.64001F, mp4a.40.2
		//Output("out%03d.nut", ffmpeg.KwArgs{"f": "segment", "segment_time": 10}). //, "acodec": "mp4a"}). // "g": 60, "flags": "+cgop", "map": 0}).

		//err := ffmpeg.Input(rtmpURL).
		//	Output("output_01.mp4", ffmpeg.KwArgs{"f": "h264", "c": "copy"}).
		//	OverWriteOutput().ErrorToStdOut().
		//	Run()
		//if err != nil {
		//	zap.S().Fatalf("problem with ffmpeg: %v", err)
		//}
		//done <- err

		err := ffmpeg.Input(rtmpURL).
			Output("output%03d.ts",
				ffmpeg.KwArgs{
					"c:v":               "libx264",
					"an":                "",
					"f":                 "segment",
					"segment_time":      "10",
					"segment_format":    "mpegts",
					"segment_list_type": "m3u8",
					"segment_list_size": 3,
					"profile:v":         "baseline",
					"movflags":          "frag_keyframe+empty_moov",
					"g":                 52,
				}).
			OverWriteOutput().ErrorToStdOut().
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

func (vs *VideoIngestService) consumeVideo(ingestID string, reader io.Reader, videoSendWaitGroup *sync.WaitGroup) {
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
			ID:      ingestID,
			FrameNo: frameCount * frameSize,
			Data:    bufCopy,
		}

		zap.S().Debugf("Video chunk: %d - %d", payload.FrameNo, len(payload.Data))
		videoSendWaitGroup.Add(1)
		vs.ks.PayloadQueue() <- payload
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

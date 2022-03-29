package services

import (
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go.uber.org/zap"

	"github.com/digicatapult/wasp-ingest-rtmp/util"
)

// VideoIngestService is a video ingest service
type VideoIngestService struct {
	outputDir string

	ks KafkaOperations
}

// NewVideoIngestService instantiates a new instance
func NewVideoIngestService(outputDir string, ks KafkaOperations) *VideoIngestService {
	return &VideoIngestService{
		outputDir: outputDir,

		ks: ks,
	}
}

// IngestVideo will ingest a video
func (vs *VideoIngestService) IngestVideo(rtmpURL string) {
	pipeReader, pipeWriter := io.Pipe()
	videoSendWaitGroup := &sync.WaitGroup{}

	shutdown := make(chan bool)

	go vs.ks.StartBackgroundSend(videoSendWaitGroup, shutdown)

	ingestID := getIngestIDFromURL(rtmpURL)
	if ingestID == "" {
		zap.S().Warn("ingestID is empty, not consuming video")

		return
	}

	videoOutputDir := filepath.Join(vs.outputDir, ingestID)
	err := util.CheckAndCreate(videoOutputDir)
	if err != nil {
		zap.S().Fatalf("unable to create temp video directory '%s': %s", videoOutputDir, err)
	}

	go vs.consumeVideo(ingestID, videoOutputDir, pipeReader, videoSendWaitGroup)

	done := make(chan error)

	go func() {
		videoOutputPath := filepath.Join(videoOutputDir, `output%03d.ts`)

		err := ffmpeg.Input(rtmpURL).
			Output(videoOutputPath, ffmpeg.KwArgs{
				"f":                 "segment",
				"c:v":               "libx264",
				"an":                "",
				"segment_time":      1,
				"segment_list":      "pipe:1",
				"segment_format":    "mpegts",
				"segment_list_type": "m3u8",
				"segment_list_size": 3,
			}).
			WithOutput(pipeWriter).
			Run()
		if err != nil {
			zap.S().Fatalf("problem with ffmpeg: %v", err)
		}
		done <- err
	}()

	err = <-done
	zap.S().Infof("Done (waiting for completion of send): %s", err)
	videoSendWaitGroup.Wait()
	shutdown <- true
}

func (vs *VideoIngestService) consumeVideo(ingestID, videoOutputDir string, reader io.Reader, videoSendWaitGroup *sync.WaitGroup) {
	frameSize := 1000
	frameCount := 0
	buf := make([]byte, frameSize)

	for {
		_, err := reader.Read(buf)
		// frameCount++
		if err != nil {
			zap.S().Fatalf("read error: %s", err)
		}
		// switch {
		// case count == 0 || errors.Is(err, io.EOF):
		// 	zap.S().Debug("end of stream reached")

		// 	return
		// case count != frameSize:
		// 	zap.S().Infof("end of stream reached, sending short chunk: %d, %s", count, err)
		// case err != nil:
		// 	zap.S().Errorf("read error: %d, %s", count, err)

		// 	if count == 0 {
		// 		return
		// 	}
		// }

		bufCopy := make([]byte, frameSize)
		copy(bufCopy, buf)

		m3u8 := strings.TrimRight(string(bufCopy), "\x00")

		lines := strings.Split(m3u8, "\n")

		lastFile := lines[len(lines)-2]

		lastFilePath := filepath.Join(videoOutputDir, lastFile)

		zap.S().Infof("Path:'%s'", lastFilePath)

		f, err := os.Open(lastFilePath)
		if err != nil {
			zap.S().Fatalf("unable to open file for reading %s: %s", lastFilePath, err)
		}

		bytes, err := ioutil.ReadAll(f)
		if err != nil {
			zap.S().Fatalf("unable to open file for reading %s: %s", lastFilePath, err)
		}

		metaPayload := &Payload{
			ID:      ingestID,
			Type:    "meta",
			FrameNo: frameCount * frameSize,
			Data:    bufCopy,
		}
		dataPayload := &Payload{
			ID:      ingestID,
			Type:    "data",
			FrameNo: frameCount * frameSize,
			Data:    bytes,
		}
		videoSendWaitGroup.Add(1)
		vs.ks.PayloadQueue() <- metaPayload
		videoSendWaitGroup.Add(1)
		vs.ks.PayloadQueue() <- dataPayload
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

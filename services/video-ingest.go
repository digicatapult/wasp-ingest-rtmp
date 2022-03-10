package services

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"log"
)

// VideoIngestService is a video ingest service
type VideoIngestService struct {
}

// NewVideoIngestService instantiates a new instance
func NewVideoIngestService() *VideoIngestService {
	return &VideoIngestService{}
}

func (vs *VideoIngestService) IngestVideo() {
	err := ffmpeg.Input("rtmp://192.168.132.13:1935/live/rfBd56ti2SMtYvSgD5xAV0YU99zampta7Z7S575KLkIZ9PYk").
		// segmented files
		Output("myoutput_%03d.flv", ffmpeg.KwArgs{"f": "segment", "segment_time": "1", "segment_list": "out.list"}).
		// piping
		//Output("pipe:", ffmpeg.KwArgs{"f": "rawvideo"}).
		Run()

	if err != nil {
		panic(err)
	}

	log.Println("Done")
}

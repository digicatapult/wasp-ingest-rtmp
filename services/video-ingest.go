package services

import (
	"log"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// VideoIngestService is a video ingest service
type VideoIngestService struct {
}

// NewVideoIngestService instantiates a new instance
func NewVideoIngestService() *VideoIngestService {
	return &VideoIngestService{}
}

// IngestVideo ...
func (vs *VideoIngestService) IngestVideo() {
	err := ffmpeg.Input("rtmp://172.16.80.40:1935/live/rfBd56ti2SMtYvSgD5xAV0YU99zampta7Z7S575KLkIZ9PYk").

		// Output("output%03d.flv", ffmpeg.KwArgs{ "ssegment": "00:00:10","c:v": "copy",}).
		Filter(
			"segment",
			ffmpeg.Args{
				"segment_time=60",
				"segment_format_options=movflags=+faststart"},
			ffmpeg.KwArgs{
				"reset_timestamps": 1,
				"filename" : "output_%03d.flv",
			},
		).
		Output("output_%03d.flv" )
		// OverWriteOutput().ErrorToStdOut().Run()

		// This works and purely creates an FLV copy of the the stream
		// Output("output.flv", ffmpeg.KwArgs{ "t" : 10,}).
		// OverWriteOutput().ErrorToStdOut().Run()

		// Chunk to 1 seconds
		//Output("output.flv", ffmpeg.KwArgs{ "t" : 1,}).

		//-c:v copy -c:a copy
		//asegment=timestamps="60|150"
		// Output("output%03d.flv", ffmpeg.KwArgs{"c:v": "copy"}).
		// OverWriteOutput().ErrorToStdOut().Run()

	//-c:v copy -c:a copy
	//asegment=timestamps="60|150"

	//while loop when stream is playing?

	// split 10 second video using the timestamp from the previous loop?

	// Generate new filename

	// Output file

	// Record timestamp

	// input := ffmpeg.Input("rtmp://172.16.80.40:1935/live/rfBd56ti2SMtYvSgD5xAV0YU99zampta7Z7S575KLkIZ9PYk").Split()
	// sectionNumber := 0
	// input.Get(fmt.Sprintf("%v",sectionNumber)).Filter("segment", ffmpeg.Args{"segment: 00:00:10"}).
	// Output(fmt.Sprintf("%v",sectionNumber)+".flv")
	// sectionNumber++
	// OverWriteOutput().ErrorToStdOut().Run()

	/* input := ffmpeg.Input("./sample_data/in1.mp4").Split()
	out1 := input.Get("0").Filter("scale", ffmpeg.Args{"1920:-1"}).
	Output("./sample_data/1920.mp4", ffmpeg.KwArgs{"b:v": "5000k"})
	out2 := input.Get("1").Filter("scale", ffmpeg.Args{"1280:-1"}).
	Output("./sample_data/1280.mp4", ffmpeg.KwArgs{"b:v": "2800k"}) */

	if err != nil {
		log.Printf("Error processing the stream: %v\n", err)
	}
}

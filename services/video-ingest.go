package services

// VideoIngestService is a video ingest service
type VideoIngestService struct {
}

// NewVideoIngestService instantiates a new instance
func NewVideoIngestService(kafka KafkaOperations) *VideoIngestService {
	return &VideoIngestService{}
}

// IngestVideo ...
func (vs *VideoIngestService) IngestVideo() {
}

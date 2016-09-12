package wttypes

const (
	TRANSCODING_QUEUED    = "queued"
	TRANSCODING_REQUESTED = "requested"
	TRANSCODING_RUNNING   = "running"
	TRANSCODING_CANCELLED = "cancelled"
	TRANSCODING_FINISHED  = "finished"
	TRANSCODING_ERROR     = "error"
	TRANSCODING_SKIPPED   = "skipped"
)

// TranscodingTask is a struct with information regarding the transcoding
type TranscodingTask struct {
	ID         string `json:"id"`
	Profile    string `json:"profile,omitempty"`
	ObjectName string `json:"object_name,omitempty"`
	Status     string `json:"status,omitempty"`
}

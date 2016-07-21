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

// TranscodingProfile is a struct with information regarding the transcoding
type TranscodingProfile struct {
	ID         string `json:"id"`
	Profile    string `json:"profile"`
	ObjectName string `json:"object_name"`
	Status     string `json:"status"`
}

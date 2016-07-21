package wttypes

const (
	JOB_QUEUED    = "queued"
	JOB_RUNNING   = "running"
	JOB_CANCELLED = "cancelled"
	JOB_FINISHED  = "finished"
	JOB_ERROR     = "error"
)

// Job is a struct that stores all needed information for the job
type Job struct {
	ID           string               `json:"id"`
	URLMedia     string               `json:"url_media"`
	VideoName    string               `json:"video_name"`
	ObjectName   string               `json:"object_name"`
	Transcodings []TranscodingProfile `json:"transcodings"`
	Status       string               `json:"status"`
}

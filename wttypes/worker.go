package wttypes

const (
	WORKER_STATUS_ONLINE  = "online"
	WORKER_STATUS_IDLE    = "idle"
	WORKER_STATUS_BUSY    = "busy"
	WORKER_STATUS_OFFLINE = "offline"
)

type WorkerStatus struct {
	Addr   string `json:"addr,omitempty"`
	Status string `json:"status,omitempty"`
}

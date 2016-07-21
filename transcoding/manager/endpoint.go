package manager

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// AddTranscoding

type addTranscodingRequest struct {
	ID         string
	ObjectName string
	Profile    string
}

type addTranscodingResponse struct {
	Err error `json:"error,omitempty"`
}

func (r addTranscodingResponse) error() error { return r.Err }

func makeAddTranscodingEndpoint(ts Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addTranscodingRequest)
		err := ts.AddTranscoding(req.ID, req.ObjectName, req.Profile)
		return addTranscodingResponse{Err: err}, nil
	}
}

// GetTotalTasksQueued

type getTotalTasksQueuedRequest struct {
}

type getTotalTasksQueuedResponse struct {
	Total int   `json:"total"`
	Err   error `json:"error,omitempty"`
}

func (r getTotalTasksQueuedResponse) error() error { return r.Err }

func makeGetTotalTasksQueuedEndpoint(tms Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		total, err := tms.GetTotalTasksQueued()
		return getTotalTasksQueuedResponse{Total: total, Err: err}, nil
	}
}

// GetTotalTasksRunning

type getTotalTasksRunningRequest struct {
}

type getTotalTasksRunningResponse struct {
	Total int   `json:"total"`
	Err   error `json:"error,omitempty"`
}

func (r getTotalTasksRunningResponse) error() error { return r.Err }

func makeGetTotalTasksRunningEndpoint(tms Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		total, err := tms.GetTotalTasksRunning()
		return getTotalTasksRunningResponse{Total: total, Err: err}, nil
	}
}

// GetNextTask

type getNextTaskRequest struct {
	WorkerAddr string
}

type getNextTaskResponse struct {
	Task TranscodingTask `json:"task,omitempty"`
	Err  error           `json:"error,omitempty"`
}

func (r getNextTaskResponse) error() error { return r.Err }

func makeGetNextTaskEndpoint(tms Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getNextTaskRequest)
		task, err := tms.GetNextTask(req.WorkerAddr)
		return getNextTaskResponse{Task: task, Err: err}, nil
	}
}

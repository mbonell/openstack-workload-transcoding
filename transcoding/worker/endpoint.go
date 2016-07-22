package worker

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// GetStatus

type getStatusRequest struct {
}

type getStatusResponse struct {
	Status string`json:"status,omitempty"`
	Err error `json:"error,omitempty"`
}

func (r getStatusResponse) error() error { return r.Err }

func makeGetStatusEndpoint(tws Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		status, err := tws.GetStatus()
		return getStatusResponse{Status: status, Err: err}, nil
	}
}

// CancelTask

type cancelTaskRequest struct {
}

type cancelTaskResponse struct {
	Err error `json:"error,omitempty"`
}

func (r cancelTaskResponse) error() error { return r.Err }

func makeCancelTaskEndpoint(tws Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		err := tws.CancelTask()
		return getStatusResponse{Err: err}, nil
	}
}


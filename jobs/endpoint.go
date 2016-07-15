package jobs

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)


// AddNewJob

type addNewJobRequest struct {
	Job wttypes.Job
}

type addNewJobResponse struct {
	ID string	`json:"job_id,omitempty"`
	Err error	`json:"error,omitempty"`
}

func (r addNewJobResponse) error() error { return r.Err }

func makeAddNewJobEndpoint(js Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addNewJobRequest)
		id, err := js.AddNewJob(req.Job)
		return addNewJobResponse{ID: id, Err: err}, nil
	}
}

// GetJobStatus

type getJobStatusRequest struct {
	ID string
}

type getJobStatusResponse struct {
	Status string `json:"status,omitempty"`
	Err error     `json:"error,omitempty"`
}

func (r getJobStatusResponse) error() error { return r.Err }

func makeGetJobStatusEndpoint(js Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getJobStatusRequest)
		status, err := js.GetJobStatus(req.ID)
		return getJobStatusResponse{Status: status, Err: err}, nil
	}
}

// CancelJob

type cancelJobRequest struct {
	ID string
}

type cancelJobResponse struct {
	Err error     `json:"error,omitempty"`
}

func (r cancelJobResponse) error() error { return r.Err }

func makeCancelJobEndpoint(js Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(cancelJobRequest)
		err := js.CancelJob(req.ID)
		return cancelJobResponse{Err: err}, nil
	}
}

// UpdateTranscodingStatus

type updateTranscodingStatusRequest struct {
	jobID string
	ttID string
	status string		`json:"status"`
	objectname string	`json:"object_name"`
}

type updateTranscodingStatusResponse struct {
	Err error     `json:"error,omitempty"`
}

func (r updateTranscodingStatusResponse) error() error { return r.Err }

func makeUpdateTranscodingStatusEndpoint(js Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateTranscodingStatusRequest)
		err := js.UpdateTranscodingStatus(req.jobID, req.ttID, req.status, req.objectname)
		return updateTranscodingStatusResponse{Err: err}, nil
	}
}
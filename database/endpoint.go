package database

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

// ListJobs

type listJobsRequest struct {
}

type listJobsResponse struct {
	Jobs []wttypes.Job `json:"job,omitempty"`
	Err  error         `json:"error,omitempty"`
}

func (r listJobsResponse) error() error { return r.Err }

func makeListJobsEndpoint(ds Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(listJobsRequest)
		jobs, err := ds.ListJobs()
		return listJobsResponse{Jobs: jobs, Err: err}, nil
	}
}

// InsertJob

type insertJobRequest struct {
	Job wttypes.Job
}

type insertJobResponse struct {
	JobIDs wttypes.JobIDs `json:"job_ids,omitempty"`
	Err    error          `json:"error,omitempty"`
}

func (r insertJobResponse) error() error { return r.Err }

func makeInsertJobEndpoint(ds Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(insertJobRequest)
		ids, err := ds.InsertJob(req.Job)
		return insertJobResponse{JobIDs: ids, Err: err}, nil
	}
}

// UpdateJob

type updateJobRequest struct {
	Job wttypes.Job
}

type updateJobResponse struct {
	Err error	`json:"error,omitempty"`
}

func (r updateJobResponse) error() error { return r.Err }

func makeUpdateJobEndpoint(ds Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateJobRequest)
		err := ds.UpdateJob(req.Job)

		return updateJobResponse{Err: err}, nil
	}
}

// GetJob

type getJobRequest struct {
	ID string
}

type getJobResponse struct {
	Job *wttypes.Job `json:"job,omitempty"`
	Err error        `json:"error,omitempty"`
}

func (r getJobResponse) error() error { return r.Err }

func makeGetJobEndpoint(ds Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getJobRequest)
		job, err := ds.GetJob(req.ID)
		return getJobResponse{Job: &job, Err: err}, nil
	}
}

// UpdateTranscoding

type updateTranscodingRequest struct {
	Transcoding wttypes.TranscodingTask
}

type updateTranscodingResponse struct {
	Err error `json:"error,omitempty"`
}

func (r updateTranscodingResponse) error() error { return r.Err }

func makeUpdateTranscodingEndpoint(ds Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateTranscodingRequest)
		err := ds.UpdateTranscoding(req.Transcoding)

		return updateTranscodingResponse{Err: err}, nil
	}
}

// GetTranscoding

type getTranscodingRequest struct {
	ID string
}

type getTranscodingResponse struct {
	Transcoding wttypes.TranscodingTask `json:"transcoding,omitempty"`
	Err         error                   `json:"error,omitempty"`
}

func (r getTranscodingResponse) error() error { return r.Err }

func makeGetTranscodingEndpoint(ds Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getTranscodingRequest)
		t, err := ds.GetTranscoding(req.ID)

		return getTranscodingResponse{Transcoding: t, Err: err}, nil
	}
}

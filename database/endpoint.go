package database

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

// InsertJob

type insertJobRequest struct {
	Job wttypes.Job
}

type insertJobResponse struct {
	ID string	`json:"job_id,omitempty"`
	Err error	`json:"error,omitempty"`
}

func (r insertJobResponse) error() error { return r.Err }

func makeInsertJobEndpoint(ds Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(insertJobRequest)
		id, err := ds.InsertJob(req.Job)
		return insertJobResponse{ID: id, Err: err}, nil
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

// ListJobs

type listJobsRequest struct {
}

type listJobsResponse struct {
	Jobs []wttypes.Job 	`json:"job,omitempty"`
	Err error        	`json:"error,omitempty"`
}

func (r listJobsResponse) error() error { return r.Err }

func makeListJobsEndpoint(ds Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(listJobsRequest)
		jobs, err := ds.ListJobs()
		return listJobsResponse{Jobs: jobs, Err: err}, nil
	}
}


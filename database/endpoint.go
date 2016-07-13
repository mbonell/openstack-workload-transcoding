package database

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

type getJobRequest struct {
	ID string
}

type getJobResponse struct {
	Job *wttypes.Job 	`json:"job,omitempty"`
	Err  error 		`json:"error,omitempty"`
}

func (r getJobResponse) error() error { return r.Err }

func makeGetJobEndpoint(ds Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getJobRequest)
		job, err := ds.GetJob(req.ID)
		return getJobResponse{Job: &job, Err: err}, nil
	}
}

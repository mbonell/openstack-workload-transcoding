package monitor

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

// Register Worker

type registerWorkerRequest struct {
	Addr string
}

type registerWorkerResponse struct {
	Err error `json:"error,omitempty"`
}

func (r registerWorkerResponse) error() error { return r.Err }

func makeRegisterWorkerEndpoint(tms Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(registerWorkerRequest)
		err := tms.RegisterWorker(req.Addr)
		return registerWorkerResponse{Err: err}, nil
	}
}

// Deregister Worker

type deregisterWorkerRequest struct {
	Addr string
}

type deregisterWorkerResponse struct {
	Err error `json:"error,omitempty"`
}

func (r deregisterWorkerResponse) error() error { return r.Err }

func makeDeregisterWorkerEndpoint(tms Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deregisterWorkerRequest)
		err := tms.DeregisterWorker(req.Addr)
		return deregisterWorkerResponse{Err: err}, nil
	}
}

// Change Worker Status

type updateWorkerStatusRequest struct {
	WS wttypes.WorkerStatus
}

type updateWorkerStatusResponse struct {
	Err error `json:"error,omitempty"`
}

func (r updateWorkerStatusResponse) error() error { return r.Err }

func makeUpdateWorkerStatusEndpoint(tms Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateWorkerStatusRequest)
		err := tms.UpdateWorkerStatus(req.WS)
		return deregisterWorkerResponse{Err: err}, nil
	}
}

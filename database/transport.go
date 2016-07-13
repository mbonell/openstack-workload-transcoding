package database

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"

	"github.com/gorilla/mux"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
	"fmt"
)

// MakeHandler returns a handler for the database service.
func MakeHandler(ctx context.Context, ds Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	// test: curl -k -H "Content-Type: application/json" -d '{"url_media":"http://obazavil-nuc/big_buck_bunny_720p_1mb.mp4", "video_name":"conejo", "transcoding_targets":[{"name":"transcode-1","profile":"iphone6","objectname":""},{"name":"transcode-2","profile":"ipadmini","object_name":""}]}' -X POST https://localhost:8080/v1/jobs
	insertJobHandler := kithttp.NewServer(
		ctx,
		makeInsertJobEndpoint(ds),
		decodeInsertJobRequest,
		encodeResponse,
		opts...,
	)

	//test: curl -k -H "Content-Type: application/json" -d '{"id":"1", "url_media":"fake_url", "video_name":"fake_conejo"}' -X PUT https://localhost:8080/v1/jobs/1
	updateJobHandler := kithttp.NewServer(
		ctx,
		makeUpdateJobEndpoint(ds),
		decodeUpdateJobRequest,
		encodeResponse,
		opts...,
	)

	// test: curl -k https://localhost:8080/v1/jobs/1
	getJobHandler := kithttp.NewServer(
		ctx,
		makeGetJobEndpoint(ds),
		decodeGetJobRequest,
		encodeResponse,
		opts...,
	)

	// test: curl -k https://localhost:8080/v1/jobs
	listJobsHandler := kithttp.NewServer(
		ctx,
		makeListJobsEndpoint(ds),
		decodeListJobsRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/v1/jobs", insertJobHandler).Methods("POST")
	r.Handle("/v1/jobs/{id}", updateJobHandler).Methods("PUT")
	r.Handle("/v1/jobs/{id}", getJobHandler).Methods("GET")
	r.Handle("/v1/jobs", listJobsHandler).Methods("GET")

	return r

}

func decodeInsertJobRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var job wttypes.Job

	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		return nil, err
	}

	//TODO: Decode not always throws error, extra validate all needed fields "decoded:  {    [] }"
	//TODO: validate ID is empty

	return insertJobRequest{Job:job}, nil
}

func decodeUpdateJobRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var job wttypes.Job

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, wttypes.ErrBadRoute
	}

	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		return nil, err
	}

	//TODO: Decode not always throws error, extra validate all needed fields "decoded:  {    [] }"

	if id != job.ID {
		return nil, wttypes.ErrMismatchID
	}

	return updateJobRequest{Job:job}, nil
}


func decodeGetJobRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, wttypes.ErrBadRoute
	}
	return getJobRequest{ID: string(id)}, nil
}

func decodeListJobsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return listJobsRequest{}, nil
}

type errorer interface {
	error() error
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch err {
	case wttypes.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case wttypes.ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

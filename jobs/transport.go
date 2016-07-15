package jobs

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
func MakeHandler(ctx context.Context, js Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	// test: curl -k -H "Content-Type: application/json" -d '{"url_media":"http://obazavil-nuc/big_buck_bunny_720p_1mb.mp4", "video_name":"conejo", "transcoding_targets":[{"name":"transcode-1","profile":"iphone6","objectname":""},{"name":"transcode-2","profile":"ipadmini","object_name":""}]}' -X POST https://localhost:8081/v1/jobs
	addNewJobHandler := kithttp.NewServer(
		ctx,
		makeAddNewJobEndpoint(js),
		decodeAddNewJobRequest,
		encodeResponse,
		opts...,
	)

	// test: curl -k https://localhost:8081/v1/jobs/1/status
	getJobStatusHandler := kithttp.NewServer(
		ctx,
		makeGetJobStatusEndpoint(js),
		decodeGetJobStatusRequest,
		encodeResponse,
		opts...,
	)

	// test: curl -k https://localhost:8081/v1/jobs/1/cancel
	cancelJobHandler := kithttp.NewServer(
		ctx,
		makeCancelJobEndpoint(js),
		decodeCancelJobRequest,
		encodeResponse,
		opts...,
	)

	// test: curl -k -H "Content-Type: application/json" -d '{"status":"running", "objectname":"myobjectname"}' -X PUT https://localhost:8081/v1/jobs/1/transcoding/1/status
	updateTranscodingStatusHandler := kithttp.NewServer(
		ctx,
		makeUpdateTranscodingStatusEndpoint(js),
		decodeUpdateTranscodingStatusRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/v1/jobs", addNewJobHandler).Methods("POST")
	r.Handle("/v1/jobs/{id}/status", getJobStatusHandler).Methods("GET")
	r.Handle("/v1/jobs/{id}/cancel", cancelJobHandler).Methods("GET")
	r.Handle("/v1/jobs/{id}/transcoding/{ttid}/status", updateTranscodingStatusHandler).Methods("PUT")

	return r

}

func decodeAddNewJobRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var job wttypes.Job

	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		return nil, err
	}

	//TODO: Decode not always throws error, extra validate all needed fields "decoded:  {    [] }"
	//TODO: validate ID is empty

	return addNewJobRequest{Job:job}, nil
}

func decodeGetJobStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		return nil, wttypes.ErrBadRoute
	}
	return getJobStatusRequest{ID: string(id)}, nil
}

func decodeCancelJobRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		return nil, wttypes.ErrBadRoute
	}
	return cancelJobRequest{ID: string(id)}, nil
}

func decodeUpdateTranscodingStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var v updateTranscodingStatusRequest

	vars := mux.Vars(r)

	jobID, ok := vars["id"]
	if !ok {
		return nil, wttypes.ErrBadRoute
	}

	ttID, ok := vars["ttid"]
	if !ok {
		return nil, wttypes.ErrBadRoute
	}

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return nil, err
	}

	//TODO: Decode not always throws error, extra validate all needed fields "decoded:  {    [] }"
	//TODO: validate malformed struct

	fmt.Println("decodeUpdateTranscodingStatusRequest:", v)

	v.jobID = jobID
	v.ttID = ttID
	return v, nil
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


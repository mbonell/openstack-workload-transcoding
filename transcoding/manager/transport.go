package manager

import (
	"fmt"
	"net/http"
	"encoding/json"

	"golang.org/x/net/context"

	"github.com/gorilla/mux"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

// MakeHandler returns a handler for the transcoding manager service.
func MakeHandler(ctx context.Context, tms Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	// test: curl -k -H "Content-Type: application/json" -d '{"id":"1", "object_name":"rabbitobject", "profile":"iPhone5s"}' -X POST https://localhost:8082/transcodings
	addTranscodingHandler := kithttp.NewServer(
		ctx,
		makeAddTranscodingEndpoint(tms),
		decodeAddTranscodingRequest,
		encodeResponse,
		opts...,
	)

	// test: curl -k -H "Content-Type: application/json" -d '{"id":"1", "job_id":"1", "object_name":"myobjectname2", "profile":"iPhone5s"}' -X POST https://localhost:8082/transcodings
	getNextTaskHandler := kithttp.NewServer(
		ctx,
		makeGetNextTaskEndpoint(tms),
		decodeGetNextTaskRequest,
		encodeResponse,
		opts...,
	)

	// test: curl -k https://localhost:8082/transcodings/queued
	getTotalTasksQueuedHandler := kithttp.NewServer(
		ctx,
		makeGetTotalTasksQueuedEndpoint(tms),
		decodeGetTotalTasksQueuedRequest,
		encodeResponse,
		opts...,
	)

	// test: curl -k https://localhost:8082/transcodings?worker=127.0.0.1
	getTotalTasksRunningHandler := kithttp.NewServer(
		ctx,
		makeGetTotalTasksRunningEndpoint(tms),
		decodeGetTotalTasksRunningRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/transcodings", addTranscodingHandler).Methods("POST")
	r.Handle("/tasks", getNextTaskHandler).Methods("GET")
	r.Handle("/transcodings/queued", getTotalTasksQueuedHandler).Methods("GET")
	r.Handle("/transcodings/running", getTotalTasksRunningHandler).Methods("GET")

	return r

}

func decodeAddTranscodingRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var t wttypes.TranscodingTask

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return nil, err
	}

	fmt.Println("[manager]", "decoded: ", t)

	//TODO: Decode not always throws error, extra validate all needed fields "decoded:  {    [] }"
	//TODO: validate ID is empty

	return addTranscodingRequest{
		ID: t.ID,
		ObjectName: t.ObjectName,
		Profile: t.Profile,
	}, nil
}

func decodeGetNextTaskRequest(_ context.Context, r *http.Request) (interface{}, error) {
	worker := r.FormValue("worker")
	if worker == "" {
		return nil, wttypes.ErrInvalidArgument
	}

	return getNextTaskRequest{
		WorkerAddr: worker,
	}, nil
}

func decodeGetTotalTasksQueuedRequest(_ context.Context, r *http.Request) (interface{}, error) {

	return getTotalTasksQueuedRequest{}, nil
}

func decodeGetTotalTasksRunningRequest(_ context.Context, r *http.Request) (interface{}, error) {

	return getTotalTasksQueuedRequest{}, nil
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


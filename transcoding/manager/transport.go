package manager

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"

	"github.com/gorilla/mux"

	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"

	"fmt"
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

	// test: curl -k -H "Content-Type: application/json" -d '{"status":"running"}' -X PUT https://localhost:8082/tasks/1/status
	updateTaskStatusHandler := kithttp.NewServer(
		ctx,
		makeUpdateTaskStatusEndpoint(tms),
		decodeUpdateTaskStatusRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/tasks", addTranscodingHandler).Methods("POST")
	r.Handle("/tasks", getNextTaskHandler).Methods("GET")
	r.Handle("/tasks/queued", getTotalTasksQueuedHandler).Methods("GET")
	r.Handle("/tasks/running", getTotalTasksRunningHandler).Methods("GET")
	r.Handle("/tasks/{id}/status", updateTaskStatusHandler).Methods("PUT")

	return r

}

func decodeAddTranscodingRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var t wttypes.TranscodingTask

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return nil, err
	}

	fmt.Println("[manager] decodeAddTranscodingRequest:", t)

	if t.ID == "" || t.ObjectName == "" || t.Profile == "" {
		return nil, wttypes.ErrInvalidArgument
	}

	return addTranscodingRequest{
		ID:         t.ID,
		ObjectName: t.ObjectName,
		Profile:    t.Profile,
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

func decodeUpdateTaskStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Status string `json:"status"`
	}

	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		return nil, wttypes.ErrBadRoute
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}


	return updateTaskStatusRequest{ID: id, Status: body.Status}, nil
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

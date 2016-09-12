package worker

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"

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

	// test: curl -k https://localhost:8083/worker/status
	getStatusHandler := kithttp.NewServer(
		ctx,
		makeGetStatusEndpoint(tms),
		decodeGetStatusRequest,
		encodeResponse,
		opts...,
	)

	// test: curl -k -X DELETE https://localhost:8083/tasks
	cancelTaskHandler := kithttp.NewServer(
		ctx,
		makeCancelTaskEndpoint(tms),
		decodeCancelTaskRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/worker/status", getStatusHandler).Methods("GET")
	r.Handle("/tasks", cancelTaskHandler).Methods("DELETE")

	return r
}

func decodeGetStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return getStatusRequest{}, nil
}

func decodeCancelTaskRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return cancelTaskRequest{}, nil
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

package monitor

import (
	"encoding/json"
	"fmt"
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

	// test: curl -k -H "Content-Type: application/json" -d '{"addr":"myip"}' -X POST https://localhost:8084/workers
	registerWorkerHandler := kithttp.NewServer(
		ctx,
		makeRegisterWorkerEndpoint(tms),
		decodeRegisterWorkerRequest,
		encodeResponse,
		opts...,
	)

	// test: curl -k -H "Content-Type: application/json" -d '{"addr":"myip"}' -X DELETE https://localhost:8084/workers
	deregisterWorkerHandler := kithttp.NewServer(
		ctx,
		makeDeregisterWorkerEndpoint(tms),
		decodeDeregisterWorkerRequest,
		encodeResponse,
		opts...,
	)

	// test: curl -k -H "Content-Type: application/json" -d '{"addr":"myip", "status":"idle"}' -X PUT https://localhost:8084/workers/status
	updateWorkerStatusHandler := kithttp.NewServer(
		ctx,
		makeUpdateWorkerStatusEndpoint(tms),
		decodeUpdateWorkerStatusRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/workers", registerWorkerHandler).Methods("POST")
	r.Handle("/workers", deregisterWorkerHandler).Methods("DELETE")

	r.Handle("/workers/status", updateWorkerStatusHandler).Methods("PUT")

	return r

}

func decodeRegisterWorkerRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var ws wttypes.WorkerStatus

	if err := json.NewDecoder(r.Body).Decode(&ws); err != nil {
		return nil, err
	}

	fmt.Println("[manager] decodeRegisterWorkerRequest:", ws)

	if ws.Addr == "" {
		return nil, wttypes.ErrInvalidArgument
	}

	return registerWorkerRequest{
		Addr: ws.Addr,
	}, nil
}

func decodeDeregisterWorkerRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var ws wttypes.WorkerStatus

	if err := json.NewDecoder(r.Body).Decode(&ws); err != nil {
		return nil, err
	}

	fmt.Println("[manager] decodeDeregisterWorkerRequest:", ws)

	if ws.Addr == "" {
		return nil, wttypes.ErrInvalidArgument
	}

	return deregisterWorkerRequest{
		Addr: ws.Addr,
	}, nil
}

func decodeUpdateWorkerStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var ws wttypes.WorkerStatus

	if err := json.NewDecoder(r.Body).Decode(&ws); err != nil {
		return nil, err
	}

	fmt.Println("[manager] decodeDeregisterWorkerRequest:", ws)

	if ws.Addr == "" || ws.Status == "" {
		return nil, wttypes.ErrInvalidArgument
	}

	return updateWorkerStatusRequest{
		WS: ws,
	}, nil
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

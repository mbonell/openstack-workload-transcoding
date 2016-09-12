package wtcommon

import (
	"net/http"
)

const (
	DATABASE_PORT = "8080"
	JOBS_PORT     = "8081"
	MANAGER_PORT  = "8082"
	WORKER_PORT   = "8083"
	MONITOR_PORT  = "8084"
)

// AccessControl returns a handler for the access control
func AccessControl(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

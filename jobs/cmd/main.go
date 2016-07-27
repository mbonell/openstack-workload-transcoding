package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"

	"github.com/obazavil/openstack-workload-transcoding/jobs"
	"github.com/obazavil/openstack-workload-transcoding/wtcommon"
)

func main() {
	var err error

	errs := make(chan error, 2)

	var (
		httpAddr = flag.String("http.addr", ":8081", "Address for HTTP (JSON) jobs server")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}

	var ctx context.Context
	{
		ctx = context.Background()
	}

	var js jobs.Service
	{
		js, err = jobs.NewService()
		if err != nil {
			errs <- err
		}
	}

	httpLogger := log.NewContext(logger).With("component", "http")

	mux := http.NewServeMux()

	mux.Handle("/", jobs.MakeHandler(ctx, js, httpLogger))

	http.Handle("/", wtcommon.AccessControl(mux))

	go func() {
		logger.Log("transport", "http", "address", *httpAddr, "msg", "listening")
		errs <- http.ListenAndServeTLS(*httpAddr, "../../certs/server.pem", "../../certs/server.key", nil)
	}()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}

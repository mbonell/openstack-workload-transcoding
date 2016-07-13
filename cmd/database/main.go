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

	"github.com/obazavil/openstack-workload-transcoding/database"
	"github.com/obazavil/openstack-workload-transcoding/wtcommon"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
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

	var ds database.Service
	{
		ds = database.NewService()
	}

	httpLogger := log.NewContext(logger).With("component", "http")

	mux := http.NewServeMux()

	mux.Handle("/v1/", database.MakeHandler(ctx, ds, httpLogger))

	http.Handle("/", wtcommon.AccessControl(mux))

	errs := make(chan error, 2)
	go func() {
		logger.Log("transport", "http", "address", *httpAddr, "msg", "listening")
		errs <- http.ListenAndServeTLS(*httpAddr, "../../ssl/server.pem", "../../ssl/server.key", nil)
	}()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"os/exec"
	"path"
	"strings"

	"golang.org/x/net/context"
	"github.com/go-resty/resty"
	"github.com/go-kit/kit/log"

	"github.com/obazavil/openstack-workload-transcoding/transcoding/worker"
	"github.com/obazavil/openstack-workload-transcoding/wtcommon"
	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

const (
	DELAY = 15 * time.Second
)

func main() {
	var err error

	errs := make(chan error, 2)

	var (
		httpAddr = flag.String("http.addr", ":8083", "Address for HTTP (JSON) transcoding worker server")
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

	var tws worker.Service
	{
		tws, err = worker.NewService()
		if err != nil {
			errs <- err
		}
	}

	tws.WorkerUpdateStatus(wttypes.WORKER_STATUS_ONLINE)

	httpLogger := log.NewContext(logger).With("component", "http")

	mux := http.NewServeMux()

	mux.Handle("/", worker.MakeHandler(ctx, tws, httpLogger))

	http.Handle("/", wtcommon.AccessControl(mux))

	go func() {
		logger.Log("transport", "http", "address", *httpAddr, "msg", "listening")
		errs <- http.ListenAndServeTLS(*httpAddr, "certs/server.pem", "certs/server.key", nil)
	}()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// Transcoding go func
	//forever := make(chan bool)
	go func() {
		// OpenStack
		provider, err := wtcommon.GetProvider()
		if err != nil {
			errs <- err
		}

		serviceObjectStorage, err := wtcommon.GetServiceObjectStorage(provider)
		if err != nil {
			errs <- err
		}

		tws.WorkerUpdateStatus(wttypes.WORKER_STATUS_IDLE)

		for {
			// Ask manager for work
			resp, err := resty.R().
				Get(wtcommon.Servers["manager"] + "/tasks?worker=" + tws.GetIP())

			// Error in communication? sleep and retry
			if err != nil {
				time.Sleep(DELAY)
				continue
			}

			// Get response
			str := resp.String()

			// There was an error? sleep and retry
			if strings.HasPrefix(str, `{"error"`) {
				time.Sleep(DELAY)
				continue
			}

			// Decode into task type
			task, err := wtcommon.JSON2Task(str)
			if err != nil {
				time.Sleep(DELAY)
				continue
			}

			fmt.Println("[worker] received task:", task)

			// Everything fine so far, let's update our status
			tws.WorkerUpdateStatus(wttypes.WORKER_STATUS_IDLE)
			tws.NotifyTaskStatus(task.ID, wttypes.TRANSCODING_RUNNING, "")

			// Names and paths of our media
			fnOriginal := path.Join(os.TempDir(),
				fmt.Sprintf("%s.mp4",
					task.ObjectName,
				))

			vnTranscoded := fmt.Sprintf("%s-%s.mp4",
						task.ObjectName,
						task.Profile,
				)
			fnTranscoded := path.Join(os.TempDir(), vnTranscoded)

			// Download media from object storage
			err = wtcommon.DownloadFromObjectStorage(serviceObjectStorage, task.ObjectName, fnOriginal)
			if err != nil {
				tws.WorkerUpdateStatus(wttypes.WORKER_STATUS_IDLE)
				tws.NotifyTaskStatus(task.ID, wttypes.TRANSCODING_ERROR, "")
				time.Sleep(DELAY)
				continue
			}

			// Get profile information
			p, ok := wttypes.NewProfile()[task.Profile]
			if !ok {
				fmt.Printf("[err] Profile %s doesn't exist.\n",
					task.Profile)

				tws.WorkerUpdateStatus(wttypes.WORKER_STATUS_IDLE)
				tws.NotifyTaskStatus(task.ID, wttypes.TRANSCODING_ERROR, "")
				time.Sleep(DELAY)
				continue
			}

			// Execute ffmpeg
			args := []string{"-i", fnOriginal}

			args = append(args, strings.Split(p.FFMPEG.Args, " ")...)

			if p.Resolution != "" {
				args = append(args, "-s")
				args = append(args, p.Resolution)
			}

			args = append(args, fnTranscoded)

			cmd := exec.Command("ffmpeg", args...)

			// Remove target file just in case before we start
			os.Remove(fnTranscoded)

			err = cmd.Start()
			if err != nil {
				fmt.Printf("[err] ffmpeg: %s.\n",
					err)

				tws.WorkerUpdateStatus(wttypes.WORKER_STATUS_IDLE)
				tws.NotifyTaskStatus(task.ID, wttypes.TRANSCODING_ERROR, "")
				time.Sleep(DELAY)
				continue
			}

			// Update process in the service (por cancellation purposes)
			tws.WorkerUpdateProcess(cmd.Process)
			fmt.Println("ENCODING...")

			// Wait for ffmpeg to finish
			exitCode := 0
			errWait := cmd.Wait()
			if errWait != nil {
				if exiterr, ok := errWait.(*exec.ExitError); ok {
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						exitCode = status.ExitStatus()
					}
				}
			}

			var status string
			switch {
			case errWait == nil:
				status = wttypes.TRANSCODING_FINISHED
			case exitCode == 255:
				status = wttypes.TRANSCODING_CANCELLED
			default:
				status = wttypes.TRANSCODING_ERROR
			}

			var objectname string
			if status == wttypes.TRANSCODING_FINISHED {
				objectname, err = wtcommon.Upload2ObjectStorage(serviceObjectStorage, fnTranscoded, vnTranscoded, wtcommon.TRANSCODED_MEDIA_CONTAINER)
				if err != nil {
					fmt.Printf("[err] object storage: %s.\n",
						err)

					status = wttypes.TRANSCODING_ERROR
				}
			}

			os.Remove(fnTranscoded)
			tws.WorkerUpdateStatus(wttypes.WORKER_STATUS_IDLE)
			tws.WorkerUpdateProcess(nil)

			tws.NotifyTaskStatus(task.ID, status, objectname)

			time.Sleep(DELAY)
		}
	}()

	logger.Log("terminated", <-errs)
	tws.WorkerUpdateStatus(wttypes.WORKER_STATUS_OFFLINE)
}

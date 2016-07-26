package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/go-resty/resty"

	"github.com/go-kit/kit/log"

	"github.com/obazavil/openstack-workload-transcoding/transcoding/worker"
	"github.com/obazavil/openstack-workload-transcoding/wtcommon"
	"github.com/obazavil/openstack-workload-transcoding/wttypes"
	"net"
	"os/exec"
	"path"
	"strings"
)

const (
	DELAY = 15 * time.Second
)

// Get outbound ip of this machine
func getIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	addr := conn.LocalAddr().String()
	idx := strings.LastIndex(addr, ":")

	return addr[0:idx], nil
}

func notifyTaskStatus(id string, status string, objectname string) {
	fmt.Println("[worker] notifyTaskStatus:", id, status, objectname)
	// Update Manager Service
	bodyM := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	fmt.Println("[main] statusM", bodyM.Status)
	resp, err := resty.R().
		SetBody(bodyM).
		Put(fmt.Sprintf("%s/tasks/%s/status",
		wtcommon.Servers["manager"],
		id))

	if err != nil {
		fmt.Println("[worker] notify err:", err)
		//TODO: do something when status update fails
	}

	str := resp.String()
	if strings.HasPrefix(str, `{"error"`) {
		fmt.Println("[worker] notify err:", err)
		//TODO: do something when status update fails
	}

	fmt.Println("notified manager and jobs:", id, status, objectname)

	// Update Jobs Service
	bodyJ := struct {
		Status     string `json:"status"`
		ObjectName string `json:"object_name,omitempty"`
	}{
		Status:     status,
		ObjectName: objectname,
	}

	fmt.Println("[main] statusJ", bodyJ.Status)
	resp, err = resty.R().
		SetBody(bodyJ).
		Put(fmt.Sprintf("%s/transcodings/%s/status",
			wtcommon.Servers["jobs"],
			id))

	if err != nil {
		fmt.Println("[worker] notify err:", err)
		//TODO: do something when status update fails
	}

	str = resp.String()
	if strings.HasPrefix(str, `{"error"`) {
		fmt.Println("[worker] notify err:", err)
		//TODO: do something when status update fails
	}
}
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
		tws = worker.NewService()
	}

	var ip string
	{
		ip, err = getIP()
		if err != nil {
			errs <- err
		}
	}

	httpLogger := log.NewContext(logger).With("component", "http")

	mux := http.NewServeMux()

	mux.Handle("/", worker.MakeHandler(ctx, tws, httpLogger))

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

		for {
			// Ask manager for work
			resp, err := resty.R().
				Get(wtcommon.Servers["manager"] + "/tasks?worker=" + ip)


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
			tws.WorkerUpdateStatus(worker.WORKER_STATUS_BUSY)
			notifyTaskStatus(task.ID, wttypes.TRANSCODING_RUNNING, "")

			// Filenames of our media
			fnOriginal := path.Join(os.TempDir(),
				fmt.Sprintf("%s.mp4",
					task.ObjectName,
				))
			fnTranscoded := path.Join(os.TempDir(),
				fmt.Sprintf("%s-%s.mp4",
					task.ObjectName,
					task.Profile,
				))

			// Download media from object storage
			err = wtcommon.DownloadFromObjectStorage(serviceObjectStorage, task.ObjectName, fnOriginal)
			if err != nil {
				tws.WorkerUpdateStatus(worker.WORKER_STATUS_IDLE)
				notifyTaskStatus(task.ID, wttypes.TRANSCODING_ERROR, "")
				time.Sleep(DELAY)
				continue
			}

			// Get profile information
			p, ok := wttypes.NewProfile()[task.Profile]
			if !ok {
				fmt.Printf("[err] Profile %s doesn't exist.\n",
					task.Profile)

				tws.WorkerUpdateStatus(worker.WORKER_STATUS_IDLE)
				notifyTaskStatus(task.ID, wttypes.TRANSCODING_ERROR, "")
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

				tws.WorkerUpdateStatus(worker.WORKER_STATUS_IDLE)
				notifyTaskStatus(task.ID, wttypes.TRANSCODING_ERROR, "")
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
				wtcommon.Upload2ObjectStorage(serviceObjectStorage, fnTranscoded, fnTranscoded)
			}

			//TODO: upload into swift


			os.Remove(fnTranscoded)
			tws.WorkerUpdateStatus(worker.WORKER_STATUS_IDLE)
			tws.WorkerUpdateProcess(nil)


			notifyTaskStatus(task.ID, status, objectname)

			time.Sleep(DELAY)
		}
	}()

	logger.Log("terminated", <-errs)
}

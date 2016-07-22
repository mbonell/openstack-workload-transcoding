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
	"path"
	"strings"
	"os/exec"
	"net"
)

const (
	DELAY = 15 * time.Second
)

// Get outbound ip of this machine
func GetIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	addr := conn.LocalAddr().String()
	idx := strings.LastIndex(addr, ":")

	return addr[0:idx], nil
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
		tws, err = worker.NewService()
		if err != nil {
			errs <- err
		}
	}

	var ip string
	{
		ip, err = GetIP()
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
		for {
			// Wait
			time.Sleep(DELAY)

			// Ask manager for work
			resp, err := resty.R().Get(wtcommon.Servers["manager"] + "/tasks?worker=" + ip)

			fmt.Println("called REST")

			// Error in communication? sleep and retry
			if err != nil {
				continue
			}

			str := resp.String()

			fmt.Println("response:", str)
			logger.Log("resp", str)

			// There was an error? sleep and retry
			if strings.HasPrefix(str, `{"error"`) {
				continue
			}

			task, err := wtcommon.JSON2Task(str)
			if err != nil {
				logger.Log("JSONErr", err.Error())

				continue
			}

			// Everything fine so far, let's update our status
			fmt.Println("transcoding started: ", str)
			tws.WorkerUpdateStatus(worker.WORKER_STATUS_BUSY)

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

			//// Download media from object storage
			//err = wtcommon.DownloadFromObjectStorage(task.ObjectName, fnOriginal)
			//if err != nil {
			//	fmt.Println("[err] Couldn't download object '%s': %s\n",
			//		task.ObjectName,
			//		err)
			//	//TODO: update status in manager
			//	tws.WorkerUpdateStatus(worker.WORKER_STATUS_IDLE)
			//	continue
			//}
			fnOriginal = "/Users/obazavil/Downloads/videos/id4.mp4"
			fmt.Println("transcoded:", fnTranscoded)

			// Get profile information
			p, ok := wttypes.NewProfile()[task.Profile]
			if !ok {
				fmt.Printf("[err] Profile %s doesn't exist.\n",
					task.Profile)

				//TODO: update status in manager
				tws.WorkerUpdateStatus(worker.WORKER_STATUS_IDLE)
				continue
			}

			// Execute ffmpeg
			args := []string {"-i", fnOriginal}

			args = append(args, strings.Split(p.FFMPEG.Args, " ")...)

			if p.Resolution != "" {
				args = append(args, "-s")
				args = append(args, p.Resolution)
			}

			args = append(args, fnTranscoded)

			cmd := exec.Command("ffmpeg", args...)

			fmt.Println("args:", cmd.Args)

			// Remove target file just in case
			os.Remove(fnTranscoded)

			err = cmd.Start()
			if err != nil {
				fmt.Printf("[err] ffmpeg: %s.\n",
					err)

				//TODO: update status in manager
				tws.WorkerUpdateStatus(worker.WORKER_STATUS_IDLE)
				continue
			}
			tws.WorkerUpdateProcess(cmd.Process)

			err = cmd.Wait()
			if err != nil {
				if exiterr, ok := err.(*exec.ExitError); ok {
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						fmt.Printf("Exit Status: %d", status.ExitStatus())
					}
				} else {
					fmt.Printf("cmd.Wait: %v", err)
				}

				//TODO: update status in manager
				tws.WorkerUpdateStatus(worker.WORKER_STATUS_IDLE)
				continue
			}

			//TODO: update status in manager
			tws.WorkerUpdateProcess(nil)
			tws.WorkerUpdateStatus(worker.WORKER_STATUS_IDLE)
			fmt.Println("transcoding finished")
		}

	}()

	logger.Log("terminated", <-errs)
}

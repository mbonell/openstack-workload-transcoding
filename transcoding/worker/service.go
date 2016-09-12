package worker

import (
	"crypto/tls"
	"fmt"
	"github.com/go-resty/resty"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

// Service is the interface that provides transcoding worker methods.
type Service interface {
	// Get the status
	GetStatus() (string, error)

	// Cancel a transcoding task
	CancelTask() error

	// No Endpoints (REST API) api for below functions

	WorkerUpdateStatus(status string)

	WorkerUpdateProcess(p *os.Process)

	NotifyWorkerStatus(status string)

	NotifyTaskStatus(id string, status string, objectname string)

	GetIP() string
}

type service struct {
	mtx     sync.RWMutex
	status  string
	process *os.Process
	ip      string

	jobs    string
	manager string
	monitor string
}

func (s *service) GetStatus() (string, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.status, nil
}

func (s *service) CancelTask() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	fmt.Println("cancelling task...")

	if s.status == wttypes.WORKER_STATUS_IDLE {
		return wttypes.ErrNoTaskRunning
	}

	if s.process == nil {
		return wttypes.ErrNoProcessRunning
	}

	s.process.Signal(syscall.SIGTERM)

	return nil
}

// No Endpoints (REST API) api for below functions

func (s *service) WorkerUpdateStatus(status string) {
	s.mtx.Lock()
	s.status = status
	s.mtx.Unlock()

	s.NotifyWorkerStatus(status)
}

func (s *service) WorkerUpdateProcess(p *os.Process) {
	s.mtx.Lock()
	s.process = p
	s.mtx.Unlock()
}

func (s *service) NotifyWorkerStatus(status string) {
	fmt.Println("[worker] notifyWorkerStatus:", status)

	// Update Monitor Service
	st := wttypes.WorkerStatus{
		Addr:   s.ip,
		Status: status,
	}
	fmt.Println("[main] worker status", st)

	resp, err := resty.R().
		SetBody(st).
		Put(fmt.Sprintf("%s/workers/status",
			s.monitor))

	if err != nil {
		fmt.Println("[worker] notify worker err:", err)
		//TODO: do something when status update fails
	}

	str := resp.String()
	if strings.HasPrefix(str, `{"error"`) {
		fmt.Println("[worker] notify worker err:", err)
		//TODO: do something when status update fails
	}
}

func (s *service) NotifyTaskStatus(id string, status string, objectname string) {
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
			s.manager,
			id))

	if err != nil {
		fmt.Println("[worker] notify task err:", err)
		//TODO: do something when status update fails
	}

	str := resp.String()
	if strings.HasPrefix(str, `{"error"`) {
		fmt.Println("[worker] notify task err:", err)
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
			s.jobs,
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

func (s *service) GetIP() string {
	return s.ip
}

// Get outbound ip of this machine
func getOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	addr := conn.LocalAddr().String()
	idx := strings.LastIndex(addr, ":")

	return addr[0:idx], nil
}

// NewService creates a transcoding worker service with necessary dependencies.
func NewService(jobs, manager, monitor string) (Service, error) {
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	ip, err := getOutboundIP()
	if err != nil {
		return &service{}, err
	}

	return &service{
		mtx:    sync.RWMutex{},
		status: wttypes.WORKER_STATUS_IDLE,
		ip:     ip,

		jobs:    jobs,
		manager: manager,
		monitor: monitor,
	}, nil
}

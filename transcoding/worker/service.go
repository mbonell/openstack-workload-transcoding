package worker

import (
	"os"
	"sync"
	"syscall"
	"crypto/tls"

	"github.com/go-resty/resty"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
	"fmt"
)

const (
	WORKER_STATUS_IDLE = "idle"
	WORKER_STATUS_BUSY = "busy"
)

// Service is the interface that provides transcoding worker methods.
type Service interface {
	// Get the status
	GetStatus() (string, error)

	// Cancel a transcoding task
	CancelTask() (error)

	WorkerUpdateStatus(status string)

	WorkerUpdateProcess(p *os.Process)
}

type service struct {
	mtx     sync.RWMutex
	status  string
	process *os.Process
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

	if s.status == WORKER_STATUS_IDLE {
		return wttypes.ErrNoTaskRunning
	}

	if s.process == nil {
		return wttypes.ErrNoProcessRunning
	}

	s.process.Signal(syscall.SIGTERM)

	return nil
}

// No REST api for below functions

func (s *service) WorkerUpdateStatus(status string) {
	s.mtx.Lock()
	s.status = status
	s.mtx.Unlock()
}

func (s *service) WorkerUpdateProcess(p *os.Process) {
	s.mtx.Lock()
	s.process = p
	s.mtx.Unlock()
}


// NewService creates a transcoding worker service with necessary dependencies.
func NewService() Service {
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	return &service{
		mtx:    sync.RWMutex{},
		status: WORKER_STATUS_IDLE,
	}
}

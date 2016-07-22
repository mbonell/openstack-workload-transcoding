package worker

import (
	"os"
	"crypto/tls"
	"sync"

	"github.com/go-resty/resty"

	"github.com/rackspace/gophercloud"

	//"github.com/obazavil/openstack-workload-transcoding/wtcommon"
	"github.com/obazavil/openstack-workload-transcoding/wttypes"
	"syscall"
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

	provider             *gophercloud.ProviderClient
	serviceObjectStorage *gophercloud.ServiceClient
}

func (s *service) GetStatus() (string, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.status, nil
}

func (s *service) CancelTask() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.status == WORKER_STATUS_IDLE {
		return wttypes.ErrNoTaskRunning
	}

	if s.process == nil {
		return wttypes.ErrNoProcessRunning
	}

	fmt.Println("read process: ", s.process)

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

	fmt.Println("updated process: ", p.Pid)
}


// NewService creates a transcoding worker service with necessary dependencies.
func NewService() (Service, error) {
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	//provider, err := wtcommon.GetProvider()
	//if err != nil {
	//	return &service{}, err
	//}
	//
	//serviceObjectStorage, err := wtcommon.GetServiceObjectStorage(provider)
	//if err != nil {
	//	return &service{}, err
	//}
	//
	//return &service{
	//	mtx:    sync.RWMutex{},
	//	status: WORKER_STATUS_IDLE,
	//
	//	provider:             provider,
	//	serviceObjectStorage: serviceObjectStorage,
	//}, nil

	return &service{
		mtx:    sync.RWMutex{},
		status: WORKER_STATUS_IDLE,
	}, nil
}

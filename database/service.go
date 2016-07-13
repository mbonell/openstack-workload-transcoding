package database

import (
	"fmt"
	"sync"
	"strconv"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

// Service is the interface that provides booking methods.
type Service interface {
	// Insert a new job
	InsertJob(job wttypes.Job) (string, error)

	// Update a Job
	UpdateJob(job wttypes.Job) error

	// Get information about a particular job
	GetJob(jobID string) (wttypes.Job, error)

	// List all jobs in DB
	ListJobs() []wttypes.Job
}

type service struct {
	mtx    sync.RWMutex
	nextID string
	m      map[string]wttypes.Job
}

func (s * service) InsertJob (job wttypes.Job) (string, error) {
	// TODO change to MongoDB
	s.mtx.Lock()
	defer s.mtx.Unlock()

	job.ID = s.nextID

	// Get next ID
	i, _ := strconv.Atoi(s.nextID)
	i = i + 1
	s.nextID = strconv.Itoa(i)

	s.m[job.ID] = job

	return job.ID, nil
}

func (s * service) UpdateJob (job wttypes.Job) error {
	// TODO change to MongoDB
	s.mtx.Lock()
	defer s.mtx.Unlock()

	_, ok := s.m[job.ID]
	if !ok {
		return wttypes.ErrNotFound
	}

	s.m[job.ID] = job

	return nil
}

func (s * service) GetJob(jobID string) (wttypes.Job, error) {
	// TODO change to MongoDB
	s.mtx.Lock()
	defer s.mtx.Unlock()

	fmt.Println("id ", jobID)

	job, ok := s.m[jobID]
	if !ok {
		fmt.Println("error in getJob!!  ", jobID)
		return wttypes.Job{}, wttypes.ErrNotFound
	}

	return job, nil
}

func (s *service) ListJobs() []wttypes.Job {
	// TODO change to MongoDB
	values := []wttypes.Job{}

	for _, v := range s.m {
		values = append(values, v)
	}

	return values
}

// NewService creates a database service with necessary dependencies.
func NewService() Service {
	return &service{
		m: map[string]wttypes.Job{"1": wttypes.Job{ID:"1"}},
		nextID: "1",
	}
}
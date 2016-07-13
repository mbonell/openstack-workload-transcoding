package jobs

import (
	"time"
	"errors"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

// Service is the interface that provides jobs methods.
type Service interface {

	// Add a new job for transcoding
	AddNewJob(j wttypes.Job) (wttypes.JobID, error)

	// Get status for a particular job
	GetJobStatus(jobID wttypes.JobID) (string, error)

	// Cancel a job and all its transcoding
	CancelJob(jobID wttypes.JobID) (wttypes.Job, error)

	// Update the status of a transcoding
	UpdateTranscodingStatus(jobID wttypes.JobID, ttID wttypes.TranscodingTargetID, status string, objectname string) error
}

type service struct {

}

func (s *service) AddNewJob(j wttypes.Job) (wttypes.JobID, error) {

}
package jobs

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	"github.com/go-resty/resty"

	"github.com/rackspace/gophercloud"

	"github.com/obazavil/openstack-workload-transcoding/wtcommon"
	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

// Service is the interface that provides jobs methods.
type Service interface {
	// Add a new job for transcoding
	AddNewJob(job wttypes.Job) (string, error)

	// Get status for a particular job
	GetJobStatus(jobID string) (string, error)

	// Cancel a job and all its transcoding
	CancelJob(jobID string) error

	// Update the status of a transcoding
	UpdateTranscodingStatus(id string, status string, objectname string) error
}

type service struct {
	provider             *gophercloud.ProviderClient
	serviceObjectStorage *gophercloud.ServiceClient
}

func (s *service) AddNewJob(job wttypes.Job) (string, error) {
	// Verify we have transcodings to perform
	if len(job.Transcodings) == 0 {
		return "", wttypes.ErrNoTranscodings
	}

	//First let's upload to Object Storage
	objectname, errOS := wtcommon.Upload2ObjectStorage(s.serviceObjectStorage, job.URLMedia, job.VideoName)
	if errOS == nil {
		job.ObjectName = objectname
		job.Status = wttypes.JOB_QUEUED
	} else {
		job.Status = wttypes.JOB_ERROR
	}

	// Ask DB to add job into DB (even with error, for logging purposes)
	resp, err := resty.R().
		SetBody(job).
		Post(wtcommon.Servers["database"] + "/jobs")

	// Error in communication
	if err != nil {
		return "", err
	}

	str := resp.String()

	// There was an error in the response?
	if strings.HasPrefix(str, `{"error"`) {
		return "", wtcommon.JSON2Err(str)
	}

	// Get ID
	id, err := wtcommon.JSON2JobID(str)
	if err != nil {
		return id, err
	}

	// If job is in ERROR status, let's notify error even if everything else was OK
	if job.Status == wttypes.JOB_ERROR {
		return id, errors.New(wttypes.ErrCantUploadObject.Error() + ": " + errOS.Error())
	}

	// Let's send all transcodings tasks to Transcoding Manager
	for _, v := range job.Transcodings {
		resp, err := resty.R().
			SetBody(v).
			Post(wtcommon.Servers["manager"] + "/tasks")

		if err != nil {
			//TODO: do something when status update fails
		}

		str := resp.String()
		if strings.HasPrefix(str, `{"error"`) {
			//TODO: do something when status update fails
		}
	}

	return id, nil
}

func (s *service) GetJobStatus(jobID string) (string, error) {
	// Ask DB to get job from DB
	resp, err := resty.R().Get(wtcommon.Servers["database"] + "/jobs/" + jobID)

	// Error in communication
	if err != nil {
		return "", err
	}

	str := resp.String()

	// There was an error in the response?
	if strings.HasPrefix(str, `{"error"`) {
		return "", wtcommon.JSON2Err(str)

	}

	// Get job
	job, err := wtcommon.JSON2Job(str)
	if err != nil {
		return "", err
	}

	return job.Status, err
}

func (s *service) CancelJob(jobID string) error {
	// Ask DB to get job from DB
	resp, err := resty.R().Get(wtcommon.Servers["database"] + "/jobs/" + jobID)

	// Error in communication
	if err != nil {
		return err
	}

	str := resp.String()

	// There was an error in the response?
	if strings.HasPrefix(str, `{"error"`) {
		return wtcommon.JSON2Err(str)

	}

	// Get job
	job, err := wtcommon.JSON2Job(str)
	if err != nil {
		return err
	}

	// Cancel job
	//TODO: Cancel workers, etc.
	job.Status = wttypes.JOB_CANCELLED

	// Update DB
	resp, err = resty.R().
		SetBody(job).
		Put(wtcommon.Servers["database"] + "/jobs/" + jobID)

	// Error in communication
	if err != nil {
		return err
	}

	str = resp.String()

	// There was an error in the response?
	if strings.HasPrefix(str, `{"error"`) {
		return wtcommon.JSON2Err(str)

	}

	fmt.Println("[jobs]", "cancelled without any problem")

	return nil
}

func (s *service) UpdateTranscodingStatus(id string, status string, objectname string) error {
	// Ask DB to get transcoding from DB
	resp, err := resty.R().Get(wtcommon.Servers["database"] + "/transcodings/" + id)

	// Error in communication
	if err != nil {
		return err
	}

	str := resp.String()

	// There was an error in the response?
	if strings.HasPrefix(str, `{"error"`) {
		return wtcommon.JSON2Err(str)

	}

	// Get transcoding
	t, err := wtcommon.JSON2Transcoding(str)
	if err != nil {
		return err
	}

	//Update fields
	t.Status = status
	if status == wttypes.TRANSCODING_FINISHED && objectname != "" {
		t.ObjectName = objectname
	}

	// Update DB
	resp, err = resty.R().
		SetBody(t).
		Put(wtcommon.Servers["database"] + "/transcodings/" + id)

	// Error in communication
	if err != nil {
		return err
	}

	str = resp.String()

	// There was an error in the response?
	if strings.HasPrefix(str, `{"error"`) {
		return wtcommon.JSON2Err(str)

	}

	return nil
}

// NewService creates a jobs service with necessary dependencies.
func NewService() (Service, error) {
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	provider, err := wtcommon.GetProvider()
	if err != nil {
		return &service{}, err
	}

	serviceObjectStorage, err := wtcommon.GetServiceObjectStorage(provider)
	if err != nil {
		return &service{}, err
	}

	return &service{
		provider:             provider,
		serviceObjectStorage: serviceObjectStorage,
	}, nil
}

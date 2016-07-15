package jobs

import (
	"fmt"
	"strings"
	"crypto/tls"

	"github.com/go-resty/resty"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
	"github.com/obazavil/openstack-workload-transcoding/wtcommon"
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
	UpdateTranscodingStatus(jobID string, ttID string, status string, objectname string) error
}

type service struct {
}

func (s *service) AddNewJob(job wttypes.Job) (string, error) {
	fmt.Println("[jobs]", "AddNewJob")

	fmt.Println("[jobs]", "calling REST in DB service")

	// Ask DB to add job into DB
	resp, err := resty.R().
		SetBody(job).
		Post(wtcommon.Servers["database"] + "/v1/jobs")

	fmt.Println("[jobs]", "after REST in DB service")
	fmt.Println("[jobs]", "resp:", resp)

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
	fmt.Println("[jobs]", "id:", id)

	return id, err
}

func (s *service) GetJobStatus(jobID string) (string, error) {
	fmt.Println("[jobs]", "GetJobStatus")

	fmt.Println("[jobs]", "calling REST in DB service")

	// Ask DB to get job from DB
	resp, err := resty.R().Get(wtcommon.Servers["database"] + "/v1/jobs/" + jobID)

	fmt.Println("[jobs]", "after REST in DB service")
	fmt.Println("[jobs]", "resp:", resp)

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

	fmt.Println("[jobs]", "job:", job)

	return job.Status, err
}

func (s *service) CancelJob(jobID string) error {
	fmt.Println("[jobs]", "CancelJob")

	fmt.Println("[jobs]", "calling REST in DB service")

	// Ask DB to get job from DB
	resp, err := resty.R().Get(wtcommon.Servers["database"] + "/v1/jobs/" + jobID)

	fmt.Println("[jobs]", "after REST in DB service")
	fmt.Println("[jobs]", "resp:", resp)

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
		Put(wtcommon.Servers["database"] + "/v1/jobs/" + jobID)

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

func (s *service) UpdateTranscodingStatus(jobID string, ttID string, status string, objectname string) error {
	fmt.Println("[jobs]", "UpdateTranscodingStatus")
	fmt.Println("[jobs]", jobID, ttID, status, objectname)


	fmt.Println("[jobs]", "calling REST in DB service")

	// Ask DB to get job from DB
	resp, err := resty.R().Get(wtcommon.Servers["database"] + "/v1/jobs/" + jobID)

	fmt.Println("[jobs]", "after REST in DB service")
	fmt.Println("[jobs]", "resp:", resp)

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

	// Look for the Target Transcoding ID
	//TODO: another table? for now is same object

	found := false
	for k, v := range job.TranscodingTargets {
		if v.ID == ttID {
			v.Status = status

			// Update objectname only when specified
			if objectname != "" {
				v.ObjectName = objectname
			}

			job.TranscodingTargets[k] = v
			found = true
			break
		}
	}

	if !found {
		return wttypes.ErrTranscodingNotFound
	}

	fmt.Println("job to be sent ttStatus: ", job)

	// Update DB
	resp, err = resty.R().
		SetBody(job).
		Put(wtcommon.Servers["database"] + "/v1/jobs/" + jobID)

	// Error in communication
	if err != nil {
		return err
	}

	str = resp.String()

	// There was an error in the response?
	if strings.HasPrefix(str, `{"error"`) {
		return wtcommon.JSON2Err(str)

	}

	fmt.Println("[jobs]", "transcoding status updated without any problem")

	return nil
}


// NewService creates a jobs service with necessary dependencies.
func NewService() Service {
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	return &service{}
}

package manager

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	"github.com/go-resty/resty"
	"gopkg.in/mgo.v2"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

// Service is the interface that provides transcoding manager methods.
type Service interface {
	// Add a new transcoding task
	AddTranscoding(id string, objectname string, profile string) error

	// Cancel a transcoding task
	CancelTranscoding(id string) error

	// Get total of queued tasks
	GetTotalTasksQueued() (int, error)

	// Get total of active tasks
	GetTotalTasksRunning() (int, error)

	// Get next Transcoding task
	GetNextTask(workerAddr string) (wttypes.TranscodingTask, error)

	// Update the status of a task
	UpdateTaskStatus(id string, status string) error
}

type service struct {
	session *mgo.Session
}

func (s *service) AddTranscoding(id string, objectname string, profile string) error {
	// Add task
	task := wttypes.TranscodingTask{
		ID:         id,
		ObjectName: objectname,
		Profile:    profile,
	}

	datastore := NewDataStore(s.session)
	defer datastore.Close()

	id, err := datastore.AddTask(task)
	if err != nil {
		return err
	}

	fmt.Println("[manager] added transcoding:", id, profile, objectname)

	return nil
}

func (s *service) GetTotalTasksQueued() (int, error) {
	datastore := NewDataStore(s.session)
	defer datastore.Close()

	total, err := datastore.GetTotalTasksQueued()

	return total, err
}

func (s *service) GetTotalTasksRunning() (int, error) {
	datastore := NewDataStore(s.session)
	defer datastore.Close()

	total, err := datastore.GetTotalTasksRunning()

	return total, err
}

func (s *service) GetNextTask(workerAddr string) (wttypes.TranscodingTask, error) {
	datastore := NewDataStore(s.session)
	defer datastore.Close()

	task, err := datastore.GetNextQueuedTask(workerAddr)
	if err != nil {
		return wttypes.TranscodingTask{}, err
	}

	return task, err
}

func (s *service) UpdateTaskStatus(id string, status string) error {
	datastore := NewDataStore(s.session)
	defer datastore.Close()

	err := datastore.UpdateTaskStatus(id, status)
	if err != nil {
		return err
	}

	//// Update Job Service
	//body := struct {
	//	Status string `json:"status"`
	//}{
	//	Status: status,
	//}
	//
	//resp, err := resty.R().
	//	SetBody(body).
	//	Put(fmt.Sprintf("%s/transcodings/%s/status",
	//	wtcommon.Servers["jobs"],
	//	id))
	//
	//if err != nil {
	//	//TODO: do something when status update fails
	//}
	//
	//str := resp.String()
	//if strings.HasPrefix(str, `{"error"`) {
	//	//TODO: do something when status update fails
	//}

	return nil
}

func (s *service) CancelTranscoding(id string) error {
	fmt.Println("received cancel request for:", id)
	datastore := NewDataStore(s.session)
	defer datastore.Close()

	addr, err := datastore.CancelTask(id)
	if err != nil {
		return err
	}

	// If addr is not "", let's ask worker to cancel
	if addr != "" {
		fmt.Println("asking worker for cancellation:", addr)
		fmt.Println("url:", addr+":8083/tasks")
		resp, err := resty.R().
			Delete(addr + ":8083/tasks")

		if err != nil {
			//TODO: do something when cancel fails
		}

		str := resp.String()
		if strings.HasPrefix(str, `{"error"`) {
			//TODO: do something when cancel fails
		}
	}

	return nil
}

// NewService creates a transcoding manager service with necessary dependencies.
func NewService() (Service, error) {
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	s, err := CreateMongoSession()
	if err != nil {
		return &service{}, errors.New("[MongoDB] " + err.Error())
	}

	return &service{
		session: s,
	}, nil
}

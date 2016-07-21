package manager

import (
	"fmt"
	"crypto/tls"

	"github.com/go-resty/resty"

	"gopkg.in/mgo.v2"
	"errors"
)

// Service is the interface that provides transcoding manager methods.
type Service interface {
	// Add a new transcoding task
	AddTranscoding(id string, objectname string, profile string) error

	//// Cancel a transcoding task
	//CancelTranscoding(ttID string) (error)

	// Get total of queued tasks
	GetTotalTasksQueued() (int, error)

	// Get total of active tasks
	GetTotalTasksRunning() (int, error)

	// Get next Transcoding task
	GetNextTask(workerAddr string) (TranscodingTask, error)
}

type service struct {
	session *mgo.Session
}

func (s *service) AddTranscoding(id string, objectname string, profile string) error {
	fmt.Println("[manager]", "AddNewTranscoding start...")
	task := TranscodingTask{
		ID: id,
		ObjectName: objectname,
		Profile: profile,
	}

	datastore := NewDataStore(s.session)
	defer datastore.Close()

	id, err := datastore.AddTask(task)
	if err != nil {
		return err
	}

	fmt.Println("[manager]", "added transcoding :", id)

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

func (s * service) GetNextTask(workerAddr string) (TranscodingTask, error) {
	datastore := NewDataStore(s.session)
	defer datastore.Close()

	task, err := datastore.GetNextQueuedTask(workerAddr)

	return task, err
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
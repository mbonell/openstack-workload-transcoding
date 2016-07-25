package database

import (
	"errors"
	"crypto/tls"

	"github.com/go-resty/resty"

	"gopkg.in/mgo.v2"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

// Service is the interface that provides booking methods.
type Service interface {
	// Insert a new job into DB
	InsertJob(job wttypes.Job) (string, error)

	//// Update a Job in DB
	//UpdateJob(job wttypes.Job) error

	// Get information from DB about a particular job
	GetJob(id string) (wttypes.Job, error)

	// List all jobs in DB
	ListJobs() ([]wttypes.Job, error)

	//Update a transcoding in DB
	UpdateTranscoding(t wttypes.TranscodingTask) error

	//Get a transcoding from DB
	GetTranscoding(id string) (wttypes.TranscodingTask, error)
}

type service struct {
	session *mgo.Session
}

func (s *service) InsertJob(job wttypes.Job) (string, error) {

	datastore := NewDataStore(s.session)
	defer datastore.Close()

	id, err := datastore.InsertJob(job)

	return id, err
}

//func (s * service) UpdateJob (job wttypes.Job) error {
//	// TODO change to MongoDB
//
//	fmt.Println("updating job", job.ID)
//
//	//_, ok := s.m[job.ID]
//	//if !ok {
//	//	return wttypes.ErrNotFound
//	//}
//	//
//	//s.m[job.ID] = job
//
//	fmt.Println("job updated in memory:", job)
//
//	return nil
//}

func (s *service) GetJob(id string) (wttypes.Job, error) {
	datastore := NewDataStore(s.session)
	defer datastore.Close()

	job, err := datastore.GetJob(id)

	return job, err
}

func (s *service) ListJobs() ([]wttypes.Job, error) {
	datastore := NewDataStore(s.session)
	defer datastore.Close()

	jobs, err := datastore.ListJobs()

	return jobs, err
}

func (s *service) UpdateTranscoding(t wttypes.TranscodingTask) error {
	datastore := NewDataStore(s.session)
	defer datastore.Close()

	err := datastore.UpdateTranscoding(t)

	return err
}

//Get a transcoding from DB
func (s *service) GetTranscoding(id string) (wttypes.TranscodingTask, error) {
	datastore := NewDataStore(s.session)
	defer datastore.Close()

	t, err := datastore.GetTranscoding(id)

	return t, err

}


// NewService creates a database service with necessary dependencies.
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

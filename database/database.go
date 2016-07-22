package database

import (
	"fmt"
	"time"
	"errors"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

const (
	MongoDB                     = "transcoding"
	MongoJobsCollection         = "jobs"
	MongoTranscodingsCollection = "transcodings"
)

type JobDB struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	URLMedia   string        `bson:"url_media"`
	VideoName  string        `bson:"video_name"`
	ObjectName string        `bson:"object_name"`
	Added      time.Time     `bson:"added"`
	Started    time.Time     `bson:"started"`
	Ended      time.Time     `bson:"ended"`
	Status     string        `bson:"status"`
}

type TranscodingProfileDB struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	JobID      bson.ObjectId `bson:"job_id,omitempty"`
	Profile    string        `bson:"profile"`
	ObjectName string        `bson:"object_name"`
	Added      time.Time     `bson:"added"`
	Started    time.Time     `bson:"started"`
	Ended      time.Time     `bson:"ended"`
	Status     string        `bson:"status"`
}

type DataStore struct {
	session *mgo.Session
}

func (ds *DataStore) Close() {
	ds.session.Close()
}

func NewDataStore(s *mgo.Session) *DataStore {
	ds := &DataStore{
		session: s.Copy(),
	}
	return ds
}

func CreateMongoSession() (*mgo.Session, error) {
	// Dial to DB
	session, err := mgo.Dial("localhost")
	if err != nil {
		return nil, err
	}

	// Session is monotonic
	session.SetMode(mgo.Monotonic, true)

	// Get "jobs" collection
	c := session.DB(MongoDB).C(MongoJobsCollection)

	// Indexes
	idxID := mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(idxID)
	if err != nil {
		return nil, err
	}

	idxStatus := mgo.Index{
		Key:        []string{"status"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(idxStatus)
	if err != nil {
		return nil, err
	}

	// Get "transcodings" collection
	c = session.DB(MongoDB).C(MongoTranscodingsCollection)

	// Indexes
	idxID = mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(idxID)
	if err != nil {
		return nil, err
	}

	idxJob := mgo.Index{
		Key:        []string{"job_id"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(idxJob)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (ds *DataStore) ListJobs() ([]wttypes.Job, error) {
	// Get "jobs" collection
	c := ds.session.DB(MongoDB).C(MongoJobsCollection)

	// Get "transcodings" collection
	ct := ds.session.DB(MongoDB).C(MongoTranscodingsCollection)

	// Get all Jons, sort by added
	results := []JobDB{}
	err := c.Find(nil).Sort("added").All(&results)
	if err != nil {
		return []wttypes.Job{}, err
	}

	fmt.Println("ListJobs total: ", len(results))

	jobs := []wttypes.Job{}
	for _, v := range results {
		job := wttypes.Job{
			ID:         v.ID.Hex(),
			URLMedia:   v.URLMedia,
			VideoName:  v.VideoName,
			ObjectName: v.ObjectName,
			Status:     v.Status,
		}

		// Query for this job transcodings
		var resultsT []TranscodingProfileDB
		err = ct.Find(bson.M{"job_id": v.ID}).All(&resultsT)
		if err != nil {
			return []wttypes.Job{}, err
		}

		//Transcodings
		transcodings := []wttypes.TranscodingProfile{}
		for _, vt := range resultsT {
			t := wttypes.TranscodingProfile{
				ID:         vt.ID.Hex(),
				Profile:    vt.Profile,
				ObjectName: vt.ObjectName,
				Status:     vt.Status,
			}
			transcodings = append(transcodings, t)
		}
		job.Transcodings = transcodings

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (ds *DataStore) InsertJob(job wttypes.Job) (string, error) {
	jid := bson.NewObjectId()

	// Create JobDB object
	j := JobDB{
		ID:         jid,
		URLMedia:   job.URLMedia,
		VideoName:  job.VideoName,
		ObjectName: job.ObjectName,
		Added:      time.Now(),
		Status:     job.Status,
	}

	// Get "jobs" collection
	c := ds.session.DB(MongoDB).C(MongoJobsCollection)

	// Insert Job
	err := c.Insert(&j)
	if err != nil {
		return "", err
	}

	// Get "transcodings" collection
	c = ds.session.DB(MongoDB).C(MongoTranscodingsCollection)

	// Insert transcodings
	for _, v := range job.Transcodings {
		var tstatus string

		if j.Status == wttypes.JOB_QUEUED {
			tstatus = wttypes.TRANSCODING_QUEUED
		} else {
			tstatus = wttypes.TRANSCODING_SKIPPED
		}

		t := TranscodingProfileDB{
			ID:         bson.NewObjectId(),
			JobID:      jid,
			Profile:    v.Profile,
			ObjectName: v.ObjectName,
			Added:      time.Now(),
			Status:     tstatus,
		}

		// Insert Transcoding
		err := c.Insert(&t)
		if err != nil {
			return "", err
		}
	}

	fmt.Println("inserted with id:", jid.Hex())

	return jid.Hex(), nil
}

func (ds *DataStore) GetJob(id string) (wttypes.Job, error) {
	// Check is a valid ID
	if !bson.IsObjectIdHex(id) {
		return wttypes.Job{}, errors.New("Invalid ID")
	}

	// Get "jobs" collection
	c := ds.session.DB(MongoDB).C(MongoJobsCollection)

	// Query for job
	jid := bson.ObjectIdHex(id)

	result := JobDB{}
	err := c.FindId(jid).One(&result)
	if err != nil {
		return wttypes.Job{}, err
	}

	job := wttypes.Job{
		ID:         result.ID.Hex(),
		URLMedia:   result.URLMedia,
		VideoName:  result.VideoName,
		ObjectName: result.ObjectName,
		Status:     result.Status,
	}

	// Get "transcodings" collection
	c = ds.session.DB(MongoDB).C(MongoTranscodingsCollection)

	// Query for transcodings
	var results []TranscodingProfileDB
	err = c.Find(bson.M{"job_id": jid}).All(&results)
	if err != nil {
		return wttypes.Job{}, err
	}

	//Transcodings
	transcodings := []wttypes.TranscodingProfile{}
	for _, v := range results {
		t := wttypes.TranscodingProfile{
			ID:         v.ID.Hex(),
			Profile:    v.Profile,
			ObjectName: v.ObjectName,
			Status:     v.Status,
		}
		transcodings = append(transcodings, t)
	}
	job.Transcodings = transcodings

	return job, nil
}

func (ds *DataStore) GetTranscoding(id string) (wttypes.TranscodingProfile, error) {
	// Check is a valid ID
	if !bson.IsObjectIdHex(id) {
		return wttypes.TranscodingProfile{}, errors.New("Invalid ID")
	}

	// Get "transcodings" collection
	c := ds.session.DB(MongoDB).C(MongoTranscodingsCollection)

	// Query for transcoding
	tid := bson.ObjectIdHex(id)

	result := TranscodingProfileDB{}
	err := c.FindId(tid).One(&result)
	if err != nil {
		return wttypes.TranscodingProfile{}, err
	}

	t := wttypes.TranscodingProfile{
		ID: result.ID.Hex(),
		Profile: result.Profile,
		ObjectName: result.ObjectName,
		Status: result.Status,
	}


	return t, nil
}

func (ds *DataStore) UpdateTranscoding(t wttypes.TranscodingProfile) error {
	// Check is a valid ID
	if !bson.IsObjectIdHex(t.ID) {
		return errors.New("Invalid ID")
	}

	// Get "transcodings" collection
	c := ds.session.DB(MongoDB).C(MongoTranscodingsCollection)

	// Query for transcoding
	tid := bson.ObjectIdHex(t.ID)

	oldt := TranscodingProfileDB{}
	err := c.FindId(tid).One(&oldt)
	if err != nil {
		return err
	}

	// Update document
	newt := TranscodingProfileDB{
		ID: tid,

		Profile:    t.Profile,
		ObjectName: t.ObjectName,
		Status:     t.Status,

		JobID:   oldt.JobID,
		Added:   oldt.Added,
		Started: oldt.Started,
		Ended:   oldt.Ended,
	}

	// Update in DB
	_, err = c.UpsertId(tid, newt)
	if err != nil {
		return err
	}

	return nil
}



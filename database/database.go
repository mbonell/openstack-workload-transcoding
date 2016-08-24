package database

import (
	"errors"
	"time"
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

const (
	MongoDB                      = "transcoding"
	MongoJobsCollection          = "jobs"
	MongoTranscodingsCollection  = "transcodings"
	MongoWorkersEventsCollection = "metrics_workers_events"
	MongoWorkersCollection       = "workers"
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
	ID         bson.ObjectId `bson:"_id"`
	JobID      string        `bson:"job_id"`
	Profile    string        `bson:"profile"`
	ObjectName string        `bson:"object_name"`
	Added      time.Time     `bson:"added"`
	Started    time.Time     `bson:"started"`
	Ended      time.Time     `bson:"ended"`
	Status     string        `bson:"status"`
}

type WorkerEventDB struct {
	Timestamp time.Time `bson:"timestamp"`
	Time      string    `bson:"time"`
	Addr      string    `bson:"addr"`
	Event     string    `bson:"event"`
}

type WorkerDB struct {
	Addr        string    `bson:"addr"`
	LastUpdated time.Time `bson:"last_updated"`
	Status      string    `bson:"status"`
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

	idxJobStatus := mgo.Index{
		Key:        []string{"job_id", "status"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(idxJobStatus)
	if err != nil {
		return nil, err
	}

	// Get "events" collection
	c = session.DB(MongoDB).C(MongoWorkersEventsCollection)

	// Indexes
	idxEvent := mgo.Index{
		Key:        []string{"event"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(idxEvent)
	if err != nil {
		return nil, err
	}

	idxEventTime := mgo.Index{
		Key:        []string{"event", "hour", "minute"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(idxEventTime)
	if err != nil {
		return nil, err
	}

	idxEventHour := mgo.Index{
		Key:        []string{"event", "hour"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(idxEventHour)
	if err != nil {
		return nil, err
	}

	// Get "workers" collection
	c = session.DB(MongoDB).C(MongoWorkersCollection)

	// Indexes
	idxWorkerAddr := mgo.Index{
		Key:        []string{"addr"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(idxWorkerAddr)
	if err != nil {
		return nil, err
	}

	idxWorkerStatus := mgo.Index{
		Key:        []string{"status"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(idxWorkerStatus)
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
		transcodings := []wttypes.TranscodingTask{}
		for _, vt := range resultsT {
			t := wttypes.TranscodingTask{
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

func (ds *DataStore) InsertJob(job wttypes.Job) (wttypes.JobIDs, error) {
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
		return wttypes.JobIDs{}, err
	}

	ids := wttypes.JobIDs{
		ID: jid.Hex(),
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

		tid := bson.NewObjectId()
		t := TranscodingProfileDB{
			ID:         tid,
			JobID:      jid.Hex(),
			Profile:    v.Profile,
			ObjectName: v.ObjectName,
			Added:      time.Now(),
			Status:     tstatus,
		}

		tt := wttypes.TranscodingTask{
			ID:      tid.Hex(),
			Profile: v.Profile,
		}

		// Insert Transcoding
		err := c.Insert(&t)
		if err != nil {
			return wttypes.JobIDs{}, err
		}

		ids.Transcodings = append(ids.Transcodings, tt)
	}

	fmt.Println("[database] ids on insert:", ids)
	return ids, nil
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
	err = c.Find(bson.M{"job_id": jid.Hex()}).All(&results)
	if err != nil {
		return wttypes.Job{}, err
	}

	//Transcodings
	transcodings := []wttypes.TranscodingTask{}
	for _, v := range results {
		t := wttypes.TranscodingTask{
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

func (ds *DataStore) GetTranscoding(id string) (wttypes.TranscodingTask, error) {
	// Check is a valid ID
	if !bson.IsObjectIdHex(id) {
		return wttypes.TranscodingTask{}, errors.New("Invalid ID")
	}

	// Get "transcodings" collection
	c := ds.session.DB(MongoDB).C(MongoTranscodingsCollection)

	// Query for transcoding
	tid := bson.ObjectIdHex(id)

	result := TranscodingProfileDB{}
	err := c.FindId(tid).One(&result)
	if err != nil {
		return wttypes.TranscodingTask{}, err
	}

	t := wttypes.TranscodingTask{
		ID:         result.ID.Hex(),
		Profile:    result.Profile,
		ObjectName: result.ObjectName,
		Status:     result.Status,
	}

	return t, nil
}

func (ds *DataStore) UpdateTranscoding(t wttypes.TranscodingTask) error {
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

	// Update Started if needed
	started := oldt.Started
	if t.Status == wttypes.TRANSCODING_RUNNING && oldt.Status == wttypes.TRANSCODING_QUEUED {
		started = time.Now()

		// Change job status to RUNNING if needed
		job, err := ds.GetJob(oldt.JobID)
		if err == nil && job.Status == wttypes.JOB_QUEUED {
			job.Status = wttypes.JOB_RUNNING
			ds.UpdateJob(job)
		}
	}

	// Update Ended if needed
	ended := oldt.Ended
	if t.Status == wttypes.TRANSCODING_FINISHED && oldt.Status == wttypes.TRANSCODING_RUNNING {
		ended = time.Now()
	}

	// Update document
	newt := TranscodingProfileDB{
		ID: tid,

		Profile:    t.Profile,
		ObjectName: t.ObjectName,
		Status:     t.Status,
		Started:    started,
		Ended:      ended,

		JobID: oldt.JobID,
		Added: oldt.Added,
	}

	// Update in DB
	_, err = c.UpsertId(tid, newt)
	if err != nil {
		return err
	}

	// If we finished, let's see if we can mark Job as FINISHED
	if t.Status == wttypes.TRANSCODING_FINISHED {
		// Look for pending transcodings of same job
		jid := bson.ObjectIdHex(oldt.JobID)

		var results []struct {
			Status string `bson:"status"`
		}
		condStatus := bson.M{"$in": []string{wttypes.TRANSCODING_QUEUED, wttypes.TRANSCODING_RUNNING}}
		err = c.Find(bson.M{"job_id": jid.Hex(), "status": condStatus}).Select(bson.M{"status": 1}).All(&results)
		if err != nil {
			return err
		}

		// Update job in case no more pending transcodings...
		if len(results) == 0 {
			fmt.Println("no more transcodings for this job!! let's see if we need to update it on DB:", jid.Hex())
			// Get "jobs" collection
			c = ds.session.DB(MongoDB).C(MongoJobsCollection)

			job := JobDB{}
			err := c.FindId(jid).One(&job)
			if err != nil {
				return err
			}

			// Only update if job was queued or running
			if job.Status == wttypes.JOB_QUEUED || job.Status == wttypes.JOB_RUNNING {
				fmt.Println("marking job as finished:", jid.Hex())
				job.Status = wttypes.JOB_FINISHED
				job.Ended = time.Now()

				// Update in DB
				_, err = c.UpsertId(jid, job)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (ds *DataStore) UpdateJob(job wttypes.Job) error {
	// Check is a valid ID
	if !bson.IsObjectIdHex(job.ID) {
		return errors.New("Invalid ID")
	}

	// Get "jobs" collection
	c := ds.session.DB(MongoDB).C(MongoJobsCollection)

	// Query for job
	jid := bson.ObjectIdHex(job.ID)

	oldj := JobDB{}
	err := c.FindId(jid).One(&oldj)
	if err != nil {
		return err
	}

	// Update Started if needed
	started := oldj.Started
	if job.Status == wttypes.JOB_RUNNING && oldj.Status == wttypes.JOB_QUEUED {
		started = time.Now()
	}

	// Update Ended if needed
	ended := oldj.Ended
	if job.Status == wttypes.JOB_FINISHED && oldj.Status == wttypes.JOB_RUNNING {
		ended = time.Now()
	}

	// Update document
	newj := JobDB{
		ID: jid,

		Started: started,
		Ended:   ended,
		Status:  job.Status,

		URLMedia:   oldj.URLMedia,
		VideoName:  oldj.VideoName,
		ObjectName: oldj.ObjectName,
		Added:      oldj.Added,
	}

	// Update in DB
	_, err = c.UpsertId(jid, newj)
	if err != nil {
		return err
	}

	return nil
}

func (ds *DataStore) AddWorkerEvent(addr string, event string) error {
	fmt.Println("AddWorkerEvent:", event)
	// Get "events" collection
	c := ds.session.DB(MongoDB).C(MongoWorkersEventsCollection)

	t := time.Now()

	e := WorkerEventDB{
		Timestamp: t,
		Time:      fmt.Sprintf("%02d%02d", t.Hour(), t.Minute()),
		Addr:      addr,
		Event:     event,
	}

	// Insert Event
	err := c.Insert(&e)

	return err
}

func (ds *DataStore) UpdateWorkerStatus(addr string, status string) error {
	fmt.Println("UpdateWorkerStatus:", addr, status)

	// Get "workers" collection
	c := ds.session.DB(MongoDB).C(MongoWorkersCollection)

	w := WorkerDB{
		Addr:        addr,
		LastUpdated: time.Now(),
		Status:      status,
	}

	// Update/Insert in DB
	_, err := c.Upsert(bson.M{"addr": addr}, w)
	if err != nil {
		return err
	}

	err = ds.AddWorkerEvent(addr, status)

	return err
}
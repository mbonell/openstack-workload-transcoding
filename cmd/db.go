package main

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/obazavil/openstack-workload-transcoding/database"
	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

func strInSlice(str string, list []string) bool {
	for _, v := range list {
		if str == v {
			return true
		}
	}
	return false
}

func main() {
	var dropDB = false

	// Dial to DB
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Session has monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	// Drop Database
	if dropDB {
		err = session.DB(database.MongoDB).DropDatabase()
		if err != nil {
			panic(err)
		}
		return
	}

	// Get "jobs" collection
	c := session.DB(database.MongoDB).C(database.MongoTranscodingsCollection)

	//// Indexes
	//idxID := mgo.Index{
	//	Key:        []string{"id"},
	//	Unique:     true,
	//	DropDups:   true,
	//	Background: true,
	//	Sparse:     true,
	//}
	//err = c.EnsureIndex(idxID)
	//if err != nil {
	//	panic(err)
	//}
	//
	//idxStatus := mgo.Index{
	//	Key:        []string{"status"},
	//	Unique:     false,
	//	DropDups:   false,
	//	Background: true,
	//	Sparse:     true,
	//}
	//err = c.EnsureIndex(idxStatus)
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println("sobrevivi")
	//
	//// Insert
	//t := database.TaskDB {
	//	JobID: "1",
	//	ObjectName: "my object",
	//	Profile: "myprofile:",
	//	WorkerAddr: "128.1.1.1",
	//	Start: time.Now(),
	//	Status: wttypes.TRANSCODING_QUEUED,
	//}
	//
	//// Add ID
	//t.ID = bson.NewObjectId()
	//
	//err = c.Insert(&t)
	//
	//if err != nil {
	//	panic(err)
	//}
	//
	// // Query
	//res := database.TaskDB{}
	//err = c.FindId(t.ID).One(&res)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("res: ", res)

	// Query All
	var results []struct {
		Status string `bson:"status"`
	}
	jid := bson.ObjectIdHex("5796d0425e82588884a9713e")

	condStatus := bson.M{"$in": []string{wttypes.TRANSCODING_QUEUED, wttypes.TRANSCODING_RUNNING}}
	fmt.Println(condStatus)
	err = c.Find(bson.M{"job_id": jid, "status": condStatus}).Select(bson.M{"status": 1}).All(&results)

	fmt.Println("len:", len(results), results)

	//// Search for pending transcodings
	//pipeline := bson.M{
	//	"job_id": jid,
	//	"$or": []interface{}{
	//		bson.M{"status": wttypes.TRANSCODING_QUEUED},
	//		bson.M{"status": wttypes.TRANSCODING_RUNNING},
	//	},
	//}
	//
	//var res []database.TranscodingProfileDB
	//pipe := c.Pipe(pipeline)
	//err = pipe.All(&res)
	//
	//if err != nil {
	//	fmt.Println("err2:", err)
	//}
	//
	//fmt.Println("len2:", len(res))
	//
	//fmt.Println("len2:", len(results))

	//var results []database.JobDB
	//err = c.Find(bson.M{"status": wttypes.JOB_QUEUED}).All(&results)

	// get next one queued
	//result := database.JobDB{}
	//err = c.Find(bson.M{"status": wttypes.JOB_QUEUED}).One(&result)
	//
	//if err != nil {
	//	panic(err)
	//}

}

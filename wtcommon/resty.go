// Useful for JSON->Go: https://mholt.github.io/json-to-go/

package wtcommon

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/obazavil/openstack-workload-transcoding/wttypes"
)

type JSONErr struct {
	Error string `json:"error"`
}

func JSON2Err(s string) error {
	var v JSONErr

	if err := json.NewDecoder(strings.NewReader(s)).Decode(&v); err != nil {
		return errors.New("Can't decode JSON: " + s)
	}

	return errors.New(v.Error)
}

type JSONJobIDs struct {
	JobIDs wttypes.JobIDs `json:"job_ids"`
}

func JSON2JobIDs(s string) (wttypes.JobIDs, error) {
	var v JSONJobIDs

	if err := json.NewDecoder(strings.NewReader(s)).Decode(&v); err != nil {
		return wttypes.JobIDs{}, errors.New("Can't decode JSON: " + s)
	}

	return v.JobIDs, nil
}

type JSONJob struct {
	Job wttypes.Job `json:"job"`
}

func JSON2Job(s string) (wttypes.Job, error) {
	var v JSONJob

	if err := json.NewDecoder(strings.NewReader(s)).Decode(&v); err != nil {
		return wttypes.Job{}, errors.New("Can't decode JSON: " + s)
	}

	return v.Job, nil
}

type JSONTranscoding struct {
	Transcoding wttypes.TranscodingTask `json:"transcoding"`
}

func JSON2Transcoding(s string) (wttypes.TranscodingTask, error) {
	var v JSONTranscoding

	if err := json.NewDecoder(strings.NewReader(s)).Decode(&v); err != nil {
		return wttypes.TranscodingTask{}, errors.New("Can't decode JSON: " + s)
	}

	return v.Transcoding, nil
}

type JSONTask struct {
	Task wttypes.TranscodingTask `json:"task"`
}

func JSON2Task(s string) (wttypes.TranscodingTask, error) {
	var v JSONTask

	if err := json.NewDecoder(strings.NewReader(s)).Decode(&v); err != nil {
		return wttypes.TranscodingTask{}, errors.New("Can't decode JSON: " + s)
	}

	return v.Task, nil
}

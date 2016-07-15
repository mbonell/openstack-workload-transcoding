package wtcommon

import (
	"errors"
	"strings"
	"encoding/json"
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

type JSONJobID struct {
	JobID string `json:"job_id"`
}

func JSON2JobID(s string) (string, error) {
	var v JSONJobID

	if err := json.NewDecoder(strings.NewReader(s)).Decode(&v); err != nil {
		return "", errors.New("Can't decode JSON: " + s)
	}

	return v.JobID, nil
}

type JSONJob struct {
	Job wttypes.Job `json:"job"`
}

func JSON2Job(s string) (wttypes.Job, error) {
	var v JSONJob

	if err := json.NewDecoder(strings.NewReader(s)).Decode(&v); err != nil {
		return v.Job, errors.New("Can't decode JSON: " + s)
	}

	return v.Job, nil
}


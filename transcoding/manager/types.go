package manager

import (

)

type TranscodingTask struct {
	ID         string `json:"id"`
	ObjectName string `json:"object_name"`
	Profile    string `json:"profile"`
}

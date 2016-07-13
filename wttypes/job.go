package wttypes

// Job is a struct that stores all needed information for the job
type Job struct {
	ID         		string			`json:"id"`
	URLMedia   		string			`json:"url_media"`
	VideoName  		string 			`json:"video_name"`
	ObjectName 		string			`json:"object_name"`
	TranscodingTargets    	[]TranscodingTarget	`json:"transcoding_targets"`
	Status 			string			`json:"status"`
}
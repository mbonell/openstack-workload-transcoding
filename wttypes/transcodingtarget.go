package wttypes

// TranscodingTarget is a struct with information regarding the transcoding
type TranscodingTarget struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DeviceProfile string `json:"device_profile"`
	ObjectName    string `json:"object_name"`
	Status        string `json:"status"`
}

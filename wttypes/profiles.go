package wttypes

const (
	TRANSCODING_QUEUED = "queued"
	TRANSCODING_RUNNING = "running"
	TRANSCODING_CANCELLED = "cancelled"
	TRANSCODING_FINISHED = "finished"
	TRANSCODING_ERRED = "erred"
)

// Profile is a struct that maps a name with ffmpeg arguments
type Profile struct {
	Name string
	Args string
}

// DeviceProfile is a struct that maps a name with
// a profile and a resolution for use in a specific device
type DeviceProfile struct {
	Name       string
	Profile    Profile
	Resolution string
}

// NewDeviceProfile returns a map with the supported devices for transcoding
func NewDeviceProfile() *map[string]DeviceProfile {
	// Profiles
	proBaseline := Profile{Name: "baseline", Args: "-movflags faststart -profile:v baseline -level 3.0"}
	proApple42 := Profile{Name: "apple-42", Args: "-profile:v high -level 4.2"}
	proApple41 := Profile{Name: "apple-41", Args: "-profile:v high -level 4.1"}

	// Profiles per device
	dp := make(map[string]DeviceProfile)

	// All supported devices
	dp["baseline"] = DeviceProfile{Name: "baseline", Profile: proBaseline}
	dp["iPhone4s"] = DeviceProfile{Name: "iPhone4s", Profile: proApple41, Resolution: "960x640"}
	dp["iPhone5s"] = DeviceProfile{Name: "iPhone5s", Profile: proApple42, Resolution: "1136x640"}
	dp["iPhonePlus6s"] = DeviceProfile{Name: "iPhonePlus6s", Profile: proApple42, Resolution: "1920x1080"}
	dp["iPadMini4"] = DeviceProfile{Name: "iPadMini4", Profile: proApple42, Resolution: "2048x1536"}

	return &dp
}

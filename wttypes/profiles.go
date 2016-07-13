package wttypes

type Profile struct {
	Name string
	Args string
}

type DeviceProfile struct {
	Profile    Profile
	Resolution string
}

func NewDeviceProfile() map[string]DeviceProfile {
	// Profiles
	proBaseline := Profile{Name: "baseline", Args: "-movflags faststart -profile:v baseline -level 3.0"}
	proApple42 := Profile{Name: "apple-42", Args: "-profile:v high -level 4.2"}
	proApple41 := Profile{Name: "apple-41", Args: "-profile:v high -level 4.1"}

	// Profiles per device
	dp := make(map[string]DeviceProfile)

	dp["baseline"] = DeviceProfile{Profile: proBaseline}
	dp["iPhone4s"] = DeviceProfile{Profile: proApple41, Resolution: "960x640"}
	dp["iPhone5s"] = DeviceProfile{Profile: proApple42, Resolution: "1136x640"}
	dp["iPhonePlus6s"] = DeviceProfile{Profile: proApple42, Resolution: "1920x1080"}
	dp["iPadMini4"] = DeviceProfile{Profile: proApple42, Resolution: "2048x1536"}

	return dp
}

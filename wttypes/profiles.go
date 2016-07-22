package wttypes

// ProfileFFMPEG is a struct that maps a name with ffmpeg arguments
type ProfileFFMPEG struct {
	Name string
	Args string
}

// Profile is a struct that maps a name with
// a FFMPEG profile and a resolution for use in a specific device or system
type Profile struct {
	Name       string
	FFMPEG     ProfileFFMPEG
	Resolution string
}

// NewProfile returns a map with the supported profiles for transcoding
func NewProfile() map[string]Profile {
	// Profiles FFMPEG
	proBaseline := ProfileFFMPEG{Name: "baseline", Args: "-movflags faststart -profile:v baseline -level 3.0"}
	proApple42 := ProfileFFMPEG{Name: "apple-42", Args: "-profile:v high -level 4.2"}
	proApple41 := ProfileFFMPEG{Name: "apple-41", Args: "-profile:v high -level 4.1"}

	// Profiles map
	p := make(map[string]Profile)

	// All supported profiles
	p["baseline"] = Profile{Name: "baseline", FFMPEG: proBaseline}
	p["iPhone4s"] = Profile{Name: "iPhone4s", FFMPEG: proApple41, Resolution: "960x640"}
	p["iPhone5s"] = Profile{Name: "iPhone5s", FFMPEG: proApple42, Resolution: "1136x640"}
	p["iPhonePlus6s"] = Profile{Name: "iPhonePlus6s", FFMPEG: proApple42, Resolution: "1920x1080"}
	p["iPadMini4"] = Profile{Name: "iPadMini4", FFMPEG: proApple42, Resolution: "2048x1536"}

	return p
}

package ofbx

// FrameRate documented here: http://docs.autodesk.com/FBX/2014/ENU/FBX-SDK-Documentation/index.html?url=cpp_ref/class_fbx_time.html,topicNumber=cpp_ref_class_fbx_time_html29087af6-8c2c-4e9d-aede-7dc5a1c2436c,hash=a837590fd5310ff5df56ffcf7c394787e
type FrameRate int

// FrameRate values
const (
	FrameRate_DEFAULT         FrameRate = iota
	FrameRate_120             FrameRate = iota
	FrameRate_100             FrameRate = iota
	FrameRate_60              FrameRate = iota
	FrameRate_50              FrameRate = iota
	FrameRate_48              FrameRate = iota
	FrameRate_30              FrameRate = iota
	FrameRate_30_DROP         FrameRate = iota
	FrameRate_NTSC_DROP_FRAME FrameRate = iota
	FrameRate_NTSC_FULL_FRAME FrameRate = iota
	FrameRate_PAL             FrameRate = iota
	FrameRate_CINEMA          FrameRate = iota
	FrameRate_1000            FrameRate = iota
	FrameRate_CINEMA_ND       FrameRate = iota
	FrameRate_CUSTOM          FrameRate = iota
)

// GetFramerateFromTimeMode gets time from a given framerate TODO: Confirm these
func GetFramerateFromTimeMode(f FrameRate, custom float32) float32 {
	switch f {
	case FrameRate_DEFAULT:
		return 1
	case FrameRate_120:
		return 120
	case FrameRate_100:
		return 100
	case FrameRate_60:
		return 60
	case FrameRate_50:
		return 50
	case FrameRate_48:
		return 48
	case FrameRate_30:
		return 30
	case FrameRate_30_DROP:
		return 30
	case FrameRate_NTSC_DROP_FRAME:
		return 29.9700262
	case FrameRate_NTSC_FULL_FRAME:
		return 29.9700262
	case FrameRate_PAL:
		return 25
	case FrameRate_CINEMA:
		return 24
	case FrameRate_1000:
		return 1000
	case FrameRate_CINEMA_ND:
		return 23.976
	case FrameRate_CUSTOM:
		return custom
	}
	return -1
}

func fbxTimeToSeconds(value int64) float64 {
	return float64(value) / float64(46186158000)
}
func secondsToFbxTime(value float64) int64 {
	return int64(value / 46186158000)
}

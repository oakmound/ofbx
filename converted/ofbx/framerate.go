package ofbx

// FrameRate documented here: http://docs.autodesk.com/FBX/2014/ENU/FBX-SDK-Documentation/index.html?url=cpp_ref/class_fbx_time.html,topicNumber=cpp_ref_class_fbx_time_html29087af6-8c2c-4e9d-aede-7dc5a1c2436c,hash=a837590fd5310ff5df56ffcf7c394787e
import (
	"time"
)

// FrameRate enumerates standard rates of how many frames should be advanced per second
type FrameRate int

// FrameRate values
const (
	FrameRateDefault       FrameRate = iota
	FrameRate120           FrameRate = iota
	FrameRate100           FrameRate = iota
	FrameRate60            FrameRate = iota
	FrameRate50            FrameRate = iota
	FrameRate48            FrameRate = iota
	FrameRate30            FrameRate = iota
	FrameRate30Drop        FrameRate = iota
	FrameRateNTSCDropFrame FrameRate = iota
	FrameRateNTSCFullFrame FrameRate = iota
	FrameRatePAL           FrameRate = iota
	FrameRateCinema        FrameRate = iota
	FrameRate1000          FrameRate = iota
	FrameRateCinemaND      FrameRate = iota
	FrameRateCustom        FrameRate = iota
)

// GetFramerateFromTimeMode gets time from a given framerate TODO: Confirm these
func GetFramerateFromTimeMode(f FrameRate, custom float32) float32 {
	switch f {
	case FrameRateDefault:
		return 1
	case FrameRate120:
		return 120
	case FrameRate100:
		return 100
	case FrameRate60:
		return 60
	case FrameRate50:
		return 50
	case FrameRate48:
		return 48
	case FrameRate30:
		return 30
	case FrameRate30Drop:
		return 30
	case FrameRateNTSCDropFrame:
		return 29.9700262
	case FrameRateNTSCFullFrame:
		return 29.9700262
	case FrameRatePAL:
		return 25
	case FrameRateCinema:
		return 24
	case FrameRate1000:
		return 1000
	case FrameRateCinemaND:
		return 23.976
	case FrameRateCustom:
		return custom
	}
	return -1
}

func fbxTimetoStdTime(value int64) time.Duration {
	return time.Second * time.Duration(value) / 46186158000
}

func fbxTimeToSeconds(value int64) float64 {
	return float64(value) / float64(46186158000)
}
func secondsToFbxTime(value float64) int64 {
	return int64(value / 46186158000)
}

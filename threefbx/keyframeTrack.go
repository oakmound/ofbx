package threefbx

type Track struct{} // ???
type KeyframeTrack struct {
	name          string
	times         []float64
	values        []float64 // as long as the times
	interpolation string
} // ???

func VectorKeyframeTrack(name string, times, values []float64) KeyframeTrack {

}

func QuaternionKeyframeTrack(name string, times, values []float64) KeyframeTrack {

}

func NumberKeyframeTrack(name string, times, values []float64) KeyframeTrack {

}

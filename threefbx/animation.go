package threefbx

type Animation struct {
}

type AnimationCurve struct {
	ID     int64
	Times  []float64
	Values []float64
}

func NewAnimationClip(name string, duration int, tracks []KeyframeTrack) Animation {
	return Animation{}
}

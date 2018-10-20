package threefbx

type Track struct{} // ???
type KeyframeTrack struct {
	name   string
	times  []float64
	values []float64 // as long as the times
	// interpolation string only need the default currently: InterpolateLinear
} // ???

func VectorKeyframeTrack(name string, times, values []float64) KeyframeTrack {
	return NumberKeyframeTrack(name, times, values)
}

func NumberKeyframeTrack(name string, times, values []float64) KeyframeTrack {
	kt := KeyframeTrack{
		name:   name,
		times:  times,
		values: values,
	}

	return kt
}

//TODO: figure out where the fact that this is quaterniion matters
func QuaternionKeyframeTrack(name string, times, values []float64) KeyframeTrack {
	kt := KeyframeTrack{
		name:   name,
		times:  times,
		values: values,
	}
	return kt
}

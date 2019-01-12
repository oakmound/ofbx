package threefbx

import "strings"

type KeyframeTrack struct {
	Name      string
	Operation string
	Times     []float64
	Values    []float64 // as long as the times
	// interpolation string; only need the default currently: InterpolateLinear
} // ???

func VectorKeyframeTrack(name string, times, values []float64) KeyframeTrack {
	return NumberKeyframeTrack(name, times, values)
}

func NumberKeyframeTrack(name string, times, values []float64) KeyframeTrack {
	kt := KeyframeTrack{
		Name:      name,
		Operation: getOperation(name),
		Times:     times,
		Values:    values,
	}

	return kt
}

//TODO: figure out where the fact that this is quaterniion matters
func QuaternionKeyframeTrack(name string, times, values []float64) KeyframeTrack {
	kt := KeyframeTrack{
		Name:      name,
		Operation: getOperation(name),
		Times:     times,
		Values:    values,
	}
	return kt
}

func getOperation(name string) string {
	splitName := strings.Split(name, ".")
	return splitName[len(splitName)-1]

}

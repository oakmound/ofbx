package ofbx

import (
	"strconv"

	"github.com/oakmound/oak/alg/floatgeom"
)

func intFromString(str, end string, val *int) string {
	v, err := strconv.Atoi(str)
	*val = v
	if err != nil {
		panic("Strconv Failed.")
	}
	iter := 0
	for iter < len(end) && str[iter] != ',' {
		iter++
	}
	if iter < len(end) {
		iter++
	}
	return str[iter:]
}

func uint64FromString(str, end string, val *uint64) string {

	v, err := strconv.ParseUint(str, 10, 64)
	*val = v
	if err != nil {
		panic("Strconv Failed.")
	}
	iter := 0
	for iter < len(end) && str[iter] != ',' {
		iter++
	}
	if iter < len(end) {
		iter++
	}
	return str[iter:]
}

func int64FromString(str, end string, val *int64) string {
	v, err := strconv.ParseInt(str, 10, 64)
	*val = v
	if err != nil {
		panic("Strconv Failed.")
	}
	iter := 0
	for iter < len(end) && str[iter] != ',' {
		iter++
	}
	if iter < len(end) {
		iter++
	}
	return str[iter:]
}

func doubleFromString(str, end string, val *float64) string {
	v, err := strconv.ParseFloat(str, 64)
	*val = v
	if err != nil {
		panic("Strconv Failed.")
	}
	iter := 0
	for iter < len(end) && str[iter] != ',' {
		iter++
	}
	if iter < len(end) {
		iter++
	}
	return str[iter:]
}

func floatFromString(str, end string, val *float32) string {
	v, err := strconv.ParseFloat(str, 32)
	*val = float32(v)
	if err != nil {
		panic("Strconv Failed.")
	}
	iter := 0
	for iter < len(end) && str[iter] != ',' {
		iter++
	}
	if iter < len(end) {
		iter++
	}
	return str[iter:]
}

func fromString(str, end string, val *float64, count int) string {
	iter := 0
	for i := 0; i < count; i++ {
		v, err := strconv.ParseFloat(str[iter:], 64)
		*val = v
		if err != nil {
			panic("Strconv Failed.")
		}
		iter := 0
		for iter < len(end) && str[iter] != ',' {
			iter++
		}
		if iter < len(end) {
			iter++
		}

		if iter == len(end) {
			return str[iter:]
		}
	}
	return str[iter:]
}

//Todo: Convert from using pointer math...
func Vec2FromString(str, end string, val *floatgeom.Point2) string {
	return fromString(str, end, &val.X, 2)
}

func Vec3FromString(str, end string, val *floatgeom.Point3) string {
	return fromString(str, end, &val.X, 3)
}

func Vec4FromString(str, end string, val *floatgeom.Point4) string {
	return fromString(str, end, &val.X, 4)
}

func matrixFromString(str, end string, val *Matrix) string {
	return fromString(str, end, &val.m[0], 16)
}

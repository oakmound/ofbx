package ofbx

import (
	"fmt"
	"math"

	"github.com/oakmound/oak/alg/floatgeom"
)

type UpVector int

// Specifies which canonical axis represents up in the system (typically Y or Z).
const (
	UpVector_AxisX UpVector = 1
	UpVector_AxisY UpVector = 2
	UpVector_AxisZ UpVector = 3
)

// Vector with origin at the screen pointing toward the camera.
type FrontVector int

const (
	FrontVector_ParityEven FrontVector = 1
	FrontVector_ParityOdd  FrontVector = 2
)

// Specifies the third vector of the system.
type CoordSystem int

const (
	CoordSystem_RightHanded CoordSystem = iota
	CoordSystem_LeftHanded  CoordSystem = iota
)

type Matrix struct {
	m [16]float64 // last 4 are translation
}

func matrixFromSlice(fs []float64) (Matrix, error) {
	if len(fs) != 16 {
		return Matrix{}, fmt.Errorf("Expected 16 values, got %d", len(fs))
	}
	var a [16]float64
	copy(a[:], fs)
	return Matrix{a}, nil
}

type Quat struct {
	X, Y, Z, w float64
}

func (m1 Matrix) Mul(m2 Matrix) Matrix {
	res := [16]float64{}
	for j := 0; j < 4; j++ {
		for i := 0; i < 4; i++ {
			tmp := 0.0
			for k := 0; k < 4; k++ {
				tmp += m1.m[i+k*4] * m2.m[k+j*4]
			}
			res[i+j*4] = tmp
		}
	}
	return Matrix{res}
}

func setTranslation(v floatgeom.Point3, m *Matrix) {
	m.m[12] = v.X()
	m.m[13] = v.Y()
	m.m[14] = v.Z()
}

func makeIdentity() Matrix {
	return Matrix{[16]float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}}
}

func rotationX(angle float64) Matrix {
	m2 := makeIdentity()
	//radian
	c := math.Cos(angle)
	s := math.Sin(angle)
	m2.m[5] = c
	m2.m[10] = c
	m2.m[9] = -s
	m2.m[6] = s
	return m2
}

func rotationY(angle float64) Matrix {
	m2 := makeIdentity()
	//radian
	c := math.Cos(angle)
	s := math.Sin(angle)
	m2.m[0] = c
	m2.m[10] = c
	m2.m[8] = s
	m2.m[2] = -s
	return m2
}

func rotationZ(angle float64) Matrix {
	m2 := makeIdentity()
	//radian
	c := math.Cos(angle)
	s := math.Sin(angle)
	m2.m[0] = c
	m2.m[5] = c
	m2.m[4] = -s
	m2.m[1] = s
	return m2
}

func getTriCountFromPoly(indices []int, idx int) (int, int) {
	count := 1
	for indices[idx+1+count] >= 0 {
		count++
	}
	return count, idx
}

func getRotationMatrix(euler *floatgeom.Point3, order RotationOrder) Matrix {
	TO_RAD := 3.1415926535897932384626433832795028 / 180.0 //TODO: Update this
	rx := rotationX(euler.X() * TO_RAD)
	ry := rotationY(euler.Y() * TO_RAD)
	rz := rotationZ(euler.Z() * TO_RAD)
	switch order {
	default:
	case SPHERIC_XYZ:
		panic("This should not happen")
	case EULER_XYZ:
		return rz.Mul(ry).Mul(rx)
	case EULER_XZY:
		return ry.Mul(rz).Mul(rx)
	case EULER_YXZ:
		return rz.Mul(rx).Mul(ry)
	case EULER_YZX:
		return rx.Mul(rz).Mul(ry)
	case EULER_ZXY:
		return ry.Mul(rx).Mul(rz)
	case EULER_ZYX:
		return rx.Mul(ry).Mul(rz)
	}
	panic("This shouldn't happen either")
}

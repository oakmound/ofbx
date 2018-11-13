package ofbx

import (
	"fmt"
	"math"

	"github.com/oakmound/oak/alg/floatgeom"
)

// UpVector specifies which canonical axis represents up in the system (typically Y or Z).
type UpVector int

// UpVector Options
const (
	UpVectorX UpVector = 1
	UpVectorY UpVector = 2
	UpVectorZ UpVector = 3
)

// FrontVector is a vector with origin at the screen pointing toward the camera.
type FrontVector int

// FrontVector Parity Options
const (
	FrontVectorParityEven FrontVector = 1
	FrontVectorParityOdd  FrontVector = 2
)

// CoordSystem specifies the third vector of the system.
type CoordSystem int

// CoordSystem options
const (
	CoordSystemRight CoordSystem = iota
	CoordSystemLeft  CoordSystem = iota
)

// Matrix is a 16 sized slice that we operate on as if it was actually a matrix
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

// Quat probably can bve removed
type Quat struct {
	X, Y, Z, w float64
}

// Mul multiplies the values of two matricies together and returns the output
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

// Transpose Gets returns the transpose of the given matrix
func (m1 Matrix) Transpose() Matrix {
	res := [16]float64{}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			res[i*4+j] = m1.m[j*4+i]
		}
	}
	return Matrix{res}
}

func (m1 Matrix) Inverse() Matrix {
	//TODO: we gotta implement
	return m1
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

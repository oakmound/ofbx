package threefbx

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/oakmound/oak/alg/floatgeom"
)

func Mat4FromSlice(fs []float64) mgl64.Mat4 {
	// This might be wrong - column major vs row major
	f16 := [16]float64{}
	copy(f16[:], fs)
	return mgl64.Mat4(f16)
}

func IsZeroMat(mat mgl64.Mat4) bool {
	return mat == mgl64.Mat4{}
}

// scaleMat4 covers the gap from three.js where you can scale a mat4 by a vec3 scale
func scaleMat4(m mgl64.Mat4, s floatgeom.Point3) mgl64.Mat4 {
	m[0] *= s.X()
	m[4] *= s.Y()
	m[8] *= s.Z()
	m[1] *= s.X()
	m[5] *= s.Y()
	m[9] *= s.Z()
	m[2] *= s.X()
	m[6] *= s.Y()
	m[10] *= s.Z()
	m[3] *= s.X()
	m[7] *= s.Y()
	m[11] *= s.Z()
	return m
}

// applyBufferAttribute so far has been used to apply pretransforms on a set of points
func applyBufferAttribute(m mgl64.Mat4, a []floatgeom.Point3) []floatgeom.Point3 {
	for i := 0; i < len(a); i++ {
		a[i] = applyMat4(m, a[i])
	}
	return a
}

func applyBufferAttributeMat3(m mgl64.Mat3, a []floatgeom.Point3) []floatgeom.Point3 {
	for i := 0; i < len(a); i++ {
		a[i] = applyMat3(m, a[i])
	}
	return a
}

// applyMat3 is puleld from three.js vector3 it applies the mat3 as a transformation
func applyMat3(m mgl64.Mat3, p floatgeom.Point3) floatgeom.Point3 {
	x := p.X()
	y := p.Y()
	z := p.Z()
	p[0] = m[0]*x + m[3]*y + m[6]*z
	p[1] = m[1]*x + m[4]*y + m[7]*z
	p[2] = m[2]*x + m[5]*y + m[8]*z
	return p
}

// applyMat4 is puleld from three.js vector3 it applies the mat4 as a transformation
func applyMat4(m mgl64.Mat4, p floatgeom.Point3) floatgeom.Point3 {
	x := p.X()
	y := p.Y()
	z := p.Z()
	w := 1 / (m[3]*x + m[7]*y + m[11]*z + m[15])

	p[0] = (m[0]*x + m[4]*y + m[8]*z + m[12]) * w
	p[1] = (m[1]*x + m[5]*y + m[9]*z + m[13]) * w
	p[2] = (m[2]*x + m[6]*y + m[10]*z + m[14]) * w
	return p
}

func decomposeMat(mat mgl64.Mat4) (floatgeom.Point3, Euler, floatgeom.Point3) {
	pos := floatgeom.Point3{mat[12], mat[13], mat[14]}
	scale := floatgeom.Point3{
		floatgeom.Point3{mat[0], mat[4], mat[8]}.Magnitude(),
		floatgeom.Point3{mat[1], mat[5], mat[9]}.Magnitude(),
		floatgeom.Point3{mat[2], mat[6], mat[10]}.Magnitude(),
	}
	rot := mgl64.Mat4{
		mat[0] / scale[0], mat[1] / scale[1], mat[2] / scale[2], 0,
		mat[4] / scale[0], mat[5] / scale[1], mat[6] / scale[2], 0,
		mat[8] / scale[0], mat[9] / scale[1], mat[10] / scale[2], 0,
		0, 0, 0, 1,
	}
	trace := rot[0] + rot[5] + rot[10] + 1
	var S, qw, qx, qy, qz float64
	if trace > 0 {
		S = math.Sqrt(trace+1.0) * 2 // S=4*qw
		qw = 0.25 * S
		qx = (rot[9] - rot[6]) / S
		qy = (rot[2] - rot[8]) / S
		qz = (rot[4] - rot[1]) / S
	} else if (rot[0] > rot[5]) && (rot[0] > rot[10]) {
		S = math.Sqrt(1.0+rot[0]-rot[5]-rot[10]) * 2 // S=4*qx
		qw = (rot[9] - rot[6]) / S
		qx = 0.25 * S
		qy = (rot[1] + rot[4]) / S
		qz = (rot[2] + rot[8]) / S
	} else if rot[5] > rot[10] {
		S = math.Sqrt(1.0+rot[5]-rot[0]-rot[10]) * 2 // S=4*qy
		qw = (rot[2] - rot[8]) / S
		qx = (rot[1] + rot[4]) / S
		qy = 0.25 * S
		qz = (rot[6] + rot[9]) / S
	} else {
		S = math.Sqrt(1.0+rot[10]-rot[0]-rot[5]) * 2 // S=4*qz
		qw = (rot[4] - rot[1]) / S
		qx = (rot[2] + rot[8]) / S
		qy = (rot[6] + rot[9]) / S
		qz = 0.25 * S
	}
	quat := floatgeom.Point4{qx, qy, qz, qw}
	// Assuming XYZ order used by wikipedia algorithm
	euler := Euler{
		floatgeom.Point3{
			math.Atan((2*(quat[0]*quat[1]+quat[2]*quat[3]))/1 - (2 * (quat[1]*quat[1] + quat[2]*quat[2]))),
			math.Asin(2 * (quat[0]*quat[2] - quat[3]*quat[1])),
			math.Atan((2*(quat[0]*quat[3]+quat[1]*quat[2]))/1 - (2 * (quat[2]*quat[2] + quat[3]*quat[3]))),
		},
		XYZOrder,
	}
	return pos, euler, scale
}

//composeMat https://github.com/mrdoob/three.js/blob/07b24f7f03e73174278152f062d98068124d6ff2/src/math/Matrix4.js#L741
func composeMat(position floatgeom.Point3, quaternion floatgeom.Point4, scale floatgeom.Point3) mgl64.Mat4 {
	te := mgl64.Mat4{}
	x := quaternion.X()
	y := quaternion.Y()
	z := quaternion.Z()
	w := quaternion.W()
	x2 := x + x
	y2 := y + y
	z2 := z + z
	xx := x * x2
	xy := x * y2
	xz := x * z2
	yy := y * y2
	yz := y * z2
	zz := z * z2
	wx := w * x2
	wy := w * y2
	wz := w * z2

	sx := scale.X()
	sy := scale.Y()
	sz := scale.Z()

	te[0] = (1 - (yy + zz)) * sx
	te[1] = (xy + wz) * sx
	te[2] = (xz - wy) * sx
	te[3] = 0

	te[4] = (xy - wz) * sy
	te[5] = (1 - (xx + zz)) * sy
	te[6] = (yz + wx) * sy
	te[7] = 0

	te[8] = (xz + wy) * sz
	te[9] = (yz - wx) * sz
	te[10] = (1 - (xx + yy)) * sz
	te[11] = 0

	te[12] = position.X()
	te[13] = position.Y()
	te[14] = position.Z()
	te[15] = 1

	return te
}

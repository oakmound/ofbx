package threefbx

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/oakmound/oak/alg/floatgeom"
)

type EulerOrder int

const (
	ZYXOrder       EulerOrder = iota // -> XYZ extrinsic
	YZXOrder       EulerOrder = iota // -> XZY extrinsic
	XZYOrder       EulerOrder = iota // -> YZX extrinsic
	ZXYOrder       EulerOrder = iota // -> YXZ extrinsic
	YXZOrder       EulerOrder = iota // -> ZXY extrinsic
	XYZOrder       EulerOrder = iota // -> ZYX extrinsic
	LastEulerOrder EulerOrder = iota
)

type Euler struct {
	floatgeom.Point3
	EulerOrder
}

// func (e Euler) ToQuaternion() floatgeom.Point4 {
// 	cy := math.Cos(e.Z() * 0.5)
// 	sy := math.Sin(e.Z() * 0.5)
// 	cr := math.Cos(e.Y() * 0.5)
// 	sr := math.Sin(e.Y() * 0.5)
// 	cp := math.Cos(e.X() * 0.5)
// 	sp := math.Sin(e.X() * 0.5)

// 	return floatgeom.Point4{
// 		cy*sr*cp - sy*cr*sp,
// 		cy*cr*sp + sy*sr*cp,
// 		sy*cr*cp - cy*sr*sp,
// 		cy*cr*cp + sy*sr*sp,
// 	}
// }

func (euler Euler) makeRotation() mgl64.Mat4 {
	te := mgl64.Mat4{}
	a := math.Cos(euler.X())
	b := math.Sin(euler.X())
	c := math.Cos(euler.Y())
	d := math.Sin(euler.Y())
	e := math.Cos(euler.Z())
	f := math.Sin(euler.Z())

	switch euler.EulerOrder {
	case XYZOrder:
		ae := a * e
		af := a * f
		be := b * e
		bf := b * f
		te[0] = c * e
		te[4] = -c * f
		te[8] = d
		te[1] = af + be*d
		te[5] = ae - bf*d
		te[9] = -b * c
		te[2] = bf - ae*d
		te[6] = be + af*d
		te[10] = a * c
	case YXZOrder:
		ce := c * e
		cf := c * f
		de := d * e
		df := d * f
		te[0] = ce + df*b
		te[4] = de*b - cf
		te[8] = a * d
		te[1] = a * f
		te[5] = a * e
		te[9] = -b
		te[2] = cf*b - de
		te[6] = df + ce*b
		te[10] = a * c
	case ZXYOrder:
		ce := c * e
		cf := c * f
		de := d * e
		df := d * f
		te[0] = ce - df*b
		te[4] = -a * f
		te[8] = de + cf*b
		te[1] = cf + de*b
		te[5] = a * e
		te[9] = df - ce*b
		te[2] = -a * d
		te[6] = b
		te[10] = a * c
	case ZYXOrder:
		ae := a * e
		af := a * f
		be := b * e
		bf := b * f

		te[0] = c * e
		te[4] = be*d - af
		te[8] = ae*d + bf

		te[1] = c * f
		te[5] = bf*d + ae
		te[9] = af*d - be

		te[2] = -d
		te[6] = b * c
		te[10] = a * c
	case YZXOrder:
		ac := a * c
		ad := a * d
		bc := b * c
		bd := b * d

		te[0] = c * e
		te[4] = bd - ac*f
		te[8] = bc*f + ad

		te[1] = f
		te[5] = a * e
		te[9] = -b * e

		te[2] = -d * e
		te[6] = ad*f + bc
		te[10] = ac - bd*f
	case XZYOrder:
		ac := a * c
		ad := a * d
		bc := b * c
		bd := b * d

		te[0] = c * e
		te[4] = -f
		te[8] = d * e

		te[1] = ac*f + bd
		te[5] = a * e
		te[9] = ad*f - bc

		te[2] = bc*f - ad
		te[6] = b * e
		te[10] = bd*f + ac
	}
	te[15] = 1
	return te
}

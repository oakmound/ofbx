package threefbx

import (
	"math"

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

func (e Euler) ToQuaternion() floatgeom.Point4 {
	cy := math.Cos(e.Z() * 0.5)
	sy := math.Sin(e.Z() * 0.5)
	cr := math.Cos(e.Y() * 0.5)
	sr := math.Sin(e.Y() * 0.5)
	cp := math.Cos(e.X() * 0.5)
	sp := math.Sin(e.X() * 0.5)

	return floatgeom.Point4{
		cy*sr*cp - sy*cr*sp,
		cy*cr*sp + sy*sr*cp,
		sy*cr*cp - cy*sr*sp,
		cy*cr*cp + sy*sr*sp,
	}
}

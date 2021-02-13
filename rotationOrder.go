package ofbx

import (
	"github.com/oakmound/oak/v2/alg"
	"github.com/oakmound/oak/v2/alg/floatgeom"
)

// RotationOrder determines the dimension order for rotation
type RotationOrder int

// A block of rotation order sets
const (
	EulerXYZ   RotationOrder = iota
	EulerXZY   RotationOrder = iota
	EulerYZX   RotationOrder = iota
	EulerYXZ   RotationOrder = iota
	EulerZXY   RotationOrder = iota
	EulerZYX   RotationOrder = iota
	SphericXYZ RotationOrder = iota // Currently unsupported. Treated as EulerXYZ.
)

func (o RotationOrder) rotationMatrix(euler floatgeom.Point3) Matrix {
	rx := rotationX(euler.X() * alg.DegToRad)
	ry := rotationY(euler.Y() * alg.DegToRad)
	rz := rotationZ(euler.Z() * alg.DegToRad)
	switch o {
	default:
	case SphericXYZ:
		panic("This should not happen")
	case EulerXYZ:
		return rz.Mul(ry).Mul(rx)
	case EulerXZY:
		return ry.Mul(rz).Mul(rx)
	case EulerYXZ:
		return rz.Mul(rx).Mul(ry)
	case EulerYZX:
		return rx.Mul(rz).Mul(ry)
	case EulerZXY:
		return ry.Mul(rx).Mul(rz)
	case EulerZYX:
		return rx.Mul(ry).Mul(rz)
	}
	panic("This shouldn't happen either")
}

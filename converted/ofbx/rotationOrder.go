package ofbx

type RotationOrder int

const (
	EULER_XYZ   RotationOrder = iota
	EULER_XZY   RotationOrder = iota
	EULER_YZX   RotationOrder = iota
	EULER_YXZ   RotationOrder = iota
	EULER_ZXY   RotationOrder = iota
	EULER_ZYX   RotationOrder = iota
	SPHERIC_XYZ RotationOrder = iota // Currently unsupported. Treated as EULER_XYZ.
)

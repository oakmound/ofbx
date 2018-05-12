package ofbx

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

type Vec2 struct {
	x, y float64
}

type Vec3 struct {
	x, y, z float64
}

type Vec4 struct {
	x, y, z, w float64
}

type Matrix struct {
	m [16]float64 // last 4 are translation
}

type Quat struct {
	x, y, z, w float64
}


func (v *Vec3) Minus() *Vec3{
	return {-v.x, -v.y, -v.z}
}

func (v *Vec3) Mul(f){
	return NewVec3(v.x * f, v.y * f, v.z * f)
}


func (v *Vec3) Add(v2 *Vec3){
	return NewVec3(v.x + v2.x, v.y + v2.y, v.z + v2.z)
}


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
	X, Y float64
}

type Vec3 struct {
	X, Y, Z float64
}

type Vec4 struct {
	X, Y, Z, w float64
}

type Matrix struct {
	m [16]float64 // last 4 are translation
}

type Quat struct {
	X, Y, Z, w float64
}


func (v *Vec3) Minus() *Vec3{
	return Vec3{-v.X, -v.Y, -v.Z}
}

func (v *Vec3) Mul(f){
	return NewVec3(v.X * f, v.Y * f, v.Z * f)
}


func (v *Vec3) Add(v2 *Vec3){
	return NewVec3(v.X + v2.X, v.Y + v2.Y, v.Z + v2.Z)
}


func (m1 *Matrix) Mul(m2){
	result := make{[]float64, 16}
	for j := 0; j < 4; j++	{
		for  i := 0; i < 4; i++{
			tmp := 0.0
			for k := 0; k < 4; k++{
				tmp += m1.m[i + k * 4] * m2.m[k + j * 4]
			}
			res.m[i + j * 4] = tmp
		}
	}
	return res
}

func makeIdentity() Matrix {  
	return Matrix{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}	
}

func (m *Matrix) rotationX(angle float64){
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

func (m *Matrix) rotationY(angle float64){
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

func (m *Matrix) rotationZ(angle float64){
	m2 := makeIdentity()
	//radian
	c := math.Cos(angle)
	s := math.Sin(angle)
	m2.m[0] =c
	m2.m[5] = c
	m2.m[4] = -s
	m2.m[1] = s
	return m2
}

func getTriCountFromPoly(indices []int, idx int) (int, int){
	count := 1
	for (indices[idx+1+count]>=0){
		count++
	}
	return count, idx
}


func getRotationMatrix(euler *Vec3, order RotationOrder) Matrix{
	TO_RAD :=  3.1415926535897932384626433832795028 / 180.0 //TODO: Update this
	rx = rotationX(euler.X * TO_RAD)
	ry = rotationY(euler.Y * TO_RAD)
	rz = rotationZ(euler.Z * TO_RAD)
	switch (order) {
	default:
	case SPHERIC_XYZ:
		errors.NewError("This should not happen")
	case EULER_XYZ:
		return rz.Mul(ry).Mul(rx)
	case EULER_XZY:
		return ry.Mul(rz).Mul(rx)
	case EULER_YXZ:
		return rz.Mul(rx).Mul( ry)
	case EULER_YZX:
		return rx.Mul(rz).Mul(ry)
	case EULER_ZXY:
		return ry.Mul(rx).Mul(rz)
	case EULER_ZYX:
		return rx.Mul(ry).Mul(rz)
}

}




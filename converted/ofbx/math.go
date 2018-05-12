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
	return {-v.X, -v.Y, -v.Z}
}

func (v *Vec3) Mul(f){
	return NewVec3(v.X * f, v.Y * f, v.Z * f)
}


func (v *Vec3) Add(v2 *Vec3){
	return NewVec3(v.X + v2.X, v.Y + v2.Y, v.Z + v2.Z)
}



//START MATRIX STUFF

static Matrix operator*(const Matrix& lhs, const Matrix& rhs)
{
	Matrix res;
	for (int j = 0; j < 4; ++j)
	{
		for (int i = 0; i < 4; ++i)
		{
			double tmp = 0;
			for (int k = 0; k < 4; ++k)
			{
				tmp += lhs.m[i + k * 4] * rhs.m[k + j * 4];
			}
			res.m[i + j * 4] = tmp;
		}
	}
	return res;
}


static Matrix makeIdentity()
{
	return {1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1};
}


static Matrix rotationX(double angle)
{
	Matrix m = makeIdentity();
	double c = cos(angle);
	double s = sin(angle);

	m.m[5] = m.m[10] = c;
	m.m[9] = -s;
	m.m[6] = s;

	return m;
}


static Matrix rotationY(double angle)
{
	Matrix m = makeIdentity();
	double c = cos(angle);
	double s = sin(angle);

	m.m[0] = m.m[10] = c;
	m.m[8] = s;
	m.m[2] = -s;

	return m;
}


static Matrix rotationZ(double angle)
{
	Matrix m = makeIdentity();
	double c = cos(angle);
	double s = sin(angle);

	m.m[0] = m.m[5] = c;
	m.m[4] = -s;
	m.m[1] = s;

	return m;
}


static Matrix getRotationMatrix(const Vec3& euler, RotationOrder order)
{
	const double TO_RAD = 3.1415926535897932384626433832795028 / 180.0;
	Matrix rx = rotationX(euler.x * TO_RAD);
	Matrix ry = rotationY(euler.y * TO_RAD);
	Matrix rz = rotationZ(euler.z * TO_RAD);
	switch (order) {
		default:
		case RotationOrder::SPHERIC_XYZ:
			assert(false);
		case RotationOrder::EULER_XYZ:
			return rz * ry * rx;
		case RotationOrder::EULER_XZY:
			return ry * rz * rx;
		case RotationOrder::EULER_YXZ:
			return rz * rx * ry;
		case RotationOrder::EULER_YZX:
			return rx * rz * ry;
		case RotationOrder::EULER_ZXY:
			return ry * rx * rz;
		case RotationOrder::EULER_ZYX:
			return rx * ry * rz;
	}
}


//END MATRIX

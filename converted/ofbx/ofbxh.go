package ofbx

import "bytes"

type Color struct {
	r, g, b float32
}

// http://docs.autodesk.com/FBX/2014/ENU/FBX-SDK-Documentation/index.html?url=cpp_ref/class_fbx_time.html,topicNumber=cpp_ref_class_fbx_time_html29087af6-8c2c-4e9d-aede-7dc5a1c2436c,hash=a837590fd5310ff5df56ffcf7c394787e
type FrameRate int

const (
	FrameRate_DEFAULT         FrameRate = iota
	FrameRate_120             FrameRate = iota
	FrameRate_100             FrameRate = iota
	FrameRate_60              FrameRate = iota
	FrameRate_50              FrameRate = iota
	FrameRate_48              FrameRate = iota
	FrameRate_30              FrameRate = iota
	FrameRate_30_DROP         FrameRate = iota
	FrameRate_NTSC_DROP_FRAME FrameRate = iota
	FrameRate_NTSC_FULL_FRAME FrameRate = iota
	FrameRate_PAL             FrameRate = iota
	FrameRate_CINEMA          FrameRate = iota
	FrameRate_1000            FrameRate = iota
	FrameRate_CINEMA_ND       FrameRate = iota
	FrameRate_CUSTOM          FrameRate = iota
)

type Type int

const (
	ROOT                 Type = iota
	GEOMETRY             Type = iota
	MATERIAL             Type = iota
	MESH                 Type = iota
	TEXTURE              Type = iota
	LIMB_NODE            Type = iota
	NULL_NODE            Type = iota
	NODE_ATTRIBUTE       Type = iota
	CLUSTER              Type = iota
	SKIN                 Type = iota
	ANIMATION_STACK      Type = iota
	ANIMATION_LAYER      Type = iota
	ANIMATION_CURVE      Type = iota
	ANIMATION_CURVE_NODE Type = iota
)

type DataView bytes.Buffer

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

// template <typename T> T* resolveObjectLink(int idx) const
// {
// 	return static_cast<T*>(resolveObjectLink(T::s_type, nullptr, idx));
// }

type TakeInfo struct {
	name                DataView
	filename            DataView
	local_time_from     float64
	local_time_to       float64
	reference_time_from float64
	reference_time_to   float64
}

func getError() string {
	return ""
}

var (
	UpAxis                  UpVector    = UpVector_AxisX
	UpAxisSign              int         = 1
	FrontAxis               FrontVector = FrontVector_ParityOdd
	FrontAxisSign           int         = 1
	CoordAxis               CoordSystem = CoordSystem_RightHanded
	CoordAxisSign           int         = 1
	OriginalUpAxis          int         = 0
	OriginalUpAxisSign      int         = 1
	UnitScaleFactor         float32     = 1
	OriginalUnitScaleFactor float32     = 1
	TimeSpanStart           uint64      = 0
	TimeSpanStop            uint64      = 0
	TimeMode                FrameRate   = FrameRate_DEFAULT
	CustomFrameRate         float32     = -1.0
)

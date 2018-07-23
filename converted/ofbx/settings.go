package ofbx

// Settings is the overall scene fbx settings
type Settings struct {
	UpAxis                  UpVector
	UpAxisSign              int
	FrontAxis               FrontVector
	FrontAxisSign           int
	CoordAxis               CoordSystem
	CoordAxisSign           int
	OriginalUpAxis          int
	OriginalUpAxisSign      int
	UnitScaleFactor         float32
	OriginalUnitScaleFactor float32
	TimeSpanStart           uint64
	TimeSpanStop            uint64
	TimeMode                FrameRate
	CustomFrameRate         float32
}

// Default Settings
var (
	UpAxis                  = UpVector_AxisX
	UpAxisSign              = 1
	FrontAxis               = FrontVector_ParityOdd
	FrontAxisSign           = 1
	CoordAxis               = CoordSystem_RightHanded
	CoordAxisSign           = 1
	OriginalUpAxis          int
	OriginalUpAxisSign              = 1
	UnitScaleFactor         float32 = 1
	OriginalUnitScaleFactor float32 = 1
	TimeSpanStart           uint64
	TimeSpanStop            uint64
	TimeMode                        = FrameRate_DEFAULT
	CustomFrameRate         float32 = -1.0
)

package ofbx

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

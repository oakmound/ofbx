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
	UpAxis                  = UpVectorX
	UpAxisSign              = 1
	FrontAxis               = FrontVectorParityOdd
	FrontAxisSign           = 1
	CoordAxis               = CoordSystemRight
	CoordAxisSign           = 1
	OriginalUpAxis          int
	OriginalUpAxisSign              = 1
	UnitScaleFactor         float32 = 1
	OriginalUnitScaleFactor float32 = 1
	TimeSpanStart           uint64
	TimeSpanStop            uint64
	TimeMode                        = FrameRateDefault
	CustomFrameRate         float32 = -1.0
)

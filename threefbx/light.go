package threefbx

import (
	"math"

	"github.com/oakmound/oak/alg/floatgeom"
)

var (
	DefaultUp = floatgeom.Point3{1, 1, 1}
)

type baseLight struct {
	*baseModel
	castShadow bool
	color      Color
	intensity  float64
	target     floatgeom.Point3
}

type Light interface {
	Model

	SetCastShadow(bool)
	SetTarget(floatgeom.Point3)
}

func (bl *baseLight) IsGroup() bool {
	return false
}

//
func (bl *baseLight) SetCastShadow(b bool) {
	bl.castShadow = b
}

func (bl *baseLight) SetTarget(target floatgeom.Point3) {
	bl.target = target
}

// PointLight is a light  emenating from a single point
type PointLight struct {
	*baseLight
	distance float64
	decay    float64
}

// GetPower returns the luminostiy of the PointLight
func (pl *PointLight) GetPower() float64 {
	return pl.intensity * 4 * math.Pi
}

// SetPower sets the luminostiy of the PointLight
func (pl *PointLight) SetPower(power float64) {
	pl.intensity = power / (4 * math.Pi)
}

// DirectionalLight is a light
type DirectionalLight struct {
	*baseLight
}

// SpotLight is a light
type SpotLight struct {
	*baseLight
	distance float64
	decay    float64
	angle    float64
	penumbra float64
}

// GetPower returns the luminostiy of the SpotLight
func (pl *SpotLight) GetPower() float64 {
	return pl.intensity * math.Pi
}

// SetPower sets the luminostiy of the SpotLight
func (pl *SpotLight) SetPower(power float64) {
	pl.intensity = power / math.Pi
}

// NewDirectionalLight creates a new directonal light with the Default upwards position
func NewDirectionalLight(color Color, intensity float64) *DirectionalLight {
	dl := DirectionalLight{
		baseLight: &baseLight{
			color:     color,
			intensity: intensity,

			baseModel: &baseModel{
				position:   DefaultUp,
				scale:      floatgeom.Point3{1, 1, 1},
				quaternion: floatgeom.Point4{0, 0, 0, 1},
			},
		},
	}
	dl.matrix = composeMat(dl.position, dl.quaternion, dl.scale)

	// this.shadow = new DirectionalLightShadow(); //TODO: Shadows are reworked later in engine
	return &dl
}

// NewPointLight creates a Point that emits a light
func NewPointLight(color Color, intensity, distance, decay float64) *PointLight {

	npl := PointLight{
		baseLight: &baseLight{
			baseModel: &baseModel{
				position:   DefaultUp,
				scale:      floatgeom.Point3{1, 1, 1},
				quaternion: floatgeom.Point4{0, 0, 0, 1},
			},
			color:     color,
			intensity: intensity,
		},
		distance: distance,
		decay:    decay,
	}
	// this.shadow = new LightShadow( new PerspectiveCamera( 90, 1, 0.5, 500 ) );
	return &npl
}

// NewSpotLight creates a spotlight
func NewSpotLight(color Color, intensity, distance, angle, penumbra, decay float64) *SpotLight {
	nsl := SpotLight{
		baseLight: &baseLight{
			baseModel: &baseModel{
				position:   DefaultUp,
				scale:      floatgeom.Point3{1, 1, 1},
				quaternion: floatgeom.Point4{0, 0, 0, 1},
			},
			color:     color,
			intensity: intensity,
		},
		distance: distance,
		decay:    decay,
		angle:    angle,
		penumbra: penumbra,
	}

	return &nsl
}

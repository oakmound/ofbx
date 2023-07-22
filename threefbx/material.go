package threefbx

import (
	"strings"

	"github.com/oakmound/oak/v4/alg/floatgeom"
)

type TextureMapping int

const (
	NoToneMapping                    TextureMapping = 0
	LinearToneMapping                TextureMapping = 1
	ReinhardToneMapping              TextureMapping = 2
	Uncharted2ToneMapping            TextureMapping = 3
	CineonToneMapping                TextureMapping = 4
	UVMapping                        TextureMapping = 300
	CubeReflectionMapping            TextureMapping = 301
	CubeRefractionMapping            TextureMapping = 302
	EquirectangularReflectionMapping TextureMapping = 303
	EquirectangularRefractionMapping TextureMapping = 304
	SphericalReflectionMapping       TextureMapping = 305
	CubeUVReflectionMapping          TextureMapping = 306
	CubeUVRefractionMapping          TextureMapping = 307
)

type NormalMap int

const (
	TangentSpaceNormalMap NormalMap = iota
	ObjectSpaceNormalMap  NormalMap = iota
)

type Operation int

const (
	MultiplyOperation Operation = iota
	MixOperation      Operation = iota
	AddOperation      Operation = iota
)

type Blending int

const (
	NoBlending          Blending = iota
	NormalBlending      Blending = iota
	AdditiveBlending    Blending = iota
	SubtractiveBlending Blending = iota
	MultiplyBlending    Blending = iota
	CustomBlending      Blending = iota
)

type Side int

const (
	FrontSide  Side = iota
	BackSide   Side = iota
	DoubleSide Side = iota
)

type DepthFunc int

const (
	NeverDepth        DepthFunc = iota
	AlwaysDepth       DepthFunc = iota
	LessDepth         DepthFunc = iota
	LessEqualDepth    DepthFunc = iota
	EqualDepth        DepthFunc = iota
	GreaterEqualDepth DepthFunc = iota
	GreaterDepth      DepthFunc = iota
	NotEqualDepth     DepthFunc = iota
)

type Equation int

const (
	AddEquation             Equation = iota
	SubtractEquation        Equation = iota
	ReverseSubtractEquation Equation = iota
	MinEquation             Equation = iota
	MaxEquation             Equation = iota
)

type Factor int

const (
	ZeroFactor             Factor = iota
	OneFactor              Factor = iota
	SrcColorFactor         Factor = iota
	OneMinusSrcColorFactor Factor = iota
	SrcAlphaFactor         Factor = iota
	OneMinusSrcAlphaFactor Factor = iota
	DstAlphaFactor         Factor = iota
	OneMinusDstAlphaFactor Factor = iota
	DstColorFactor         Factor = iota
	OneMinusDstColorFactor Factor = iota
	SrcAlphaSaturateFactor Factor = iota
)

type VColoring int

const (
	NoColors     VColoring = iota
	FaceColors   VColoring = iota
	VertexColors VColoring = iota
)

type Material struct {
	MaterialParameters
	Name string

	fog         bool
	lights      bool
	flatShading bool

	//These are probably ints..
	blending     Blending
	side         Side
	vertexColors VColoring

	blendSrc      Factor
	blendDst      Factor
	blendEquation Equation

	blendSrcAlpha      *int
	blendDstAlpha      *int
	blendEquationAlpha *int

	depthFunc  DepthFunc
	depthTest  bool
	depthWrite bool

	clippingPlanes   []floatgeom.Point4
	clipIntersection bool
	clipShadows      bool

	shadowSide Side

	colorWrite bool

	precision float64 // override the renderer's default precision for material

	polygonOffset       bool
	polygonOffsetFactor int
	polygonOffsetUnits  int

	dithering bool

	alphaTest          int
	premultipliedAlpha bool

	overdraw float64 // Overdrawn pixels (typically between 0 and 1) for fixing antialiasing gaps in CanvasRenderer

	visible bool

	color             Color
	lightMapIntensity float64
	aoMapIntensity    float64

	emissiveIntensity float64
	combine           Operation
	reflectivity      int

	refractionRatio float64

	wireframeLinewidth int
	wireframeLinecap   string
	wireframeLinejoin  string
	skinning           bool
	morphTargets       bool
	morphNormals       bool

	bumpScale     int
	normalMapType NormalMap
	normalScale   floatgeom.Point2

	displacementScale int

	lineWidth float64

	needsUpdate bool
}

// MaterialParameteres are extracted from a material node and applied to Material
type MaterialParameters struct {
	BumpFactor         float64
	Diffuse            Color
	DisplacementFactor float64
	Emissive           Color
	EmissiveFactor     float64
	Opacity            float64
	ReflectionFactor   float64
	Shininess          float64
	Specular           Color
	Transparent        bool

	bumpMap         Texture
	normalMap       Texture
	specularMap     Texture
	emissiveMap     Texture
	diffuseMap      Texture
	alphaMap        Texture
	displacementMap Texture
	envMap          Texture
}

// ShadingModel is a parameter that determines the type of underlying material to be used.
type ShadingModel string

// Material creates a basic version of a material based on the type of shading model
func (sm *ShadingModel) Material() Material {

	if sm == nil {
		return NewMeshPhong()
	}
	if strings.ToLower(string(*sm)) == "lambert" {
		return NewMeshLambert()
	}
	return NewMeshPhong()
}

//TODO: Get Material creation working

func NewMaterial() Material {
	return Material{
		MaterialParameters: MaterialParameters{Opacity: 1},
		fog:                true,
		lights:             true,
		blending:           NormalBlending,

		blendSrc:      SrcAlphaFactor,
		blendDst:      OneMinusSrcAlphaFactor,
		blendEquation: AddEquation,
		depthFunc:     LessEqualDepth,
		depthTest:     true,
		depthWrite:    true,
		colorWrite:    true,
		visible:       true,
		needsUpdate:   true,
	}
}

// NewMeshLambert creates a new basic Lambert material
func NewMeshLambert() Material {

	mp := Material{

		MaterialParameters: MaterialParameters{Emissive: Color{0, 0, 0}},

		color:             Color{255, 255, 255},
		lightMapIntensity: 1.0,
		aoMapIntensity:    1.0,

		emissiveIntensity: 1.0,
		combine:           MultiplyOperation,
		reflectivity:      1,
		refractionRatio:   0.98,

		wireframeLinewidth: 1,
		wireframeLinecap:   "round",
		wireframeLinejoin:  "round",
		skinning:           false,
		morphTargets:       false,
		morphNormals:       false,
	}
	// setValues(parameters),
	return mp
}

// NewMeshPhong creates a basic Phong Material
func NewMeshPhong() Material { //https://github.com/mrdoob/three.js/blob/34dc2478c684066257e4e39351731a93c6107ef5/src/materials/MeshPhongMaterial.js#L57
	mp := Material{
		MaterialParameters: MaterialParameters{
			Specular:  Color{17, 17, 17},
			Shininess: 30,
		},

		color: Color{255, 255, 255},

		lightMapIntensity: 1.0,
		aoMapIntensity:    1.0,
		emissiveIntensity: 1.0,
		bumpScale:         1,

		normalMapType: TangentSpaceNormalMap, //ToDO: Create constant
		normalScale:   floatgeom.Point2{1, 1},

		displacementScale: 1,
		combine:           MultiplyOperation,

		reflectivity:    1,
		refractionRatio: 0.98,

		wireframeLinewidth: 1,
		wireframeLinecap:   "round",
		wireframeLinejoin:  "round",
	}
	// setValues(parameters),

	return mp
}

func NewLineBasicMaterial(c Color, opacity, lineWidth float64) Material {
	mat := NewMaterial()
	mat.lights = false
	mat.lineWidth = lineWidth
	mat.color = c
	mat.Opacity = opacity
	return mat
}

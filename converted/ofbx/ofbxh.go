package ofbx

import "bytes"

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

type Color struct {
	r, g, b float32
}

// Specifies which canonical axis represents up in the system (typically Y or Z).

type UpVector int

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

type PropertyType rune

const (
	LONG         PropertyType = 'L'
	INTEGER      PropertyType = 'I'
	STRING       PropertyType = 'S'
	FLOAT        PropertyType = 'F'
	DOUBLE       PropertyType = 'D'
	ARRAY_DOUBLE PropertyType = 'd'
	ARRAY_INT    PropertyType = 'i'
	ARRAY_LONG   PropertyType = 'l'
	ARRAY_FLOAT  PropertyType = 'f'
)

type IElementProperty struct {
}

func (iep *IElementProperty) getType() Type {
	return 0
}

func (iep *IElementProperty) getNext() *IElementProperty {
	return nil
}

func (iep *IElementProperty) getValue() DataView {
	return DataView{}
}

func (iep *IElementProperty) getCount() int {
	return 0
}

func (iep *IElementProperty) getValuesF64(values []float64, max_size int) bool {
	return false
}

func (iep *IElementProperty) getValuesInt(values []int, max_size int) bool {
	return false
}

func (iep *IElementProperty) getValuesF32(values []float32, max_size int) bool {
	return false
}

func (iep *IElementProperty) getValuesUInt64(values []uint64, max_size int) bool {
	return false
}

func (iep *IElementProperty) getValuesInt64(values []int64, max_size int) bool {
	return false
}

type IElement struct{}

func (ie *IElement) getFirstChild() *IElement {
	return nil
}
func (ie *IElement) getSibling() *IElement {
	return nil
}
func (ie *IElement) getID() DataView {
	return DataView{}
}
func (ie *IElement) getFirstProperty() *IElementProperty {
	return nil
}

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

type Object struct {
	ID             uint64
	Name           string
	Element        *IElement
	Node_attribute *Object

	is_node bool
	scene   *Scene
}

func NewObject(scene *Scene, element *IElement) *Object {
	return nil
}

func (o *Object) getType() Type {
	return 0
}

func (o *Object) getScene() *IScene {
	return nil
}

func (o *Object) resolveObjectLinkIndex(idx int) *Object {
	return nil
}
func (o *Object) resolveObjectLink(typ Type, property string, idx int) *Object {
	return nil
}
func (o *Object) resolveObjectLinkReverse(typ Type) *Object {
	return nil
}
func (o *Object) getParent() *Object {
	return nil
}
func (o *Object) getRotationOrder() RotationOrder {
	return RotationOrder{}
}
func (o *Object) getRotationOffset() Vec3 {
	return Vec3{}
}

func (o *Object) getRotationPivot() Vec3 {
	return Vec3{}
}

func (o *Object) getPostRotation() Vec3 {
	return Vec3{}
}

func (o *Object) getScalingOffset() Vec3 {
	return Vec3{}
}

func (o *Object) getScalingPivot() Vec3 {
	return Vec3{}
}

func (o *Object) getPreRotation() Vec3 {
	return Vec3{}
}

func (o *Object) getLocalTranslation() Vec3 {
	return Vec3{}
}

func (o *Object) getLocalRotation() Vec3 {
	return Vec3{}
}

func (o *Object) getLocalScaling() Vec3 {
	return Vec3{}
}

func (o *Object) getGlobalTransform() Matrix {
	return Matrix{}
}

func (o *Object) getLocalTransform() Matrix {
	return Matrix{}
}

func (o *Object) evalLocal(translation, rotation *Vec3) Matrix {
	return Matrix{}
}

func (o *Object) evalLocalScaling(translation, rotation, scaling *Vec3) Matrix {
	return Matrix{}
}

func (o *Object) isNode() bool {
	return o.is_node
}

// template <typename T> T* resolveObjectLink(int idx) const
// {
// 	return static_cast<T*>(resolveObjectLink(T::s_type, nullptr, idx));
// }

type TextureType int

const (
	DIFFUSE TextureType = iota
	NORMAL  TextureType = iota
	COUNT   TextureType = iota
)

type Texture struct {
	Object
}

func NewTexture(scene *Scene, element *IElement) *Texture {
	return nil
}

func (t *Texture) Type() Type {
	return TEXTURE
}

func (t *Texture) getFileName() DataView {
	return DataView{}
}

func (t *Texture) getRelativeFileName() DataView {
	return DataView{}
}

type Material struct {
	Object
}

func NewMaterial(scene *Scene, element *IElement) *Material {
	return nil
}

func (m *Material) Type() Type {
	return MATERIAL
}

func (m *Material) getDiffuseColor() Color {
	return Color{}
}

func (m *Material) getTexture(typ TextureType) *Texture {
	return nil
}

type Cluster struct {
	Object
}

func NewCluster(scene *Scene, element *IElement) *Cluster {
	return nil
}

func (c *Cluster) Type() Type {
	return CLUSTER
}

func (c *Cluster) getIndices() []int {
	return nil
}

func (c *Cluster) getWeights() *float64 {
	return nil
}

func (c *Cluster) getTransformMatrix() Matrix {
	return Matrix{}
}

func (c *Cluster) getTransformLinkMatrix() Matrix {
	return Matrix{}
}

func (c *Cluster) getLink() *Object {
	return nil
}

type Skin struct {
	Object
}

func NewSkin(scene *Scene, element *IElement) *Skin {
	return nil
}

func (s *Skin) Type() Type {
	return SKIN
}

func (s *Skin) getCluster() []Cluster {
	return nil
}

type NodeAttribute struct {
	Object
}

func NewNodeAttribute(scene *Scene, element *IElement) *NodeAttribute {
	return nil
}

func (na *NodeAttribute) Type() Type {
	return NODE_ATTRIBUTE
}

func (na NodeAttribute) getAttributeType() DataView {
	return DataView{}
}

type Geometry struct {
	Object
}

func NewGeometry(scene *Scene, element *IElement) *Geometry {
	return nil
}

func (g *Geometry) Type() Type {
	return GEOMETRY
}

func (g *Geometry) UVSMax() int {
	return 4
}

func (g *Geometry) getVertices() []Vec3 {
	return nil
}

func (g *Geometry) getNormals() *Vec3 {
	return nil
}

func (g *Geometry) getUVs(index int) *Vec2 {
	return nil
}

func (g *Geometry) getColors() *Vec4 {
	return nil
}

func (g *Geometry) getTangents() *Vec3 {
	return nil
}

func (g *Geometry) getSkin() *Skin {
	return nil
}

func (g *Geometry) getMaterials() *int {
	return nil
}

type Mesh struct {
	Object
}

func NewMesh(scene *Scene, element *IElement) *Mesh {
	return nil
}

func (m *Mesh) Type() Type {
	return MESH
}

func (m *Mesh) getGeometry() *Geometry {
	return nil
}

func (m *Mesh) getGeometricMatrix() Matrix {
	return Matrix{}
}

func (m *Mesh) getMaterial(idx int) []Material {
	return nil
}

type AnimationStack struct {
	Object
}

func NewAnimationStack(scene *Scene, element *IElement) *AnimationStack {
	return nil
}

func (as *AnimationStack) Type() Type {
	return ANIMATION_STACK
}

func (as *AnimationStack) getLayer() int {
	return 0
}

type AnimationLayer struct {
	Object
}

func NewAnimationLayer(scene *Scene, element *IElement) *AnimationLayer {
	return nil
}

func (as *AnimationLayer) Type() Type {
	return ANIMATION_LAYER
}

func (as *AnimationLayer) getCurveNodeIndex(index int) *AnimationCurveNode {
	return 0
}

func (as *AnimationLayer) getCurveNodeIndex(bone *Object, property string) *AnimationCurveNode {
	return 0
}

type AnimationCurve struct {
	Object
}

func NewAnimationCurve(scene *Scene, element *IElement) *AnimationCurve {
	return nil
}

func (ac *AnimationCurve) Type() Type {
	return ANIMATION_CURVE
}

// 200sc note: this may be the length of the next two functions, i.e. unneeded
func (ac *AnimationCurve) getKeyCount() int {
	return 0
}

func (ac *AnimationCurve) getKeyTime() *int64 {
	return nil
}

func (ac *AnimationCurve) getKeyValue() *float32 {
	return nil
}

type AnimationCurveNode struct {
	Object
}

func (acn *AnimationCurveNode) Type() Type {
	return ANIMATION_CURVE_NODE
}

func (acn *AnimationCurveNode) getNodeLocalTransform(time float64) Vec3 {
	return Vec3{}
}

func (acn *AnimationCurveNode) getBone() *Object {
	return nil
}

type TakeInfo struct {
	name                DataView
	filename            DataView
	local_time_from     float64
	local_time_to       float64
	reference_time_from float64
	reference_time_to   float64
}

type Scene struct{}

type IScene struct{}

func (is *IScene) getRootElement() *IElement {
	return nil
}
func (is *IScene) getRoot() *Object {
	return nil
}
func (is *IScene) getTakeInfo(name string) *TakeInfo {
	return nil
}
func (is *IScene) getSceneFrameRate() float32 {
	return 0
}
func (is *IScene) getMesh(int index) []Mesh {
	return nil
}
func (is *IScene) getAnimationStack(index int) []AnimationStack {
	return nil
}
func (is *IScene) getAllObjects() []Object {
	return nil
}

func load(data []byte) *Iscene {
	return nil
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

package ofbx

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
	ROOT Type = iota
	GEOMETRY Type = iota
	MATERIAL Type = iota
	MESH Type = iota
	TEXTURE Type = iota
	LIMB_NODE Type = iota
	NULL_NODE Type = iota
	NODE_ATTRIBUTE Type = iota
	CLUSTER Type = iota
	SKIN Type = iota
	ANIMATION_STACK Type = iota
	ANIMATION_LAYER Type = iota
	ANIMATION_CURVE Type = iota
	ANIMATION_CURVE_NODE Type = iota
) 

type Object {

}
	Object(const Scene& _scene, const IElement& _element);

	virtual ~Object() {}
	virtual Type getType() const = 0;
	
	const IScene& getScene() const;
	Object* resolveObjectLink(int idx) const;
	Object* resolveObjectLink(Type type, const char* property, int idx) const;
	Object* resolveObjectLinkReverse(Type type) const;
	Object* getParent() const;

    RotationOrder getRotationOrder() const;
	Vec3 getRotationOffset() const;
	Vec3 getRotationPivot() const;
	Vec3 getPostRotation() const;
	Vec3 getScalingOffset() const;
	Vec3 getScalingPivot() const;
	Vec3 getPreRotation() const;
	Vec3 getLocalTranslation() const;
	Vec3 getLocalRotation() const;
	Vec3 getLocalScaling() const;
	Matrix getGlobalTransform() const;
	Matrix getLocalTransform() const;
	Matrix evalLocal(const Vec3& translation, const Vec3& rotation) const;
	Matrix evalLocal(const Vec3& translation, const Vec3& rotation, const Vec3& scaling) const;
	bool isNode() const { return is_node; }


	template <typename T> T* resolveObjectLink(int idx) const
	{
		return static_cast<T*>(resolveObjectLink(T::s_type, nullptr, idx));
	}

	uint64 id;
	char name[128];
	const IElement& element;
	const Object* node_attribute;

protected:
	bool is_node;
	const Scene& scene;
};


type Texture struct { Object }
{
	enum TextureType
	{
		DIFFUSE,
		NORMAL,

		COUNT
	};

	static const Type s_type = Type::TEXTURE;

	Texture(const Scene& _scene, const IElement& _element);
	virtual DataView getFileName() const = 0;
	virtual DataView getRelativeFileName() const = 0;
};


type Material struct { Object }
{
	static const Type s_type = Type::MATERIAL;

	Material(const Scene& _scene, const IElement& _element);

	virtual Color getDiffuseColor() const = 0;
	virtual const Texture* getTexture(Texture::TextureType type) const = 0;
};


type Cluster struct { Object }
{
	static const Type s_type = Type::CLUSTER;

	Cluster(const Scene& _scene, const IElement& _element);

	virtual const int* getIndices() const = 0;
	virtual int getIndicesCount() const = 0;
	virtual const double* getWeights() const = 0;
	virtual int getWeightsCount() const = 0;
	virtual Matrix getTransformMatrix() const = 0;
	virtual Matrix getTransformLinkMatrix() const = 0;
	virtual const Object* getLink() const = 0;
};


type Skin struct { Object }
{
	static const Type s_type = Type::SKIN;

	Skin(const Scene& _scene, const IElement& _element);

	virtual int getClusterCount() const = 0;
	virtual const Cluster* getCluster(int idx) const = 0;
};



type NodeAttribute struct { 
	Object 
}

func NewNodeAttribute(scene *Scene, element *IElement) *NodeAttribute {
	return nil
}

func (na *NodeAttribute) Type() Type {
	return NODE_ATTRIBUTE
}

{
	static const Type s_type = Type::NODE_ATTRIBUTE;

	NodeAttribute(const Scene& _scene, const IElement& _element);

	virtual DataView getAttributeType() const = 0;
};


type Geometry struct { Object }
{
	static const Type s_type = Type::GEOMETRY;
    static const int s_uvs_max = 4;

	Geometry(const Scene& _scene, const IElement& _element);

	virtual const Vec3* getVertices() const = 0;
	virtual int getVertexCount() const = 0;

	virtual const Vec3* getNormals() const = 0;
	virtual const Vec2* getUVs(int index = 0) const = 0;
	virtual const Vec4* getColors() const = 0;
	virtual const Vec3* getTangents() const = 0;
	virtual const Skin* getSkin() const = 0;
	virtual const int* getMaterials() const = 0;
};

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

func (m *Mesh) getMaterial(idx int) *Material {
	return nil
}

func (m *Mesh) getMaterialCount() int {
	return 0
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

func (ac *AnimationCurve) int getKeyCount() int {
	return 0
}

func (ac *AnimationCurve) const int64* getKeyTime() *int64 {
	return nil
}

func (ac *AnimationCurve) const float* getKeyValue() *float32 {
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
	name DataView
	filename DataView
	local_time_from float64
	local_time_to float64
	reference_time_from float64
	reference_time_to float64
}

var (
	UpAxis UpVector = UpVector_AxisX
	UpAxisSign int = 1
	FrontAxis FrontVector = FrontVector_ParityOdd
	FrontAxisSign int = 1
	CoordAxis CoordSystem = CoordSystem_RightHanded
	CoordAxisSign int = 1
	OriginalUpAxis int = 0
	OriginalUpAxisSign int = 1
	UnitScaleFactor float32 = 1
	OriginalUnitScaleFactor float32 = 1
	TimeSpanStart uint64 = 0
	TimeSpanStop uint64 = 0
	TimeMode FrameRate = FrameRate_DEFAULT
	CustomFrameRate float32 = -1.0
)
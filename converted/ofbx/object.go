package ofbx

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

RotationOrder Object::getRotationOrder() const
{
	// This assumes that the default rotation order is EULER_XYZ.
	return (RotationOrder) resolveEnumProperty(*this, "RotationOrder", (int) RotationOrder::EULER_XYZ);
}


Vec3 Object::getRotationOffset() const
{
	return resolveVec3Property(*this, "RotationOffset", {0, 0, 0});
}


Vec3 Object::getRotationPivot() const
{
	return resolveVec3Property(*this, "RotationPivot", {0, 0, 0});
}


Vec3 Object::getPostRotation() const
{
	return resolveVec3Property(*this, "PostRotation", {0, 0, 0});
}


Vec3 Object::getScalingOffset() const
{
	return resolveVec3Property(*this, "ScalingOffset", {0, 0, 0});
}


Vec3 Object::getScalingPivot() const
{
	return resolveVec3Property(*this, "ScalingPivot", {0, 0, 0});
}


Matrix Object::evalLocal(const Vec3& translation, const Vec3& rotation) const
{
	return evalLocal(translation, rotation, getLocalScaling());
}


Matrix Object::evalLocal(const Vec3& translation, const Vec3& rotation, const Vec3& scaling) const
{
	Vec3 rotation_pivot = getRotationPivot();
	Vec3 scaling_pivot = getScalingPivot();
	RotationOrder rotation_order = getRotationOrder();

	Matrix s = makeIdentity();
	s.m[0] = scaling.x;
	s.m[5] = scaling.y;
	s.m[10] = scaling.z;

	Matrix t = makeIdentity();
	setTranslation(translation, &t);

	Matrix r = getRotationMatrix(rotation, rotation_order);
	Matrix r_pre = getRotationMatrix(getPreRotation(), RotationOrder::EULER_XYZ);
	Matrix r_post_inv = getRotationMatrix(-getPostRotation(), RotationOrder::EULER_ZYX);

	Matrix r_off = makeIdentity();
	setTranslation(getRotationOffset(), &r_off);

	Matrix r_p = makeIdentity();
	setTranslation(rotation_pivot, &r_p);

	Matrix r_p_inv = makeIdentity();
	setTranslation(-rotation_pivot, &r_p_inv);

	Matrix s_off = makeIdentity();
	setTranslation(getScalingOffset(), &s_off);

	Matrix s_p = makeIdentity();
	setTranslation(scaling_pivot, &s_p);

	Matrix s_p_inv = makeIdentity();
	setTranslation(-scaling_pivot, &s_p_inv);

	// http://help.autodesk.com/view/FBX/2017/ENU/?guid=__files_GUID_10CDD63C_79C1_4F2D_BB28_AD2BE65A02ED_htm
	return t * r_off * r_p * r_pre * r * r_post_inv * r_p_inv * s_off * s_p * s * s_p_inv;
}


Vec3 Object::getLocalTranslation() const
{
	return resolveVec3Property(*this, "Lcl Translation", {0, 0, 0});
}


Vec3 Object::getPreRotation() const
{
	return resolveVec3Property(*this, "PreRotation", {0, 0, 0});
}


Vec3 Object::getLocalRotation() const
{
	return resolveVec3Property(*this, "Lcl Rotation", {0, 0, 0});
}


Vec3 Object::getLocalScaling() const
{
	return resolveVec3Property(*this, "Lcl Scaling", {1, 1, 1});
}


Matrix Object::getGlobalTransform() const
{
	const Object* parent = getParent();
	if (!parent) return evalLocal(getLocalTranslation(), getLocalRotation());

	return parent.getGlobalTransform() * evalLocal(getLocalTranslation(), getLocalRotation());
}


Matrix Object::getLocalTransform() const
{
    return evalLocal(getLocalTranslation(), getLocalRotation(), getLocalScaling());
}


Object* Object::resolveObjectLinkReverse(Object::Type type) const
{
	uint64 id = element.getFirstProperty() ? element.getFirstProperty().getValue().touint64() : 0;
	for (auto& connection : scene.m_connections)
	{
		if (connection.from == id && connection.to != 0)
		{
			Object* obj = scene.m_object_map.find(connection.to).second.object;
			if (obj && obj.getType() == type) return obj;
		}
	}
	return nullptr;
}

func (o *Object) getScene() IScene{
	return o.scene
}

Object* Object::resolveObjectLink(int idx) const
{
	uint64 id = element.getFirstProperty() ? element.getFirstProperty().getValue().touint64() : 0;
	for (auto& connection : scene.m_connections)
	{
		if (connection.to == id && connection.from != 0)
		{
			Object* obj = scene.m_object_map.find(connection.from).second.object;
			if (obj)
			{
				if (idx == 0) return obj;
				--idx;
			}
		}
	}
	return nullptr;
}


Object* Object::resolveObjectLink(Object::Type type, const char* property, int idx) const
{
	uint64 id = element.getFirstProperty() ? element.getFirstProperty().getValue().touint64() : 0;
	for (auto& connection : scene.m_connections)
	{
		if (connection.to == id && connection.from != 0)
		{
			Object* obj = scene.m_object_map.find(connection.from).second.object;
			if (obj && obj.getType() == type)
			{
				if (property == nullptr || connection.property == property)
				{
					if (idx == 0) return obj;
					--idx;
				}
			}
		}
	}
	return nullptr;
}

func (o *Object) getParent() *Object{
	scene := o.getScene()
	for con := range scene.m_connections{
		if con.from == o.id{
			obj := scene.m_object_map[con.to]
			
			if (obj && obj.is_node) 
			{
				return obj
			}
		}
	}
}

Object* Object::getParent() const
{


	Object* parent = nullptr;
	for (auto& connection : scene.m_connections)
	{
		if (connection.from == id)
		{
			Object* obj = scene.m_object_map.find(connection.to).second.object;
			if (obj && obj.is_node)
			{
				assert(parent == nullptr);
				parent = obj;
			}
		}
	}
	return parent;
}
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
	return Object{}
}

// We'll need to worry about how this is used:
// it's used right now to be able to iterate over objects
// and call into types that have nested objects to check their
// type. 
// func (o *Object) getType() Type {
// 	return 0
// }

func (o *Object) getScene() *IScene {
	return o.scene
}

func (o *Object) resolveObjectLinkIndex(idx int) *Object {
	var id uint64
	if element.getFirstProperty() != nil {
		id = element.getFirstProperty().getValue().touint64()
	}
	for _, conn := range o.scene.m_connections {
		if connection.to == id && connection.from != 0 {
			obj := o.scene.m_object_map.find(connection.from).second.object
			if obj != nil {
				if idx == 0 {
					return obj
				}
				idx--
			}
		}
	}
	return nil
}

func (o *Object) resolveObjectLink(typ Type, property string, idx int) *Object {
	var id uint64
	if element.getFirstProperty() != nil {
		id = element.getFirstProperty().getValue().touint64()
	}
	for _, conn := range o.scene.m_connections {
		if connection.to == id && connection.from != 0 {
			// obj here should not be *Object, but an interface with GetObject and GetType
			obj := scene.m_object_map[connection.from].second.object
			if obj != nil && obj.getType() == typ {
				if property == nil || connection.property == property {
					if idx == 0 {
						return obj
					}
					idx--
				}
			}
		}
	}
	return nil
}

func (o *Object) resolveObjectLinkReverse(typ Type) *Object {
	var id uint64
	if element.getFirstProperty() != nil {
		id = element.getFirstProperty().getValue().touint64()
	}
	for _, conn := range o.scene.m_connections {
		if connection.to == id && connection.from != 0 {
			obj := scene.m_object_map[connection.to].second.object
			
			if obj != nil && obj.getType() == type { 
				return obj
			}
		}
	}
	return nil
}

func (o *Object) getParent() *Object{
	for _, con := range o.scene.m_connections{
		if con.from == o.id {
			obj := scene.m_object_map[con.to].second.object
			
			if (obj && obj.is_node) 
			{
				return obj
			}
		}
	}
	return nil
}

func (o *Object) getRotationOrder() RotationOrder {
	return RotationOrder(resolveEnumProperty(*this, "RotationOrder", (int) RotationOrder::EULER_XYZ))
}

func (o *Object) getRotationOffset() Vec3 {
	return resolveVec3Property(*this, "RotationOffset", {0, 0, 0})
}

func (o *Object) getRotationPivot() Vec3 {
	return resolveVec3Property(*this, "RotationPivot", {0, 0, 0})
}

func (o *Object) getPostRotation() Vec3 {
	return resolveVec3Property(*this, "PostRotation", {0, 0, 0})
}

func (o *Object) getScalingOffset() Vec3 {
	return resolveVec3Property(*this, "ScalingOffset", {0, 0, 0})
}

func (o *Object) getScalingPivot() Vec3 {
	return resolveVec3Property(*this, "ScalingPivot", {0, 0, 0})
}

func (o *Object) getPreRotation() Vec3 {
	return resolveVec3Property(*this, "PreRotation", {0, 0, 0})
}

func (o *Object) getLocalTranslation() Vec3 {
	return resolveVec3Property(*this, "Lcl Translation", {0, 0, 0})
}

func (o *Object) getLocalRotation() Vec3 {
	return resolveVec3Property(*this, "Lcl Rotation", {0, 0, 0})
}

func (o *Object) getLocalScaling() Vec3 {
	return resolveVec3Property(*this, "Lcl Scaling", {1, 1, 1})
}

func (o *Object) getGlobalTransform() Matrix {
	parent := o.getParent()
	if parent == nil { 
		return o.evalLocal(getLocalTranslation(), getLocalRotation())
	}

	return parent.getGlobalTransform() * o.evalLocal(o.getLocalTranslation(), o.getLocalRotation())
}

func (o *Object) getLocalTransform() Matrix {
	return o.evalLocalScaling(o.getLocalTranslation(), o.getLocalRotation(), o.getLocalScaling())
}

func (o *Object) evalLocal(translation, rotation *Vec3) Matrix {
	return evalLocalScaling(translation, rotation, o.getLocalScaling())
}

func (o *Object) evalLocalScaling(translation, rotation, scaling *Vec3) Matrix {
	rotation_pivot := o.getRotationPivot()
	scaling_pivot := o.getScalingPivot()
	rotation_order := o.getRotationOrder()

	s := makeIdentity()
	s.m[0] = scaling.x
	s.m[5] = scaling.y
	s.m[10] = scaling.z

	t := makeIdentity()
	setTranslation(translation, &t)

	r := getRotationMatrix(rotation, rotation_order)
	r_pre := getRotationMatrix(getPreRotation(), RotationOrder::EULER_XYZ)
	r_post_inv := getRotationMatrix(-getPostRotation(), RotationOrder::EULER_ZYX)

	r_off := makeIdentity()
	setTranslation(getRotationOffset(), &r_off)

	r_p := makeIdentity()
	setTranslation(rotation_pivot, &r_p)

	r_p_inv := makeIdentity()
	setTranslation(-rotation_pivot, &r_p_inv)

	s_off := makeIdentity()
	setTranslation(getScalingOffset(), &s_off)

	s_p := makeIdentity()
	setTranslation(scaling_pivot, &s_p)

	s_p_inv := makeIdentity()
	setTranslation(-scaling_pivot, &s_p_inv)

	// http://help.autodesk.com/view/FBX/2017/ENU/?guid=__files_GUID_10CDD63C_79C1_4F2D_BB28_AD2BE65A02ED_htm
	return t.Mul(r_off).Mul(r_p).Mul(r_pre).Mul(r).Mul(r_post_inv).Mul(r_p_inv).Mul(s_off).Mul(s_p).Mul(s).Mul(s_p_inv)
}

func (o *Object) isNode() bool {
	return o.is_node
}
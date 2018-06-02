package ofbx

type Object struct {
	id             uint64
	name           string
	element        *Element
	node_attribute Obj

	is_node bool
	scene   *Scene
}

type Obj interface {
	ID() uint64
	SetID(uint64)
	Name() string
	Element() *Element
	Node_attribute() Obj
	SetNodeAttribute(na Obj)
	IsNode() bool
	Scene() *Scene
	Type() Type
}

func (o *Object) ID() uint64 {
	return o.id
}
func (o *Object) SetID(i uint64) {
	o.id = i
}

func (o *Object) Name() string {
	return o.name
}
func (o *Object) Element() *Element {
	return o.element
}
func (o *Object) Node_attribute() Obj {
	return o.node_attribute
}
func (o *Object) SetNodeAttribute(na Obj) {
	o.node_attribute = na
}

func (o *Object) IsNode() bool {
	return o.is_node
}
func (o *Object) Scene() *Scene {
	return o.scene
}

func NewObject(scene *Scene, e *Element) *Object {
	o := &Object{
		scene:   scene,
		element: e,
		is_node: false,
	}
	if e.first_property != nil && e.first_property.next != nil {
		o.name = e.first_property.next.value.String()
	}
	return o
}

func resolveObjectLinkIndex(o Obj, idx int) Obj {
	var id uint64
	if o.Element().getFirstProperty() != nil {
		id = o.Element().getFirstProperty().getValue().touint64()
	}
	for _, conn := range o.Scene().m_connections {
		if conn.to == id && conn.from != 0 {
			obj := o.Scene().m_object_map[conn.from].object
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

func resolveObjectLink(o Obj, typ Type, property string, idx int) Obj {
	var id uint64
	if o.Element().getFirstProperty() != nil {
		id = o.Element().getFirstProperty().getValue().touint64()
	}
	for _, conn := range o.Scene().m_connections {
		if conn.to == id && conn.from != 0 {
			// obj here should not be *Object, but an interface with GetObject and GetType
			obj := o.Scene().m_object_map[conn.from].object
			if obj != nil && obj.Type() == typ {
				if property == "" || conn.property == property {
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

func resolveObjectLinkReverse(o Obj, typ Type) Obj {
	var id uint64
	if o.Element().getFirstProperty() != nil {
		id = o.Element().getFirstProperty().getValue().touint64()
	}
	for _, conn := range o.Scene().m_connections {
		if conn.to == id && conn.from != 0 {
			obj := o.Scene().m_object_map[conn.to].object

			if obj != nil && obj.Type() == typ {
				return obj
			}
		}
	}
	return nil
}

func getParent(o Obj) Obj {
	for _, con := range o.Scene().m_connections {
		if con.from == o.ID() {
			obj := o.Scene().m_object_map[con.to].object

			if obj != nil && obj.IsNode() {
				return obj
			}
		}
	}
	return nil
}

func getRotationOrder(o Obj) RotationOrder {
	return RotationOrder(resolveEnumProperty(o, "RotationOrder", int(EULER_XYZ)))
}

func getRotationOffset(o Obj) Vec3 {
	return resolveVec3Property(o, "RotationOffset", Vec3{})
}

func getRotationPivot(o Obj) Vec3 {
	return resolveVec3Property(o, "RotationPivot", Vec3{})
}

func getPostRotation(o Obj) Vec3 {
	return resolveVec3Property(o, "PostRotation", Vec3{})
}

func getScalingOffset(o Obj) Vec3 {
	return resolveVec3Property(o, "ScalingOffset", Vec3{})
}

func getScalingPivot(o Obj) Vec3 {
	return resolveVec3Property(o, "ScalingPivot", Vec3{})
}

func getPreRotation(o Obj) Vec3 {
	return resolveVec3Property(o, "PreRotation", Vec3{})
}

func getLocalTranslation(o Obj) Vec3 {
	return resolveVec3Property(o, "Lcl Translation", Vec3{})
}

func getLocalRotation(o Obj) Vec3 {
	return resolveVec3Property(o, "Lcl Rotation", Vec3{})
}

func getLocalScaling(o Obj) Vec3 {
	return resolveVec3Property(o, "Lcl Scaling", Vec3{1, 1, 1})
}

func getGlobalTransform(o Obj) Matrix {
	parent := getParent(o)
	if parent == nil {
		return evalLocal(o, getLocalTranslation(o), getLocalRotation(o))
	}

	return getGlobalTransform(parent).Mul(evalLocal(o, getLocalTranslation(o), getLocalRotation(o)))
}

func getLocalTransform(o Obj) Matrix {
	return evalLocalScaling(o, getLocalTranslation(o), getLocalRotation(o), getLocalScaling(o))
}

func evalLocal(o Obj, translation, rotation Vec3) Matrix {
	return evalLocalScaling(o, translation, rotation, getLocalScaling(o))
}

func evalLocalScaling(o Obj, translation, rotation, scaling Vec3) Matrix {
	rotation_pivot := getRotationPivot(o)
	scaling_pivot := getScalingPivot(o)
	rotation_order := getRotationOrder(o)

	s := makeIdentity()
	s.m[0] = scaling.X
	s.m[5] = scaling.Y
	s.m[10] = scaling.Z

	t := makeIdentity()
	setTranslation(translation, &t)

	r := getRotationMatrix(&rotation, rotation_order)
	pr := getPreRotation(o)
	r_pre := getRotationMatrix(&pr, EULER_XYZ)
	psr := getPostRotation(o)
	r_post_inv := getRotationMatrix(psr.Mul(-1), EULER_ZYX)

	r_off := makeIdentity()
	setTranslation(getRotationOffset(o), &r_off)

	r_p := makeIdentity()
	setTranslation(rotation_pivot, &r_p)

	r_p_inv := makeIdentity()
	setTranslation(*rotation_pivot.Mul(-1), &r_p_inv)

	s_off := makeIdentity()
	setTranslation(getScalingOffset(o), &s_off)

	s_p := makeIdentity()
	setTranslation(scaling_pivot, &s_p)

	s_p_inv := makeIdentity()
	setTranslation(*scaling_pivot.Mul(-1), &s_p_inv)

	// http://help.autodesk.com/view/FBX/2017/ENU/?guid=__files_GUID_10CDD63C_79C1_4F2D_BB28_AD2BE65A02ED_htm
	return t.Mul(r_off).Mul(r_p).Mul(r_pre).Mul(r).Mul(r_post_inv).Mul(r_p_inv).Mul(s_off).Mul(s_p).Mul(s).Mul(s_p_inv)
}

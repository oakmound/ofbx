package ofbx

import (
	"fmt"
	"io"

	"github.com/oakmound/oak/alg/floatgeom"
)

// Object is the top level general class in fbx
type Object struct {
	id             uint64
	name           string
	element        *Element
	node_attribute Obj

	is_node bool
	scene   *Scene
}

// Obj interface version of Object
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
	String() string
	stringPrefix(string) string
}

// ID returns the Object's integer id value
func (o *Object) ID() uint64 {
	return o.id
}

// SetID sets the Objects ID
func (o *Object) SetID(i uint64) {
	o.id = i
}

// Name gets the Objects Name
func (o *Object) Name() string {
	return o.name
}

// Element gets the Element on the Object
func (o *Object) Element() *Element {
	return o.element
}

// Node_attribute should be deprecated and in favor of exporting the attribute
func (o *Object) Node_attribute() Obj {
	return o.node_attribute
}

// SetNodeAttribute sets the attribute but should just exported field
func (o *Object) SetNodeAttribute(na Obj) {
	o.node_attribute = na
}

// IsNode ret[urns whether this is a node
func (o *Object) IsNode() bool {
	return o.is_node
}

// Scene returns the scene used for the object
func (o *Object) Scene() *Scene {
	return o.scene
}

func (o *Object) String() string {
	return o.stringPrefix("")
}
func (o *Object) stringPrefix(prefix string) string {
	s := "" //:= prefix //+ "Object: " + fmt.Sprintf("%d", o.id) + ", " + o.name
	if o.element != nil {
		s += o.element.stringPrefix(prefix)
	}
	if o.node_attribute != nil {
		if strn, ok := o.node_attribute.(fmt.Stringer); ok {
			s += ", node=" + strn.String()
		}
	}
	// if o.is_node {
	// 	s += "(is_node)"
	// }
	return s
}

// NewObject creates a new object
func NewObject(scene *Scene, e *Element) *Object {
	o := &Object{
		scene:   scene,
		element: e,
		is_node: false,
	}
	if prop := e.getProperty(1); prop != nil {
		o.name = prop.value.String()
	}
	return o
}

func resolveAllObjectLinks(o Obj) []Obj {
	return resolveObjectLinks(o, NOTYPE, []string{""})
}

func resolveObjectLinkIndex(o Obj, idx int) Obj {
	return resolveObjectLink(o, NOTYPE, "", idx)
}

func resolveObjectLink(o Obj, typ Type, property string, idx int) Obj {
	id := o.ID()
	for _, conn := range o.Scene().connections {
		if conn.to == id && conn.from != 0 {
			obj := o.Scene().objectMap[conn.from].object
			if obj != nil && (obj.Type() == typ || typ == NOTYPE) {
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

func resolveObjectLinks(o Obj, typ Type, properties []string) []Obj {
	id := o.ID()
	out := make([]Obj, 0)
	for _, conn := range o.Scene().connections {
		if conn.to == id && conn.from != 0 {
			obj := o.Scene().objectMap[conn.from].object
			if obj != nil && (obj.Type() == typ || typ == NOTYPE) {
				for _, prop2 := range properties {
					if prop2 == "" || conn.property == prop2 {
						out = append(out, obj)
						break
					}
				}
			}
		}
	}
	return out
}

func resolveObjectLinkReverse(o Obj, typ Type) Obj {
	var id uint64
	if prop := o.Element().getProperty(0); prop != nil {
		rdr := prop.value
		rdr.Seek(0, io.SeekStart)
		id = rdr.touint64()
	}
	for _, conn := range o.Scene().connections {
		//fmt.Println("Connection iterated", id, conn.from, conn.to)
		if conn.from == id && conn.to != 0 {
			obj := o.Scene().objectMap[conn.to].object
			if obj != nil && obj.Type() == typ {
				return obj
			}
		}
	}
	return nil
}

func getParent(o Obj) Obj {
	for _, con := range o.Scene().connections {
		if con.from == o.ID() {
			obj := o.Scene().objectMap[con.to].object

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

func getRotationOffset(o Obj) floatgeom.Point3 {
	return resolveVec3Property(o, "RotationOffset", floatgeom.Point3{})
}

func getRotationPivot(o Obj) floatgeom.Point3 {
	return resolveVec3Property(o, "RotationPivot", floatgeom.Point3{})
}

func getPostRotation(o Obj) floatgeom.Point3 {
	return resolveVec3Property(o, "PostRotation", floatgeom.Point3{})
}

func getScalingOffset(o Obj) floatgeom.Point3 {
	return resolveVec3Property(o, "ScalingOffset", floatgeom.Point3{})
}

func getScalingPivot(o Obj) floatgeom.Point3 {
	return resolveVec3Property(o, "ScalingPivot", floatgeom.Point3{})
}

func getPreRotation(o Obj) floatgeom.Point3 {
	return resolveVec3Property(o, "PreRotation", floatgeom.Point3{})
}

func getLocalTranslation(o Obj) floatgeom.Point3 {
	return resolveVec3Property(o, "Lcl Translation", floatgeom.Point3{})
}

func getLocalRotation(o Obj) floatgeom.Point3 {
	return resolveVec3Property(o, "Lcl Rotation", floatgeom.Point3{})
}

func getLocalScaling(o Obj) floatgeom.Point3 {
	return resolveVec3Property(o, "Lcl Scaling", floatgeom.Point3{1, 1, 1})
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

func evalLocal(o Obj, translation, rotation floatgeom.Point3) Matrix {
	return evalLocalScaling(o, translation, rotation, getLocalScaling(o))
}

func evalLocalScaling(o Obj, translation, rotation, scaling floatgeom.Point3) Matrix {
	rotation_pivot := getRotationPivot(o)
	scaling_pivot := getScalingPivot(o)
	rotation_order := getRotationOrder(o)

	s := makeIdentity()
	s.m[0] = scaling.X()
	s.m[5] = scaling.Y()
	s.m[10] = scaling.Z()

	t := makeIdentity()
	setTranslation(translation, &t)

	r := getRotationMatrix(rotation, rotation_order)
	pr := getPreRotation(o)
	r_pre := getRotationMatrix(pr, EULER_XYZ)
	psr := getPostRotation(o)
	r_post_inv := getRotationMatrix(psr.MulConst(-1), EULER_ZYX)

	r_off := makeIdentity()
	setTranslation(getRotationOffset(o), &r_off)

	r_p := makeIdentity()
	setTranslation(rotation_pivot, &r_p)

	r_p_inv := makeIdentity()
	setTranslation(rotation_pivot.MulConst(-1), &r_p_inv)

	s_off := makeIdentity()
	setTranslation(getScalingOffset(o), &s_off)

	s_p := makeIdentity()
	setTranslation(scaling_pivot, &s_p)

	s_p_inv := makeIdentity()
	setTranslation(scaling_pivot.MulConst(-1), &s_p_inv)

	// http://help.autodesk.com/view/FBX/2017/ENU/?guid=__files_GUID_10CDD63C_79C1_4F2D_BB28_AD2BE65A02ED_htm
	return t.Mul(r_off).Mul(r_p).Mul(r_pre).Mul(r).Mul(r_post_inv).Mul(r_p_inv).Mul(s_off).Mul(s_p).Mul(s).Mul(s_p_inv)
}

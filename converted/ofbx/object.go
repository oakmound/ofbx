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

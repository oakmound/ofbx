package ofbx

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

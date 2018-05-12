package ofbx

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

package ofbx

type AnimationStack struct {
	Object
}

func NewAnimationStack(scene *Scene, element *Element) *AnimationStack {
	o := *NewObject(scene, element)
	return &AnimationStack{o}
}

func (as *AnimationStack) Type() Type {
	return ANIMATION_STACK
}

func (as *AnimationStack) getLayer(index int) *AnimationLayer {
	return resolveObjectLinkIndex(as, index).(*AnimationLayer)
}

type AnimationLayer struct {
	Object
	CurveNodes []*AnimationCurveNode
}

func NewAnimationLayer(scene *Scene, element *Element) *AnimationLayer {
	o := *NewObject(scene, element)
	return &AnimationLayer{o, nil}
}

func (as *AnimationLayer) Type() Type {
	return ANIMATION_LAYER
}

func (as *AnimationLayer) getCurveNode(bone Obj, property string) *AnimationCurveNode {
	for _, node := range as.CurveNodes {
		if node.boneLinkProp == property && node.Bone == bone {
			return node
		}
	}
	return nil
}

func (as *AnimationLayer) String() string {
	s := "AnimationLayer: " + as.Object.String()
	if len(as.CurveNodes) != 0 {
		s += " curveNodes="
		for _, curve := range as.CurveNodes {
			s += " " + curve.String()
		}
	}
	return s

}

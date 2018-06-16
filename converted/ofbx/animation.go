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
	curve_nodes []*AnimationCurveNode
}

func NewAnimationLayer(scene *Scene, element *Element) *AnimationLayer {
	o := *NewObject(scene, element)
	return &AnimationLayer{o, nil}
}

func (as *AnimationLayer) Type() Type {
	return ANIMATION_LAYER
}

func (as *AnimationLayer) getCurveNodeIndex(index int) *AnimationCurveNode {
	return as.curve_nodes[index]
}

func (as *AnimationLayer) getCurveNode(bone Obj, property string) *AnimationCurveNode {
	for _, node := range as.curve_nodes {
		if node.bone_link_property == property && node.bone == bone {
			return node
		}
	}
	return nil
}

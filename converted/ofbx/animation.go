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


AnimationStack::AnimationStack(const Scene& _scene, const IElement& _element)
	: Object(_scene, _element)
{
}


AnimationLayer::AnimationLayer(const Scene& _scene, const IElement& _element)
	: Object(_scene, _element)
{
}


struct AnimationStackImpl : AnimationStack
{
	AnimationStackImpl(const Scene& _scene, const IElement& _element)
		: AnimationStack(_scene, _element)
	{
	}


	const AnimationLayer* getLayer(int index) const override
	{
		return resolveObjectLink<AnimationLayer>(index);
	}


	Type getType() const override { return Type::ANIMATION_STACK; }
};

struct AnimationLayerImpl : AnimationLayer
{
	AnimationLayerImpl(const Scene& _scene, const IElement& _element)
		: AnimationLayer(_scene, _element)
	{
	}


	Type getType() const override { return Type::ANIMATION_LAYER; }


	const AnimationCurveNode* getCurveNode(int index) const override
	{
		if (index >= curve_nodes.size() || index < 0) return nullptr;
		return curve_nodes[index];
	}


	const AnimationCurveNode* getCurveNode(const Object& bone, const char* prop) const override
	{
		for (const AnimationCurveNodeImpl* node : curve_nodes)
		{
			if (node.bone_link_property == prop && node.bone == &bone) return node;
		}
		return nullptr;
	}


	std::vector<AnimationCurveNodeImpl*> curve_nodes;
};
package ofbx

type NodeAttribute struct {
	Object
}

func NewNodeAttribute(scene *Scene, element *IElement) *NodeAttribute {
	return nil
}

func (na *NodeAttribute) Type() Type {
	return NODE_ATTRIBUTE
}

func (na NodeAttribute) getAttributeType() DataView {
	return DataView{}
}



// From CPP



static OptionalError<Object*> parseNodeAttribute(const Scene& scene, const Element& element)
{
	NodeAttributeImpl* obj = new NodeAttributeImpl(scene, element);
	const Element* type_flags = findChild(element, "TypeFlags");
	if (type_flags && type_flags.first_property)
	{
		obj.attribute_type = type_flags.first_property.value;
	}
	return obj;
}

NodeAttribute::NodeAttribute(const Scene& _scene, const IElement& _element)
	: Object(_scene, _element)
{
}


struct NodeAttributeImpl : NodeAttribute
{
	NodeAttributeImpl(const Scene& _scene, const IElement& _element)
		: NodeAttribute(_scene, _element)
	{
	}
	Type getType() const override { return Type::NODE_ATTRIBUTE; }
	DataView getAttributeType() const override { return attribute_type; }


	DataView attribute_type;
};
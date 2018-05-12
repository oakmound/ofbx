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

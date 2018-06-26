package ofbx

type NodeAttribute struct {
	Object
	attribute_type *DataView
}

func NewNodeAttribute(scene *Scene, element *Element) *NodeAttribute {
	o := *NewObject(scene, element)

	return &NodeAttribute{o, nil}
}

func (na *NodeAttribute) Type() Type {
	return NODE_ATTRIBUTE
}

func (na NodeAttribute) getAttributeType() *DataView {
	return na.attribute_type
}

func parseNodeAttribute(scene *Scene, element *Element) (*NodeAttribute, error) {
	na := NewNodeAttribute(scene, element)
	typeFlags := findChildProperty(element, "TypeFlags")
	if typeFlags != nil {
		na.attribute_type = typeFlags.value
	}
	return na, nil
}

// From CPP
func (na *NodeAttribute) getType() Type {
	return na.Type()
}

package ofbx

type NodeAttribute struct {
	Object
	attribute_type DataView
}

func NewNodeAttribute(scene *Scene, element *Element) *NodeAttribute {
	return nil
}

func (na *NodeAttribute) Type() Type {
	return NODE_ATTRIBUTE
}

func (na NodeAttribute) getAttributeType() DataView {
	return na.attribute_type
}

func parseNodeAttribute(scene *Scene, element *Element) (*NodeAttribute, error) {
	na := NewNodeAttribute(scene, element)
	type_flags := findChild(element, "TypeFlags")
	if type_flags != nil && type_flags.first_property {
		na.attribute_type = type_flags.first_property.value
	}
	return na, nil
}

// From CPP
func (na *NodeAttribute) getType() Type {
	return na.Type()
}

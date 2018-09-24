package ofbx

// NodeAttribute is an formattable Attribute on a Node
type NodeAttribute struct {
	Object
	Attribute *DataView
}

// NewNodeAttribute creates a stub NodeAttribute
func NewNodeAttribute(scene *Scene, element *Element) *NodeAttribute {
	o := *NewObject(scene, element)
	return &NodeAttribute{o, nil}
}

// Type returns NODE_ATTRIBUTE
func (na *NodeAttribute) Type() Type {
	return NODE_ATTRIBUTE
}

func parseNodeAttribute(scene *Scene, element *Element) (*NodeAttribute, error) {
	na := NewNodeAttribute(scene, element)
	typeFlags := findChildProperty(element, "TypeFlags")
	if len(typeFlags) != 0 {
		na.Attribute = typeFlags[0].value
	}
	return na, nil
}

func (na *NodeAttribute) String() string {
	s := "NodeAttribute: " + na.Object.String()
	s += ", Attribute=" + na.Attribute.String()
	return s
}

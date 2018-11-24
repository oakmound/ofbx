package threefbx

func typedProperty(name string, root, toType *Node) bool {

	switch toType.name {
	case "Transform", "TransformLink", "Indexes", "Weights", "Vertices", "PolygonVertexIndex":
		// fmt.Println(toType.name ," has props ", toType.propertyList[0].Payload)
		root.props[toType.name] = toType.propertyList[0]

	case "LayerElementMaterial":
		// fmt.Println(toType.name, " looks like  ", toType)
		root.props[toType.name] = NodeProperty(toType)
	default:
		return false
	}

	return true
}

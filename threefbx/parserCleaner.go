package threefbx

func typedProperty(name string, root, toType *Node) bool {

	switch toType.name {
	case "Transform", "TransformLink", "Indexes", "Weights", "Vertices", "PolygonVertexIndex",
		"MappingInformationType", "ReferenceInformationType", "Materials", "Normals",
		"KeyTime":
		// fmt.Println(toType.name ," has props ", toType.propertyList[0].Payload)
		root.props[toType.name] = toType.propertyList[0]

	case "LayerElementMaterial", "LayerElementNormal":
		// fmt.Println(toType.name, " looks like  ", toType)
		root.props[toType.name] = NodeProperty(toType)
	case "PoseNode":
		if _, ok := root.props["PoseNode"]; !ok {
			root.props["PoseNode"] = Property{Payload: []*Node{toType}}
		} else {
			nodes := root.props["PoseNode"].Payload.([]*Node)
			root.props["PoseNode"] = Property{Payload: append(nodes, toType)}
		}
	case "KeyValueFloat":

		temp := toType.propertyList[0].Payload.([]float32)

		mat := make([]float64, len(temp))

		for i, c := range temp {
			mat[i] = float64(c)
		}
		root.props[toType.name] = Property{Payload: mat}

	case "Matrix":
		floats := toType.propertyList[0].Payload.([]float64)
		mat := Mat4FromSlice(floats)
		root.props[toType.name] = Property{Payload: mat}

	default:
		return false
	}

	return true
}

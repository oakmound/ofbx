package ofbx

import "github.com/oakmound/oak/alg/floatgeom"

func resolveEnumProperty(object Obj, name string, default_value int) int {
	element := resolveProperty(object, name)
	if element == nil {
		return default_value
	}
	x := element.getProperty(4)
	if x == nil {
		return default_value
	}

	return int(x.value.toInt32())
}

func resolveVec3Property(object Obj, name string, default_value floatgeom.Point3) floatgeom.Point3 {
	element := resolveProperty(object, name)
	if element == nil {
		return default_value
	}
	if len(element.properties) < 6 {
		return default_value
	}

	return floatgeom.Point3{
		element.getProperty(4).value.toDouble(),
		element.getProperty(5).value.toDouble(),
		element.getProperty(6).value.toDouble(),
	}
}

func splatVec2(mapping VertexDataMapping, data []floatgeom.Point2, indices []int, original_indices []int) (out []floatgeom.Point2) {
	if mapping == BY_POLYGON_VERTEX {
		if len(indices) == 0 {
			out = make([]floatgeom.Point2, len(data))
			copy(out, data)
		} else {
			out = make([]floatgeom.Point2, len(indices))
			for i := 0; i < len(indices); i++ {
				if indices[i] < len(data) {
					out[i] = data[indices[i]]
				} else {
					out[i] = floatgeom.Point2{}
				}
			}
		}
	} else if mapping == BY_VERTEX {
		//  v0  v1 ...
		// uv0 uv1 ...

		out := make([]floatgeom.Point2, len(original_indices))

		for i := 0; i < len(original_indices); i++ {
			idx := original_indices[i]
			if idx < 0 {
				idx = -idx - 1
			}
			if idx < len(data) {
				out[i] = data[idx]
			} else {
				out[i] = floatgeom.Point2{}
			}
		}
	} else {
		panic("oh no")
	}
	return out
}

func splatVec3(mapping VertexDataMapping, data []floatgeom.Point3, indices []int, original_indices []int) (out []floatgeom.Point3) {
	if mapping == BY_POLYGON_VERTEX {
		if len(indices) == 0 {
			out = make([]floatgeom.Point3, len(data))
			copy(out, data)
		} else {
			out = make([]floatgeom.Point3, len(indices))
			for i := 0; i < len(indices); i++ {
				if indices[i] < len(data) {
					out[i] = data[indices[i]]
				} else {
					out[i] = floatgeom.Point3{}
				}
			}
		}
	} else if mapping == BY_VERTEX {
		//  v0  v1 ...
		// uv0 uv1 ...

		out := make([]floatgeom.Point3, len(original_indices))

		for i := 0; i < len(original_indices); i++ {
			idx := original_indices[i]
			if idx < 0 {
				idx = -idx - 1
			}
			if idx < len(data) {
				out[i] = data[idx]
			} else {
				out[i] = floatgeom.Point3{}
			}
		}
	} else {
		panic("oh no")
	}
	return out
}

func splatVec4(mapping VertexDataMapping, data []floatgeom.Point4, indices []int, original_indices []int) (out []floatgeom.Point4) {
	if mapping == BY_POLYGON_VERTEX {
		if len(indices) == 0 {
			out = make([]floatgeom.Point4, len(data))
			copy(out, data)
		} else {
			out = make([]floatgeom.Point4, len(indices))
			for i := 0; i < len(indices); i++ {
				if indices[i] < len(data) {
					out[i] = data[indices[i]]
				} else {
					out[i] = floatgeom.Point4{}
				}
			}
		}
	} else if mapping == BY_VERTEX {
		//  v0  v1 ...
		// uv0 uv1 ...

		out := make([]floatgeom.Point4, len(original_indices))

		for i := 0; i < len(original_indices); i++ {
			idx := original_indices[i]
			if idx < 0 {
				idx = -idx - 1
			}
			if idx < len(data) {
				out[i] = data[idx]
			} else {
				out[i] = floatgeom.Point4{}
			}
		}
	} else {
		panic("oh no")
	}
	return out
}

func remapVec2(out *[]floatgeom.Point2, m []int) {
	if len(*out) == 0 {
		return
	}

	old := make([]floatgeom.Point2, len(*out))
	copy(old, *out)
	for i := 0; i < len(m); i++ {
		if m[i] < len(old) {
			*out = append(*out, old[m[i]])
		} else {
			*out = append(*out, floatgeom.Point2{})
		}
	}
}

func remapVect3(out *[]floatgeom.Point3, m []int) {
	if len(*out) == 0 {
		return
	}

	old := make([]floatgeom.Point3, len(*out))
	copy(old, *out)
	for i := 0; i < len(m); i++ {
		if m[i] < len(old) {
			*out = append(*out, old[m[i]])
		} else {
			*out = append(*out, floatgeom.Point3{})
		}
	}
}

func remapVec4(out *[]floatgeom.Point4, m []int) {
	if len(*out) == 0 {
		return
	}

	old := make([]floatgeom.Point4, len(*out))
	copy(old, *out)
	for i := 0; i < len(m); i++ {
		if m[i] < len(old) {
			*out = append(*out, old[m[i]])
		} else {
			*out = append(*out, floatgeom.Point4{})
		}
	}
}

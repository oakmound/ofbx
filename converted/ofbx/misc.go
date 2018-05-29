package ofbx

func resolveEnumProperty(object *Object, name string, default_value int) int {
	element := resolveProperty(object, name)
	if element == nil {
		return default_value
	}
	x := element.getProperty(4)
	if x == nil {
		return default_value
	}

	return x.value.toInt()
}

func resolveVec3Property(object *Object, name string, default_value Vec3) Vec3 {
	element := resolveProperty(object, name)
	if element == nil {
		return default_value
	}
	x := element.getProperty(4)
	if x == nil || x.next == nil || x.next.next == nil {
		return default_value
	}

	return Vec3{
		x.value.toDouble(),
		x.next.value.toDouble(),
		x.next.next.value.toDouble(),
	}
}

type LimbNode struct {
	Object
}

func NewLimbNode(scene *Scene, element *Element) *LimbNode {
	ln := &LimbNode{}
	ln.Object = *NewObject(scene, element)
	ln.is_node = true
	return ln
}

func (ln *LimbNode) getType() Type {
	return LIMB_NODE
}

type Null struct {
	Object
}

func NewNull(scene *Scene, element *Element) *Null {
	n := &Null{}
	n.Object = *NewObject(scene, element)
	n.is_node = true
	return n
}

func (n *Null) getType() Type {
	return NULL_NODE
}

type Root struct {
	Object
	name string
}

func NewRoot(scene *Scene, element *Element) *Root {
	r := &Root{}
	r.Object = *NewObject(scene, element)
	r.name = "RootNode" // might not need this
	r.is_node = true
	return r
}

func (r *Root) getType() Type {
	return ROOT
}

func splatVec2(mapping VertexDataMapping, data []Vec2, indices []int, original_indices []int) (out []Vec2) {
	if mapping == BY_POLYGON_VERTEX {
		if len(indices) == 0 {
			out = make([]Vec2, len(data))
			copy(data, out)
		} else {
			out = make([]Vec2, len(indices))
			for i := 0; i < len(indices); i++ {
				if indices[i] < len(data) {
					out[i] = data[indices[i]]
				} else {
					out[i] = Vec2{}
				}
			}
		}
	} else if mapping == BY_VERTEX {
		//  v0  v1 ...
		// uv0 uv1 ...

		out := make([]Vec2, len(original_indices))

		for i := 0; i < len(original_indices); i++ {
			idx := original_indices[i]
			if idx < 0 {
				idx = -idx - 1
			}
			if idx < len(data) {
				out[i] = data[idx]
			} else {
				out[i] = Vec2{}
			}
		}
	} else {
		panic("oh no")
	}
	return out
}

func splatVec3(mapping VertexDataMapping, data []Vec3, indices []int, original_indices []int) (out []Vec3) {
	if mapping == BY_POLYGON_VERTEX {
		if len(indices) == 0 {
			out = make([]Vec3, len(data))
			copy(data, out)
		} else {
			out = make([]Vec3, len(indices))
			for i := 0; i < len(indices); i++ {
				if indices[i] < len(data) {
					out[i] = data[indices[i]]
				} else {
					out[i] = Vec3{}
				}
			}
		}
	} else if mapping == BY_VERTEX {
		//  v0  v1 ...
		// uv0 uv1 ...

		out := make([]Vec3, len(original_indices))

		for i := 0; i < len(original_indices); i++ {
			idx := original_indices[i]
			if idx < 0 {
				idx = -idx - 1
			}
			if idx < len(data) {
				out[i] = data[idx]
			} else {
				out[i] = Vec3{}
			}
		}
	} else {
		panic("oh no")
	}
	return out
}

func splatVec4(mapping VertexDataMapping, data []Vec4, indices []int, original_indices []int) (out []Vec4) {
	if mapping == BY_POLYGON_VERTEX {
		if len(indices) == 0 {
			out = make([]Vec4, len(data))
			copy(data, out)
		} else {
			out = make([]Vec4, len(indices))
			for i := 0; i < len(indices); i++ {
				if indices[i] < len(data) {
					out[i] = data[indices[i]]
				} else {
					out[i] = Vec4{}
				}
			}
		}
	} else if mapping == BY_VERTEX {
		//  v0  v1 ...
		// uv0 uv1 ...

		out := make([]Vec4, len(original_indices))

		for i := 0; i < len(original_indices); i++ {
			idx := original_indices[i]
			if idx < 0 {
				idx = -idx - 1
			}
			if idx < len(data) {
				out[i] = data[idx]
			} else {
				out[i] = Vec4{}
			}
		}
	} else {
		panic("oh no")
	}
	return out
}

func remapVec2(out *[]Vec2, m []int) {
	if len(*out) == 0 {
		return
	}

	old := make([]Vec2, len(*out))
	copy(old, *out)
	for i := 0; i < len(m); i++ {
		if m[i] < len(old) {
			*out = append(*out, old[m[i]])
		} else {
			*out = append(*out, Vec2{})
		}
	}
}

func remapVec3(out *[]Vec3, m []int) {
	if len(*out) == 0 {
		return
	}

	old := make([]Vec3, len(*out))
	copy(old, *out)
	for i := 0; i < len(m); i++ {
		if m[i] < len(old) {
			*out = append(*out, old[m[i]])
		} else {
			*out = append(*out, Vec3{})
		}
	}
}

func remapVec4(out *[]Vec4, m []int) {
	if len(*out) == 0 {
		return
	}

	old := make([]Vec4, len(*out))
	copy(old, *out)
	for i := 0; i < len(m); i++ {
		if m[i] < len(old) {
			*out = append(*out, old[m[i]])
		} else {
			*out = append(*out, Vec4{})
		}
	}
}

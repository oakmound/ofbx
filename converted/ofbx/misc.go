package ofbx

func resolveEnumProperty(object *Object name string, default_value int) int {
	element := resolveProperty(object, name).(*Element)
	if element == nil {
		return default_value
	}
	x := element.getProperty(4).(*Property)
	if x == nil { 
		return default_value
	}

	return x.value.toInt()
}

func resolveVec3Property(object *Object, name string, default_value *Vec3) Vec3 {
	element := resolveProperty(object, name).(*Element)
	if element == nil { 
		return default_value
	}
	x := element.getProperty(4).(*Property)
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

func NewLimbNode(scene *Scene, element *Element) {
	ln := &LimbNode{}
	ln.Object = NewObject(scene, element)
	ln.is_node = true
	return ln
}

func (ln *LimbNode) getType() Type {
	return LIMB_NODE
}

type Null struct {
	Object
}

func NewNull(scene *Scene, element *Element) {
	n := &Null{}
	n.Object = NewObject(scene, element)
	n.is_node = true
	return ln
}

func (n *Null) getType() Type {
	return NULL_NODE
}

type Root struct {
	Object
}

func NewRoot(scene *Scene, element *Element) {
	r := &Root{}
	r.Object = NewObject(scene, element)
	r.name = "RootNode" // might not need this
	r.is_node = true
	return ln
}

func (r *Root) getType() Type {
	return Root_NODE
}

func splat(out []T, mapping VertexDataMapping, data []T, indices []int, original_indices []int) {
	if mapping == BY_POLYGON_VERTEX {
		if indices.empty() {
			out = make([]T, len(data))
			memcpy(&(*out)[0], &data[0], sizeof(data[0]) * data.size())
		} else {
			out = make([]T, len(indices))
			for i := 0; i < len(indices); i++ {
				if indices[i] < len(data)) {
					(*out)[i] = data[indices[i]]
				} else {
					(*out)[i] = T{}
				}
			}
		}
	} else if mapping == BY_VERTEX {
		//  v0  v1 ...
		// uv0 uv1 ...

		out := make([]T, len(original_indices))

		for i := 0; i < len(original_indices); i++ {
			idx := original_indices[i]
			if idx < 0 {
				idx = -idx - 1
			}
			if idx < len(data) {
				(*out)[i] = data[idx]
			} else { 
				(*out)[i] = T{}
			}
		}
	} else {
		panic()
	}
}

func remap([]T out, []int m) {
	if out.empty() {
		return
	}

	old := make([]T, len(out))
	copy(old, out)
	for i := 0; i < len(m); i++ {
		if m[i] < len(old)) {
			out.push_back(old[m[i]])
		}
		else {
			out.push_back(T())
		}
	}
}

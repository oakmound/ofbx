package ofbx

static int resolveEnumProperty(const Object& object, const char* name, int default_value) {
	Element* element = (Element*)resolveProperty(object, name);
	if (!element) return default_value;
	Property* x = (Property*)element.getProperty(4);
	if (!x) return default_value;

	return x.value.toInt();
}

static Vec3 resolveVec3Property(const Object& object, const char* name, const Vec3& default_value) {
	Element* element = (Element*)resolveProperty(object, name);
	if (!element) return default_value;
	Property* x = (Property*)element.getProperty(4);
	if (!x || !x.next || !x.next.next) return default_value;

	return {x.value.toDouble(), x.next.value.toDouble(), x.next.next.value.toDouble()};
}

Object::Object(const Scene& _scene, const IElement& _element)
	: scene(_scene)
	, element(_element)
	, is_node(false)
	, node_attribute(nullptr) {
	auto& e = (Element&)_element;
	if (e.first_property && e.first_property.next) {
		e.first_property.next.value.toString(name);
	}
	else {
		name[0] = '\0';
	}
}

struct LimbNodeImpl : Object {
	LimbNodeImpl(const Scene& _scene, const IElement& _element)
		: Object(_scene, _element) {
		is_node = true;
	}
	Type getType() const override { return Type::LIMB_NODE; }
};

struct NullImpl : Object {
	NullImpl(const Scene& _scene, const IElement& _element)
		: Object(_scene, _element) {
		is_node = true;
	}
	Type getType() const override { return Type::NULL_NODE; }
};

struct Root : Object {
	Root(const Scene& _scene, const IElement& _element)
		: Object(_scene, _element) {
		copyString(name, "RootNode");
		is_node = true;
	}
	Type getType() const override { return Type::ROOT; }
};

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

template <typename T> static void remap(std::vector<T>* out, const std::vector<int>& map) {
	if (out.empty()) return;

	std::vector<T> old;
	old.swap(*out);
	int old_size = (int)old.size();
	for (int i = 0, c = (int)map.size(); i < c; ++i) {
		if(map[i] < old_size) out.push_back(old[map[i]]);
		else out.push_back(T());
	}
}

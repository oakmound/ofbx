// We might care starting here... but probs not

struct Cursor {
	const uint8* current;
	const uint8* begin;
	const uint8* end;
};

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

template <typename T>
static void splat(std::vector<T>* out,
	GeometryImpl::VertexDataMapping mapping,
	const std::vector<T>& data,
	const std::vector<int>& indices,
	const std::vector<int>& original_indices) {
	assert(out);
	assert(!data.empty());

	if (mapping == GeometryImpl::BY_POLYGON_VERTEX) {
		if (indices.empty()) {
			out.resize(data.size());
			memcpy(&(*out)[0], &data[0], sizeof(data[0]) * data.size());
		}
		else {
			out.resize(indices.size());
			int data_size = (int)data.size();
			for (int i = 0, c = (int)indices.size(); i < c; ++i) {
				if(indices[i] < data_size) (*out)[i] = data[indices[i]];
				else (*out)[i] = T();
			}
		}
	}
	else if (mapping == GeometryImpl::BY_VERTEX) {
		//  v0  v1 ...
		// uv0 uv1 ...
		assert(indices.empty());

		out.resize(original_indices.size());

		int data_size = (int)data.size();
		for (int i = 0, c = (int)original_indices.size(); i < c; ++i) {
			int idx = original_indices[i];
			if (idx < 0) idx = -idx - 1;
			if(idx < data_size) (*out)[i] = data[idx];
			else (*out)[i] = T();
		}
	}
	else {
		assert(false);
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

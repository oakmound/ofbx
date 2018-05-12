
func fbxTimeToSeconds(value int64) float64{
	return float64(value)/float64(46186158000)
}
func secondsToFbxTime(value float64) int64{
	return int64(value /46186158000)
}

//-------------------------------------

struct Header
{
	uint8 magic[21];
	uint8 reserved[2];
	uint32 version;
};


// We might care starting here... but probs not


struct Cursor
{
	const uint8* current;
	const uint8* begin;
	const uint8* end;
};


static void setTranslation(const Vec3& t, Matrix* mtx)
{
	mtx.m[12] = t.x;
	mtx.m[13] = t.y;
	mtx.m[14] = t.z;
}

template <int SIZE> static bool copyString(char (&destination)[SIZE], const char* source)
{
	const char* src = source;
	char* dest = destination;
	int length = SIZE;
	if (!src) return false;

	while (*src && length > 1)
	{
		*dest = *src;
		--length;
		++dest;
		++src;
	}
	*dest = 0;
	return *src == '\0';
}


uint64 DataView::touint64() const
{
	if (is_binary)
	{
		assert(end - begin == sizeof(uint64));
		return *(uint64*)begin;
	}
	static_assert(sizeof(unsigned long long) >= sizeof(uint64), "can't use strtoull");
	return strtoull((const char*)begin, nullptr, 10);
}


int64 DataView::toint64() const
{
	if (is_binary)
	{
		assert(end - begin == sizeof(int64));
		return *(int64*)begin;
	}
	static_assert(sizeof(long long) >= sizeof(int64), "can't use atoll");
	return atoll((const char*)begin);
}


int DataView::toInt() const
{
	if (is_binary)
	{
		assert(end - begin == sizeof(int));
		return *(int*)begin;
	}
	return atoi((const char*)begin);
}


uint32 DataView::touint32() const
{
	if (is_binary)
	{
		assert(end - begin == sizeof(uint32));
		return *(uint32*)begin;
	}
	return (uint32)atoll((const char*)begin);
}


double DataView::toDouble() const
{
	if (is_binary)
	{
		assert(end - begin == sizeof(double));
		return *(double*)begin;
	}
	return atof((const char*)begin);
}


float DataView::toFloat() const
{
	if (is_binary)
	{
		assert(end - begin == sizeof(float));
		return *(float*)begin;
	}
	return (float)atof((const char*)begin);
}


bool DataView::operator==(const char* rhs) const
{
	const char* c = rhs;
	const char* c2 = (const char*)begin;
	while (*c && c2 != (const char*)end)
	{
		if (*c != *c2) return 0;
		++c;
		++c2;
	}
	return c2 == (const char*)end && *c == '\0';
}


struct Property;
template <typename T> static bool parseArrayRaw(const Property& property, T* out, int max_size);
template <typename T> static bool parseBinaryArray(const Property& property, std::vector<T>* out);


struct Property : IElementProperty
{
	~Property() { delete next; }
	Type getType() const override { return (Type)type; }
	IElementProperty* getNext() const override { return next; }
	DataView getValue() const override { return value; }
	int getCount() const override
	{
		assert(type == ARRAY_DOUBLE || type == ARRAY_INT || type == ARRAY_FLOAT || type == ARRAY_LONG);
		if (value.is_binary)
		{
			return int(*(uint32*)value.begin);
		}
		return count;
	}

	bool getValues(double* values, int max_size) const override { return parseArrayRaw(*this, values, max_size); }

	bool getValues(float* values, int max_size) const override { return parseArrayRaw(*this, values, max_size); }

	bool getValues(uint64* values, int max_size) const override { return parseArrayRaw(*this, values, max_size); }
	
	bool getValues(int64* values, int max_size) const override { return parseArrayRaw(*this, values, max_size); }

	bool getValues(int* values, int max_size) const override { return parseArrayRaw(*this, values, max_size); }

	int count;
	uint8 type;
	DataView value;
	Property* next = nullptr;
};


struct Element : IElement
{
	IElement* getFirstChild() const override { return child; }
	IElement* getSibling() const override { return sibling; }
	DataView getID() const override { return id; }
	IElementProperty* getFirstProperty() const override { return first_property; }
	IElementProperty* getProperty(int idx) const
	{
		IElementProperty* prop = first_property;
		for (int i = 0; i < idx; ++i)
		{
			if (prop == nullptr) return nullptr;
			prop = prop.getNext();
		}
		return prop;
	}

	DataView id;
	Element* child = nullptr;
	Element* sibling = nullptr;
	Property* first_property = nullptr;
};


static const Element* findChild(const Element& element, const char* id)
{
	Element* const* iter = &element.child;
	while (*iter)
	{
		if ((*iter).id == id) return *iter;
		iter = &(*iter).sibling;
	}
	return nullptr;
}


static IElement* resolveProperty(const Object& obj, const char* name)
{
	const Element* props = findChild((const Element&)obj.element, "Properties70");
	if (!props) return nullptr;

	Element* prop = props.child;
	while (prop)
	{
		if (prop.first_property && prop.first_property.value == name)
		{
			return prop;
		}
		prop = prop.sibling;
	}
	return nullptr;
}


static int resolveEnumProperty(const Object& object, const char* name, int default_value)
{
	Element* element = (Element*)resolveProperty(object, name);
	if (!element) return default_value;
	Property* x = (Property*)element.getProperty(4);
	if (!x) return default_value;

	return x.value.toInt();
}


static Vec3 resolveVec3Property(const Object& object, const char* name, const Vec3& default_value)
{
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
	, node_attribute(nullptr)
{
	auto& e = (Element&)_element;
	if (e.first_property && e.first_property.next)
	{
		e.first_property.next.value.toString(name);
	}
	else
	{
		name[0] = '\0';
	}
}




template <typename T> static OptionalError<T> read(Cursor* cursor)
{
	if (cursor.current + sizeof(T) > cursor.end) return Error("Reading past the end");
	T value = *(const T*)cursor.current;
	cursor.current += sizeof(T);
	return value;
}


static OptionalError<DataView> readShortString(Cursor* cursor)
{
	DataView value;
	OptionalError<uint8> length = read<uint8>(cursor);
	if (length.isError()) return Error();

	if (cursor.current + length.getValue() > cursor.end) return Error("Reading past the end");
	value.begin = cursor.current;
	cursor.current += length.getValue();

	value.end = cursor.current;

	return value;
}


static OptionalError<DataView> readLongString(Cursor* cursor)
{
	DataView value;
	OptionalError<uint32> length = read<uint32>(cursor);
	if (length.isError()) return Error();

	if (cursor.current + length.getValue() > cursor.end) return Error("Reading past the end");
	value.begin = cursor.current;
	cursor.current += length.getValue();

	value.end = cursor.current;

	return value;
}


static OptionalError<Property*> readProperty(Cursor* cursor)
{
	if (cursor.current == cursor.end) return Error("Reading past the end");

	std::unique_ptr<Property> prop = std::make_unique<Property>();
	prop.next = nullptr;
	prop.type = *cursor.current;
	++cursor.current;
	prop.value.begin = cursor.current;

	switch (prop.type)
	{
		case 'S':
		{
			OptionalError<DataView> val = readLongString(cursor);
			if (val.isError()) return Error();
			prop.value = val.getValue();
			break;
		}
		case 'Y': cursor.current += 2; break;
		case 'C': cursor.current += 1; break;
		case 'I': cursor.current += 4; break;
		case 'F': cursor.current += 4; break;
		case 'D': cursor.current += 8; break;
		case 'L': cursor.current += 8; break;
		case 'R':
		{
			OptionalError<uint32> len = read<uint32>(cursor);
			if (len.isError()) return Error();
			if (cursor.current + len.getValue() > cursor.end) return Error("Reading past the end");
			cursor.current += len.getValue();
			break;
		}
		case 'b':
		case 'f':
		case 'd':
		case 'l':
		case 'i':
		{
			OptionalError<uint32> length = read<uint32>(cursor);
			OptionalError<uint32> encoding = read<uint32>(cursor);
			OptionalError<uint32> comp_len = read<uint32>(cursor);
			if (length.isError() | encoding.isError() | comp_len.isError()) return Error();
			if (cursor.current + comp_len.getValue() > cursor.end) return Error("Reading past the end");
			cursor.current += comp_len.getValue();
			break;
		}
		default: return Error("Unknown property type");
	}
	prop.value.end = cursor.current;
	return prop.release();
}


static void deleteElement(Element* el)
{
	if (!el) return;

	delete el.first_property;
	deleteElement(el.child);
	Element* iter = el;
	// do not use recursion to avoid stack overflow
	do
	{
		Element* next = iter.sibling;
		delete iter;
		iter = next;
	} while (iter);
}


static OptionalError<uint64> readElementOffset(Cursor* cursor, uint16 version)
{
	if (version >= 7500)
	{
		OptionalError<uint64> tmp = read<uint64>(cursor);
		if (tmp.isError()) return Error();
		return tmp.getValue();
	}

	OptionalError<uint32> tmp = read<uint32>(cursor);
	if (tmp.isError()) return Error();
	return tmp.getValue();
}


static OptionalError<Element*> readElement(Cursor* cursor, uint32 version)
{
	OptionalError<uint64> end_offset = readElementOffset(cursor, version);
	if (end_offset.isError()) return Error();
	if (end_offset.getValue() == 0) return nullptr;

	OptionalError<uint64> prop_count = readElementOffset(cursor, version);
	OptionalError<uint64> prop_length = readElementOffset(cursor, version);
	if (prop_count.isError() || prop_length.isError()) return Error();

	const char* sbeg = 0;
	const char* send = 0;
	OptionalError<DataView> id = readShortString(cursor);
	if (id.isError()) return Error();

	Element* element = new Element();
	element.first_property = nullptr;
	element.id = id.getValue();

	element.child = nullptr;
	element.sibling = nullptr;

	Property** prop_link = &element.first_property;
	for (uint32 i = 0; i < prop_count.getValue(); ++i)
	{
		OptionalError<Property*> prop = readProperty(cursor);
		if (prop.isError())
		{
			deleteElement(element);
			return Error();
		}

		*prop_link = prop.getValue();
		prop_link = &(*prop_link).next;
	}

	if (cursor.current - cursor.begin >= (ptrdiff_t)end_offset.getValue()) return element;

	int BLOCK_SENTINEL_LENGTH = version >= 7500 ? 25 : 13;

	Element** link = &element.child;
	while (cursor.current - cursor.begin < ((ptrdiff_t)end_offset.getValue() - BLOCK_SENTINEL_LENGTH))
	{
		OptionalError<Element*> child = readElement(cursor, version);
		if (child.isError())
		{
			deleteElement(element);
			return Error();
		}

		*link = child.getValue();
		link = &(*link).sibling;
	}

	if (cursor.current + BLOCK_SENTINEL_LENGTH > cursor.end)
	{
		deleteElement(element); 
		return Error("Reading past the end");
	}

	cursor.current += BLOCK_SENTINEL_LENGTH;
	return element;
}


static bool isEndLine(const Cursor& cursor)
{
	return *cursor.current == '\n';
}


static void skipInsignificantWhitespaces(Cursor* cursor)
{
	while (cursor.current < cursor.end && isspace(*cursor.current) && *cursor.current != '\n')
	{
		++cursor.current;
	}
}


static void skipLine(Cursor* cursor)
{
	while (cursor.current < cursor.end && !isEndLine(*cursor))
	{
		++cursor.current;
	}
	if (cursor.current < cursor.end) ++cursor.current;
	skipInsignificantWhitespaces(cursor);
}


static void skipWhitespaces(Cursor* cursor)
{
	while (cursor.current < cursor.end && isspace(*cursor.current))
	{
		++cursor.current;
	}
	while (cursor.current < cursor.end && *cursor.current == ';') skipLine(cursor);
}


static bool isTextTokenChar(char c)
{
	return isalnum(c) || c == '_';
}


static DataView readTextToken(Cursor* cursor)
{
	DataView ret;
	ret.begin = cursor.current;
	while (cursor.current < cursor.end && isTextTokenChar(*cursor.current))
	{
		++cursor.current;
	}
	ret.end = cursor.current;
	return ret;
}


static OptionalError<Property*> readTextProperty(Cursor* cursor)
{
	std::unique_ptr<Property> prop = std::make_unique<Property>();
	prop.value.is_binary = false;
	prop.next = nullptr;
	if (*cursor.current == '"')
	{
		prop.type = 'S';
		++cursor.current;
		prop.value.begin = cursor.current;
		while (cursor.current < cursor.end && *cursor.current != '"')
		{
			++cursor.current;
		}
		prop.value.end = cursor.current;
		if (cursor.current < cursor.end) ++cursor.current; // skip '"'
		return prop.release();
	}
	
	if (isdigit(*cursor.current) || *cursor.current == '-')
	{
		prop.type = 'L';
		prop.value.begin = cursor.current;
		if (*cursor.current == '-') ++cursor.current;
		while (cursor.current < cursor.end && isdigit(*cursor.current))
		{
			++cursor.current;
		}
		prop.value.end = cursor.current;

		if (cursor.current < cursor.end && *cursor.current == '.')
		{
			prop.type = 'D';
			++cursor.current;
			while (cursor.current < cursor.end && isdigit(*cursor.current))
			{
				++cursor.current;
			}
			if (cursor.current < cursor.end && (*cursor.current == 'e' || *cursor.current == 'E'))
			{
				// 10.5e-013
				++cursor.current;
				if (cursor.current < cursor.end && *cursor.current == '-') ++cursor.current;
				while (cursor.current < cursor.end && isdigit(*cursor.current)) ++cursor.current;
			}


			prop.value.end = cursor.current;
		}
		return prop.release();
	}
	
	if (*cursor.current == 'T' || *cursor.current == 'Y')
	{
		// WTF is this
		prop.type = *cursor.current;
		prop.value.begin = cursor.current;
		++cursor.current;
		prop.value.end = cursor.current;
		return prop.release();
	}

	if (*cursor.current == '*')
	{
		prop.type = 'l';
		++cursor.current;
		// Vertices: *10740 { a: 14.2760353088379,... }
		while (cursor.current < cursor.end && *cursor.current != ':')
		{
			++cursor.current;
		}
		if (cursor.current < cursor.end) ++cursor.current; // skip ':'
		skipInsignificantWhitespaces(cursor);
		prop.value.begin = cursor.current;
		prop.count = 0;
		bool is_any = false;
		while (cursor.current < cursor.end && *cursor.current != '}')
		{
			if (*cursor.current == ',')
			{
				if (is_any) ++prop.count;
				is_any = false;
			}
			else if (!isspace(*cursor.current) && *cursor.current != '\n') is_any = true;
			if (*cursor.current == '.') prop.type = 'd';
			++cursor.current;
		}
		if (is_any) ++prop.count;
		prop.value.end = cursor.current;
		if (cursor.current < cursor.end) ++cursor.current; // skip '}'
		return prop.release();
	}

	assert(false);
	return Error("TODO");
}


static OptionalError<Element*> readTextElement(Cursor* cursor)
{
	DataView id = readTextToken(cursor);
	if (cursor.current == cursor.end) return Error("Unexpected end of file");
	if(*cursor.current != ':') return Error("Unexpected end of file");
	++cursor.current;

	skipWhitespaces(cursor);
	if (cursor.current == cursor.end) return Error("Unexpected end of file");

	Element* element = new Element;
	element.id = id;

	Property** prop_link = &element.first_property;
	while (cursor.current < cursor.end && *cursor.current != '\n' && *cursor.current != '{')
	{
		OptionalError<Property*> prop = readTextProperty(cursor);
		if (prop.isError())
		{
			deleteElement(element);
			return Error();
		}
		if (cursor.current < cursor.end && *cursor.current == ',')
		{
			++cursor.current;
			skipWhitespaces(cursor);
		}
		skipInsignificantWhitespaces(cursor);

		*prop_link = prop.getValue();
		prop_link = &(*prop_link).next;
	}
	
	Element** link = &element.child;
	if (*cursor.current == '{')
	{
		++cursor.current;
		skipWhitespaces(cursor);
		while (cursor.current < cursor.end && *cursor.current != '}')
		{
			OptionalError<Element*> child = readTextElement(cursor);
			if (child.isError())
			{
				deleteElement(element);
				return Error();
			}
			skipWhitespaces(cursor);

			*link = child.getValue();
			link = &(*link).sibling;
		}
		if (cursor.current < cursor.end) ++cursor.current; // skip '}'
	}
	return element;
}


static OptionalError<Element*> tokenizeText(const uint8* data, size_t size)
{
	Cursor cursor;
	cursor.begin = data;
	cursor.current = data;
	cursor.end = data + size;

	Element* root = new Element();
	root.first_property = nullptr;
	root.id.begin = nullptr;
	root.id.end = nullptr;
	root.child = nullptr;
	root.sibling = nullptr;

	Element** element = &root.child;
	while (cursor.current < cursor.end)
	{
		if (*cursor.current == ';' || *cursor.current == '\r' || *cursor.current == '\n')
		{
			skipLine(&cursor);
		}
		else
		{
			OptionalError<Element*> child = readTextElement(&cursor);
			if (child.isError())
			{
				deleteElement(root);
				return Error();
			}
			*element = child.getValue();
			if (!*element) return root;
			element = &(*element).sibling;
		}
	}

	return root;
}


static OptionalError<Element*> tokenize(const uint8* data, size_t size)
{
	Cursor cursor;
	cursor.begin = data;
	cursor.current = data;
	cursor.end = data + size;

	const Header* header = (const Header*)cursor.current;
	cursor.current += sizeof(*header);

	Element* root = new Element();
	root.first_property = nullptr;
	root.id.begin = nullptr;
	root.id.end = nullptr;
	root.child = nullptr;
	root.sibling = nullptr;

	Element** element = &root.child;
	for (;;)
	{
		OptionalError<Element*> child = readElement(&cursor, header.version);
		if (child.isError())
		{
			deleteElement(root);
			return Error();
		}
		*element = child.getValue();
		if (!*element) return root;
		element = &(*element).sibling;
	}
}


static void parseTemplates(const Element& root)
{
	const Element* defs = findChild(root, "Definitions");
	if (!defs) return;

	std::unordered_map<std::string, Element*> templates;
	Element* def = defs.child;
	while (def)
	{
		if (def.id == "ObjectType")
		{
			Element* subdef = def.child;
			while (subdef)
			{
				if (subdef.id == "PropertyTemplate")
				{
					DataView prop1 = def.first_property.value;
					DataView prop2 = subdef.first_property.value;
					std::string key((const char*)prop1.begin, prop1.end - prop1.begin);
					key += std::string((const char*)prop1.begin, prop1.end - prop1.begin);
					templates[key] = subdef;
				}
				subdef = subdef.sibling;
			}
		}
		def = def.sibling;
	}
	// TODO
}



Material::Material(const Scene& _scene, const IElement& _element)
	: Object(_scene, _element)
{
}


struct MaterialImpl : Material
{
	MaterialImpl(const Scene& _scene, const IElement& _element)
		: Material(_scene, _element)
	{
		for (const Texture*& tex : textures) tex = nullptr;
	}

	Type getType() const override { return Type::MATERIAL; }


	const Texture* getTexture(Texture::TextureType type) const override { return textures[type]; }
	Color getDiffuseColor() const override { return diffuse_color; }

	const Texture* textures[Texture::TextureType::COUNT];
	Color diffuse_color;
};


struct LimbNodeImpl : Object
{
	LimbNodeImpl(const Scene& _scene, const IElement& _element)
		: Object(_scene, _element)
	{
		is_node = true;
	}
	Type getType() const override { return Type::LIMB_NODE; }
};


struct NullImpl : Object
{
	NullImpl(const Scene& _scene, const IElement& _element)
		: Object(_scene, _element)
	{
		is_node = true;
	}
	Type getType() const override { return Type::NULL_NODE; }
};


NodeAttribute::NodeAttribute(const Scene& _scene, const IElement& _element)
	: Object(_scene, _element)
{
}


struct NodeAttributeImpl : NodeAttribute
{
	NodeAttributeImpl(const Scene& _scene, const IElement& _element)
		: NodeAttribute(_scene, _element)
	{
	}
	Type getType() const override { return Type::NODE_ATTRIBUTE; }
	DataView getAttributeType() const override { return attribute_type; }


	DataView attribute_type;
};


Geometry::Geometry(const Scene& _scene, const IElement& _element)
	: Object(_scene, _element)
{
}


struct GeometryImpl : Geometry
{
	enum VertexDataMapping
	{
		BY_POLYGON_VERTEX,
		BY_POLYGON,
		BY_VERTEX
	};

	struct NewVertex
	{
		~NewVertex() { delete next; }

		int index = -1;
		NewVertex* next = nullptr;
	};

	std::vector<Vec3> vertices;
	std::vector<Vec3> normals;
	std::vector<Vec2> uvs[s_uvs_max];
	std::vector<Vec4> colors;
	std::vector<Vec3> tangents;
	std::vector<int> materials;

	const Skin* skin = nullptr;

	std::vector<int> to_old_vertices;
	std::vector<NewVertex> to_new_vertices;

	GeometryImpl(const Scene& _scene, const IElement& _element)
		: Geometry(_scene, _element)
	{
	}


	Type getType() const override { return Type::GEOMETRY; }
	int getVertexCount() const override { return (int)vertices.size(); }
	const Vec3* getVertices() const override { return &vertices[0]; }
	const Vec3* getNormals() const override { return normals.empty() ? nullptr : &normals[0]; }
	const Vec2* getUVs(int index = 0) const override { return index < 0 || index >= s_uvs_max || uvs[index].empty() ? nullptr : &uvs[index][0]; }
	const Vec4* getColors() const override { return colors.empty() ? nullptr : &colors[0]; }
	const Vec3* getTangents() const override { return tangents.empty() ? nullptr : &tangents[0]; }
	const Skin* getSkin() const override { return skin; }
	const int* getMaterials() const override { return materials.empty() ? nullptr : &materials[0]; }


	void triangulate(const std::vector<int>& old_indices, std::vector<int>* indices, std::vector<int>* to_old)
	{
		assert(indices);
		assert(to_old);

		auto getIdx = [&old_indices](int i) . int {
			int idx = old_indices[i];
			return idx < 0 ? -idx - 1 : idx;
		};

		int in_polygon_idx = 0;
		for (int i = 0; i < old_indices.size(); ++i)
		{
			int idx = getIdx(i);
			if (in_polygon_idx <= 2)
			{
				indices.push_back(idx);
				to_old.push_back(i);
			}
			else
			{
				indices.push_back(old_indices[i - in_polygon_idx]);
				to_old.push_back(i - in_polygon_idx);
				indices.push_back(old_indices[i - 1]);
				to_old.push_back(i - 1);
				indices.push_back(idx);
				to_old.push_back(i);
			}
			++in_polygon_idx;
			if (old_indices[i] < 0)
			{
				in_polygon_idx = 0;
			}
		}
	}
};







struct Root : Object
{
	Root(const Scene& _scene, const IElement& _element)
		: Object(_scene, _element)
	{
		copyString(name, "RootNode");
		is_node = true;
	}
	Type getType() const override { return Type::ROOT; }
};

struct OptionalError<Object*> parseTexture(const Scene& scene, const Element& element)
{
	TextureImpl* texture = new TextureImpl(scene, element);
	const Element* texture_filename = findChild(element, "FileName");
	if (texture_filename && texture_filename.first_property)
	{
		texture.filename = texture_filename.first_property.value;
	}
	const Element* texture_relative_filename = findChild(element, "RelativeFilename");
	if (texture_relative_filename && texture_relative_filename.first_property)
	{
		texture.relative_filename = texture_relative_filename.first_property.value;
	}
	return texture;
}


template <typename T> static OptionalError<Object*> parse(const Scene& scene, const Element& element)
{
	T* obj = new T(scene, element);
	return obj;
}




static OptionalError<Object*> parseNodeAttribute(const Scene& scene, const Element& element)
{
	NodeAttributeImpl* obj = new NodeAttributeImpl(scene, element);
	const Element* type_flags = findChild(element, "TypeFlags");
	if (type_flags && type_flags.first_property)
	{
		obj.attribute_type = type_flags.first_property.value;
	}
	return obj;
}


static OptionalError<Object*> parseLimbNode(const Scene& scene, const Element& element)
{
	if (!element.first_property
		|| !element.first_property.next
		|| !element.first_property.next.next
		|| element.first_property.next.next.value != "LimbNode")
	{
		return Error("Invalid limb node");
	}

	LimbNodeImpl* obj = new LimbNodeImpl(scene, element);
	return obj;
}


static OptionalError<Object*> parseMesh(const Scene& scene, const Element& element)
{
	if (!element.first_property
		|| !element.first_property.next
		|| !element.first_property.next.next
		|| element.first_property.next.next.value != "Mesh")
	{
		return Error("Invalid mesh");
	}

	return new MeshImpl(scene, element);
}


static OptionalError<Object*> parseMaterial(const Scene& scene, const Element& element)
{
	MaterialImpl* material = new MaterialImpl(scene, element);
	const Element* prop = findChild(element, "Properties70");
	material.diffuse_color = { 1, 1, 1 };
	if (prop) prop = prop.child;
	while (prop)
	{
		if (prop.id == "P" && prop.first_property)
		{
			if (prop.first_property.value == "DiffuseColor")
			{
				material.diffuse_color.r = (float)prop.getProperty(4).getValue().toDouble();
				material.diffuse_color.g = (float)prop.getProperty(5).getValue().toDouble();
				material.diffuse_color.b = (float)prop.getProperty(6).getValue().toDouble();
			}
		}
		prop = prop.sibling;
	}
	return material;
}


template<typename T> static bool parseTextArrayRaw(const Property& property, T* out, int max_size);

template <typename T> static bool parseArrayRaw(const Property& property, T* out, int max_size)
{
	if (property.value.is_binary)
	{
		assert(out);

		int elem_size = 1;
		switch (property.type)
		{
			case 'l': elem_size = 8; break;
			case 'd': elem_size = 8; break;
			case 'f': elem_size = 4; break;
			case 'i': elem_size = 4; break;
			default: return false;
		}

		const uint8* data = property.value.begin + sizeof(uint32) * 3;
		if (data > property.value.end) return false;

		uint32 count = property.getCount();
		uint32 enc = *(const uint32*)(property.value.begin + 4);
		uint32 len = *(const uint32*)(property.value.begin + 8);

		if (enc == 0)
		{
			if ((int)len > max_size) return false;
			if (data + len > property.value.end) return false;
			memcpy(out, data, len);
			return true;
		}
		else if (enc == 1)
		{
			if (int(elem_size * count) > max_size) return false;
			return decompress(data, len, (uint8*)out, elem_size * count);
		}

		return false;
	}

	return parseTextArrayRaw(property, out, max_size);
}


template <typename T> const char* fromString(const char* str, const char* end, T* val);
template <> const char* fromString<int>(const char* str, const char* end, int* val)
{
	*val = atoi(str);
	const char* iter = str;
	while (iter < end && *iter != ',') ++iter;
	if (iter < end) ++iter; // skip ','
	return (const char*)iter;
}


template <> const char* fromString<uint64>(const char* str, const char* end, uint64* val)
{
	*val = strtoull(str, nullptr, 10);
	const char* iter = str;
	while (iter < end && *iter != ',') ++iter;
	if (iter < end) ++iter; // skip ','
	return (const char*)iter;
}


template <> const char* fromString<int64>(const char* str, const char* end, int64* val)
{
	*val = atoll(str);
	const char* iter = str;
	while (iter < end && *iter != ',') ++iter;
	if (iter < end) ++iter; // skip ','
	return (const char*)iter;
}


template <> const char* fromString<double>(const char* str, const char* end, double* val)
{
	*val = atof(str);
	const char* iter = str;
	while (iter < end && *iter != ',') ++iter;
	if (iter < end) ++iter; // skip ','
	return (const char*)iter;
}


template <> const char* fromString<float>(const char* str, const char* end, float* val)
{
	*val = (float)atof(str);
	const char* iter = str;
	while (iter < end && *iter != ',') ++iter;
	if (iter < end) ++iter; // skip ','
	return (const char*)iter;
}


const char* fromString(const char* str, const char* end, double* val, int count)
{
	const char* iter = str;
	for (int i = 0; i < count; ++i)
	{
		*val = atof(iter);
		++val;
		while (iter < end && *iter != ',') ++iter;
		if (iter < end) ++iter; // skip ','

		if (iter == end) return iter;

	}
	return (const char*)iter;
}


template <> const char* fromString<Vec2>(const char* str, const char* end, Vec2* val)
{
	return fromString(str, end, &val.x, 2);
}


template <> const char* fromString<Vec3>(const char* str, const char* end, Vec3* val)
{
	return fromString(str, end, &val.x, 3);
}


template <> const char* fromString<Vec4>(const char* str, const char* end, Vec4* val)
{
	return fromString(str, end, &val.x, 4);
}


template <> const char* fromString<Matrix>(const char* str, const char* end, Matrix* val)
{
	return fromString(str, end, &val.m[0], 16);
}


template<typename T> static void parseTextArray(const Property& property, std::vector<T>* out)
{
	const uint8* iter = property.value.begin;
	for(int i = 0; i < property.count; ++i)
	{
		T val;
		iter = (const uint8*)fromString<T>((const char*)iter, (const char*)property.value.end, &val);
		out.push_back(val);
	}
}


template<typename T> static bool parseTextArrayRaw(const Property& property, T* out_raw, int max_size)
{
	const uint8* iter = property.value.begin;
	
	T* out = out_raw;
	while (iter < property.value.end)
	{
		iter = (const uint8*)fromString<T>((const char*)iter, (const char*)property.value.end, out);
		++out;
		if (out - out_raw == max_size / sizeof(T)) return true;
	}
	return out - out_raw == max_size / sizeof(T);
}


template <typename T> static bool parseBinaryArray(const Property& property, std::vector<T>* out)
{
	assert(out);
	if (property.value.is_binary)
	{
		uint32 count = property.getCount();
		int elem_size = 1;
		switch (property.type)
		{
			case 'd': elem_size = 8; break;
			case 'f': elem_size = 4; break;
			case 'i': elem_size = 4; break;
			default: return false;
		}
		int elem_count = sizeof(T) / elem_size;
		out.resize(count / elem_count);

		if (count == 0) return true;
		return parseArrayRaw(property, &(*out)[0], int(sizeof((*out)[0]) * out.size()));
	}
	else
	{
		parseTextArray(property, out);
		return true;
	}
}


template <typename T> static bool parseDoubleVecData(Property& property, std::vector<T>* out_vec)
{
	assert(out_vec);
	if (!property.value.is_binary)
	{
		parseTextArray(property, out_vec);
		return true;
	}

	if (property.type == 'd')
	{
		return parseBinaryArray(property, out_vec);
	}

	assert(property.type == 'f');
	assert(sizeof((*out_vec)[0].x) == sizeof(double));
	std::vector<float> tmp;
	if (!parseBinaryArray(property, &tmp)) return false;
	int elem_count = sizeof((*out_vec)[0]) / sizeof((*out_vec)[0].x);
	out_vec.resize(tmp.size() / elem_count);
	double* out = &(*out_vec)[0].x;
	for (int i = 0, c = (int)tmp.size(); i < c; ++i)
	{
		out[i] = tmp[i];
	}
	return true;
}


template <typename T>
static bool parseVertexData(const Element& element,
	const char* name,
	const char* index_name,
	std::vector<T>* out,
	std::vector<int>* out_indices,
	GeometryImpl::VertexDataMapping* mapping)
{
	assert(out);
	assert(mapping);
	const Element* data_element = findChild(element, name);
	if (!data_element || !data_element.first_property) 	return false;

	const Element* mapping_element = findChild(element, "MappingInformationType");
	const Element* reference_element = findChild(element, "ReferenceInformationType");

	if (mapping_element && mapping_element.first_property)
	{
		if (mapping_element.first_property.value == "ByPolygonVertex")
		{
			*mapping = GeometryImpl::BY_POLYGON_VERTEX;
		}
		else if (mapping_element.first_property.value == "ByPolygon")
		{
			*mapping = GeometryImpl::BY_POLYGON;
		}
		else if (mapping_element.first_property.value == "ByVertice" ||
					mapping_element.first_property.value == "ByVertex")
		{
			*mapping = GeometryImpl::BY_VERTEX;
		}
		else
		{
			return false;
		}
	}
	if (reference_element && reference_element.first_property)
	{
		if (reference_element.first_property.value == "IndexToDirect")
		{
			const Element* indices_element = findChild(element, index_name);
			if (indices_element && indices_element.first_property)
			{
				if (!parseBinaryArray(*indices_element.first_property, out_indices)) return false;
			}
		}
		else if (reference_element.first_property.value != "Direct")
		{
			return false;
		}
	}
	return parseDoubleVecData(*data_element.first_property, out);
}


template <typename T>
static void splat(std::vector<T>* out,
	GeometryImpl::VertexDataMapping mapping,
	const std::vector<T>& data,
	const std::vector<int>& indices,
	const std::vector<int>& original_indices)
{
	assert(out);
	assert(!data.empty());

	if (mapping == GeometryImpl::BY_POLYGON_VERTEX)
	{
		if (indices.empty())
		{
			out.resize(data.size());
			memcpy(&(*out)[0], &data[0], sizeof(data[0]) * data.size());
		}
		else
		{
			out.resize(indices.size());
			int data_size = (int)data.size();
			for (int i = 0, c = (int)indices.size(); i < c; ++i)
			{
				if(indices[i] < data_size) (*out)[i] = data[indices[i]];
				else (*out)[i] = T();
			}
		}
	}
	else if (mapping == GeometryImpl::BY_VERTEX)
	{
		//  v0  v1 ...
		// uv0 uv1 ...
		assert(indices.empty());

		out.resize(original_indices.size());

		int data_size = (int)data.size();
		for (int i = 0, c = (int)original_indices.size(); i < c; ++i)
		{
			int idx = original_indices[i];
			if (idx < 0) idx = -idx - 1;
			if(idx < data_size) (*out)[i] = data[idx];
			else (*out)[i] = T();
		}
	}
	else
	{
		assert(false);
	}
}


template <typename T> static void remap(std::vector<T>* out, const std::vector<int>& map)
{
	if (out.empty()) return;

	std::vector<T> old;
	old.swap(*out);
	int old_size = (int)old.size();
	for (int i = 0, c = (int)map.size(); i < c; ++i)
	{
		if(map[i] < old_size) out.push_back(old[map[i]]);
		else out.push_back(T());
	}
}


static OptionalError<Object*> parseAnimationCurve(const Scene& scene, const Element& element)
{
	std::unique_ptr<AnimationCurveImpl> curve = std::make_unique<AnimationCurveImpl>(scene, element);

	const Element* times = findChild(element, "KeyTime");
	const Element* values = findChild(element, "KeyValueFloat");

	if (times && times.first_property)
	{
		curve.times.resize(times.first_property.getCount());
		if (!times.first_property.getValues(&curve.times[0], (int)curve.times.size() * sizeof(curve.times[0])))
		{
			return Error("Invalid animation curve");
		}
	}

	if (values && values.first_property)
	{
		curve.values.resize(values.first_property.getCount());
		if (!values.first_property.getValues(&curve.values[0], (int)curve.values.size() * sizeof(curve.values[0])))
		{
			return Error("Invalid animation curve");
		}
	}

	if (curve.times.size() != curve.values.size()) return Error("Invalid animation curve");

	return curve.release();
}


static int getTriCountFromPoly(const std::vector<int>& indices, int* idx)
{
	int count = 1;
	while (indices[*idx + 1 + count] >= 0)
	{
		++count;
	}

	*idx = *idx + 2 + count;
	return count;
}


static void add(GeometryImpl::NewVertex& vtx, int index)
{
	if (vtx.index == -1)
	{
		vtx.index = index;
	}
	else if (vtx.next)
	{
		add(*vtx.next, index);
	}
	else
	{
		vtx.next = new GeometryImpl::NewVertex;
		vtx.next.index = index;
	}
}


static OptionalError<Object*> parseGeometry(const Scene& scene, const Element& element)
{
	assert(element.first_property);

	const Element* vertices_element = findChild(element, "Vertices");
	if (!vertices_element || !vertices_element.first_property) return Error("Vertices missing");

	const Element* polys_element = findChild(element, "PolygonVertexIndex");
	if (!polys_element || !polys_element.first_property) return Error("Indices missing");

	std::unique_ptr<GeometryImpl> geom = std::make_unique<GeometryImpl>(scene, element);

	std::vector<Vec3> vertices;
	if (!parseDoubleVecData(*vertices_element.first_property, &vertices)) return Error("Failed to parse vertices");
	std::vector<int> original_indices;
	if (!parseBinaryArray(*polys_element.first_property, &original_indices)) return Error("Failed to parse indices");

	std::vector<int> to_old_indices;
	geom.triangulate(original_indices, &geom.to_old_vertices, &to_old_indices);
	geom.vertices.resize(geom.to_old_vertices.size());

	for (int i = 0, c = (int)geom.to_old_vertices.size(); i < c; ++i)
	{
		geom.vertices[i] = vertices[geom.to_old_vertices[i]];
	}

	geom.to_new_vertices.resize(vertices.size()); // some vertices can be unused, so this isn't necessarily the same size as to_old_vertices.
	const int* to_old_vertices = geom.to_old_vertices.empty() ? nullptr : &geom.to_old_vertices[0];
	for (int i = 0, c = (int)geom.to_old_vertices.size(); i < c; ++i)
	{
		int old = to_old_vertices[i];
		add(geom.to_new_vertices[old], i);
	}

	const Element* layer_material_element = findChild(element, "LayerElementMaterial");
	if (layer_material_element)
	{
		const Element* mapping_element = findChild(*layer_material_element, "MappingInformationType");
		const Element* reference_element = findChild(*layer_material_element, "ReferenceInformationType");

		std::vector<int> tmp;

		if (!mapping_element || !reference_element) return Error("Invalid LayerElementMaterial");

		if (mapping_element.first_property.value == "ByPolygon" &&
			reference_element.first_property.value == "IndexToDirect")
		{
			geom.materials.reserve(geom.vertices.size() / 3);
			for (int& i : geom.materials) i = -1;

			const Element* indices_element = findChild(*layer_material_element, "Materials");
			if (!indices_element || !indices_element.first_property) return Error("Invalid LayerElementMaterial");

			if (!parseBinaryArray(*indices_element.first_property, &tmp)) return Error("Failed to parse material indices");

			int tmp_i = 0;
			for (int poly = 0, c = (int)tmp.size(); poly < c; ++poly)
			{
				int tri_count = getTriCountFromPoly(original_indices, &tmp_i);
				for (int i = 0; i < tri_count; ++i)
				{
					geom.materials.push_back(tmp[poly]);
				}
			}
		}
		else
		{
			if (mapping_element.first_property.value != "AllSame") return Error("Mapping not supported");
		}
	}

	const Element* layer_uv_element = findChild(element, "LayerElementUV");
    while (layer_uv_element)
    {
        const int uv_index = layer_uv_element.first_property ? layer_uv_element.first_property.getValue().toInt() : 0;
        if (uv_index >= 0 && uv_index < Geometry::s_uvs_max)
        {
            std::vector<Vec2>& uvs = geom.uvs[uv_index];

            std::vector<Vec2> tmp;
            std::vector<int> tmp_indices;
            GeometryImpl::VertexDataMapping mapping;
            if (!parseVertexData(*layer_uv_element, "UV", "UVIndex", &tmp, &tmp_indices, &mapping)) return Error("Invalid UVs");
            if (!tmp.empty())
            {
                uvs.resize(tmp_indices.empty() ? tmp.size() : tmp_indices.size());
                splat(&uvs, mapping, tmp, tmp_indices, original_indices);
                remap(&uvs, to_old_indices);
            }
        }

        do
        {
            layer_uv_element = layer_uv_element.sibling;
        } while (layer_uv_element && layer_uv_element.id != "LayerElementUV");
    }

	const Element* layer_tangent_element = findChild(element, "LayerElementTangents");
	if (layer_tangent_element)
	{
		std::vector<Vec3> tmp;
		std::vector<int> tmp_indices;
		GeometryImpl::VertexDataMapping mapping;
		if (findChild(*layer_tangent_element, "Tangents"))
		{
			if (!parseVertexData(*layer_tangent_element, "Tangents", "TangentsIndex", &tmp, &tmp_indices, &mapping)) return Error("Invalid tangets");
		}
		else
		{
			if (!parseVertexData(*layer_tangent_element, "Tangent", "TangentIndex", &tmp, &tmp_indices, &mapping))  return Error("Invalid tangets");
		}
		if (!tmp.empty())
		{
			splat(&geom.tangents, mapping, tmp, tmp_indices, original_indices);
			remap(&geom.tangents, to_old_indices);
		}
	}

	const Element* layer_color_element = findChild(element, "LayerElementColor");
	if (layer_color_element)
	{
		std::vector<Vec4> tmp;
		std::vector<int> tmp_indices;
		GeometryImpl::VertexDataMapping mapping;
		if (!parseVertexData(*layer_color_element, "Colors", "ColorIndex", &tmp, &tmp_indices, &mapping)) return Error("Invalid colors");
		if (!tmp.empty())
		{
			splat(&geom.colors, mapping, tmp, tmp_indices, original_indices);
			remap(&geom.colors, to_old_indices);
		}
	}

	const Element* layer_normal_element = findChild(element, "LayerElementNormal");
	if (layer_normal_element)
	{
		std::vector<Vec3> tmp;
		std::vector<int> tmp_indices;
		GeometryImpl::VertexDataMapping mapping;
		if (!parseVertexData(*layer_normal_element, "Normals", "NormalsIndex", &tmp, &tmp_indices, &mapping)) return Error("Invalid normals");
		if (!tmp.empty())
		{
			splat(&geom.normals, mapping, tmp, tmp_indices, original_indices);
			remap(&geom.normals, to_old_indices);
		}
	}

	return geom.release();
}


static bool isString(const Property* prop)
{
	if (!prop) return false;
	return prop.getType() == Property::STRING;
}


static bool isLong(const Property* prop)
{
	if (!prop) return false;
	return prop.getType() == Property::LONG;
}


static bool parseConnections(const Element& root, Scene* scene)
{
	assert(scene);

	const Element* connections = findChild(root, "Connections");
	if (!connections) return true;

	const Element* connection = connections.child;
	while (connection)
	{
		if (!isString(connection.first_property)
			|| !isLong(connection.first_property.next)
			|| !isLong(connection.first_property.next.next))
		{
			Error::s_message = "Invalid connection";
			return false;
		}

		Scene::Connection c;
		c.from = connection.first_property.next.value.touint64();
		c.to = connection.first_property.next.next.value.touint64();
		if (connection.first_property.value == "OO")
		{
			c.type = Scene::Connection::OBJECT_OBJECT;
		}
		else if (connection.first_property.value == "OP")
		{
			c.type = Scene::Connection::OBJECT_PROPERTY;
			if (!connection.first_property.next.next.next)
			{
				Error::s_message = "Invalid connection";
				return false;
			}
			c.property = connection.first_property.next.next.next.value;
		}
		else
		{
			assert(false);
			Error::s_message = "Not supported";
			return false;
		}
		scene.m_connections.push_back(c);

		connection = connection.sibling;
	}
	return true;
}


static bool parseTakes(Scene* scene)
{
	const Element* takes = findChild((const Element&)*scene.getRootElement(), "Takes");
	if (!takes) return true;

	const Element* object = takes.child;
	while (object)
	{
		if (object.id == "Take")
		{
			if (!isString(object.first_property))
			{
				Error::s_message = "Invalid name in take";
				return false;
			}

			TakeInfo take;
			take.name = object.first_property.value;
			const Element* filename = findChild(*object, "FileName");
			if (filename)
			{
				if (!isString(filename.first_property))
				{
					Error::s_message = "Invalid filename in take";
					return false;
				}
				take.filename = filename.first_property.value;
			}
			const Element* local_time = findChild(*object, "LocalTime");
			if (local_time)
			{
				if (!isLong(local_time.first_property) || !isLong(local_time.first_property.next))
				{
					Error::s_message = "Invalid local time in take";
					return false;
				}

				take.local_time_from = fbxTimeToSeconds(local_time.first_property.value.toint64());
				take.local_time_to = fbxTimeToSeconds(local_time.first_property.next.value.toint64());
			}
			const Element* reference_time = findChild(*object, "ReferenceTime");
			if (reference_time)
			{
				if (!isLong(reference_time.first_property) || !isLong(reference_time.first_property.next))
				{
					Error::s_message = "Invalid reference time in take";
					return false;
				}

				take.reference_time_from = fbxTimeToSeconds(reference_time.first_property.value.toint64());
				take.reference_time_to = fbxTimeToSeconds(reference_time.first_property.next.value.toint64());
			}

			scene.m_take_infos.push_back(take);
		}

		object = object.sibling;
	}

	return true;
}

static void parseGlobalSettings(const Element& root, Scene* scene)
{
	for (ofbx::Element* settings = root.child; settings; settings = settings.sibling)
	{
		if (settings.id == "GlobalSettings")
		{
			for (ofbx::Element* props70 = settings.child; props70; props70 = props70.sibling)
			{
				if (props70.id == "Properties70")
				{
					for (ofbx::Element* node = props70.child; node; node = node.sibling)
					{
						if (!node.first_property)
							continue;

#define get_property(name, field, type) if(node.first_property.value == name) \
						{ \
							ofbx::IElementProperty* prop = node.getProperty(4); \
							if (prop) \
							{ \
								ofbx::DataView value = prop.getValue(); \
								scene.m_settings.field = *(type*)value.begin; \
							} \
						}

						get_property("UpAxis", UpAxis, UpVector);
						get_property("UpAxisSign", UpAxisSign, int);
						get_property("FrontAxis", FrontAxis, FrontVector);
						get_property("FrontAxisSign", FrontAxisSign, int);
						get_property("CoordAxis", CoordAxis, CoordSystem);
						get_property("CoordAxisSign", CoordAxisSign, int);
						get_property("OriginalUpAxis", OriginalUpAxis, int);
						get_property("OriginalUpAxisSign", OriginalUpAxisSign, int);
						get_property("UnitScaleFactor", UnitScaleFactor, float);
						get_property("OriginalUnitScaleFactor", OriginalUnitScaleFactor, float);
						get_property("TimeSpanStart", TimeSpanStart, uint64);
						get_property("TimeSpanStop", TimeSpanStop, uint64);
						get_property("TimeMode", TimeMode, FrameRate);
						get_property("CustomFrameRate", CustomFrameRate, float);

#undef get_property

						scene.m_scene_frame_rate = getFramerateFromTimeMode(scene.m_settings.TimeMode, scene.m_settings.CustomFrameRate);
					}
					break;
				}
			}
			break;
		}
	}
}


static bool parseObjects(const Element& root, Scene* scene)
{
	const Element* objs = findChild(root, "Objects");
	if (!objs) return true;

	scene.m_root = new Root(*scene, root);
	scene.m_root.id = 0;
	scene.m_object_map[0] = {&root, scene.m_root};

	const Element* object = objs.child;
	while (object)
	{
		if (!isLong(object.first_property))
		{
			Error::s_message = "Invalid";
			return false;
		}

		uint64 id = object.first_property.value.touint64();
		scene.m_object_map[id] = {object, nullptr};
		object = object.sibling;
	}

	for (auto iter : scene.m_object_map)
	{
		OptionalError<Object*> obj = nullptr;

		if (iter.second.object == scene.m_root) continue;

		if (iter.second.element.id == "Geometry")
		{
			Property* last_prop = iter.second.element.first_property;
			while (last_prop.next) last_prop = last_prop.next;
			if (last_prop && last_prop.value == "Mesh")
			{
				obj = parseGeometry(*scene, *iter.second.element);
			}
		}
		else if (iter.second.element.id == "Material")
		{
			obj = parseMaterial(*scene, *iter.second.element);
		}
		else if (iter.second.element.id == "AnimationStack")
		{
			obj = parse<AnimationStackImpl>(*scene, *iter.second.element);
			if (!obj.isError())
			{
				AnimationStackImpl* stack = (AnimationStackImpl*)obj.getValue();
				scene.m_animation_stacks.push_back(stack);
			}
		}
		else if (iter.second.element.id == "AnimationLayer")
		{
			obj = parse<AnimationLayerImpl>(*scene, *iter.second.element);
		}
		else if (iter.second.element.id == "AnimationCurve")
		{
			obj = parseAnimationCurve(*scene, *iter.second.element);
		}
		else if (iter.second.element.id == "AnimationCurveNode")
		{
			obj = parse<AnimationCurveNodeImpl>(*scene, *iter.second.element);
		}
		else if (iter.second.element.id == "Deformer")
		{
			IElementProperty* class_prop = iter.second.element.getProperty(2);

			if (class_prop)
			{
				if (class_prop.getValue() == "Cluster")
					obj = parseCluster(*scene, *iter.second.element);
				else if (class_prop.getValue() == "Skin")
					obj = parse<SkinImpl>(*scene, *iter.second.element);
			}
		}
		else if (iter.second.element.id == "NodeAttribute")
		{
			obj = parseNodeAttribute(*scene, *iter.second.element);
		}
		else if (iter.second.element.id == "Model")
		{
			IElementProperty* class_prop = iter.second.element.getProperty(2);

			if (class_prop)
			{
				if (class_prop.getValue() == "Mesh")
				{
					obj = parseMesh(*scene, *iter.second.element);
					if (!obj.isError())
					{
						Mesh* mesh = (Mesh*)obj.getValue();
						scene.m_meshes.push_back(mesh);
						obj = mesh;
					}
				}
				else if (class_prop.getValue() == "LimbNode")
					obj = parseLimbNode(*scene, *iter.second.element);
				else if (class_prop.getValue() == "Null")
					obj = parse<NullImpl>(*scene, *iter.second.element);
				else if (class_prop.getValue() == "Root")
					obj = parse<NullImpl>(*scene, *iter.second.element);
			}
		}
		else if (iter.second.element.id == "Texture")
		{
			obj = parseTexture(*scene, *iter.second.element);
		}

		if (obj.isError()) return false;

		scene.m_object_map[iter.first].object = obj.getValue();
		if (obj.getValue())
		{
			scene.m_all_objects.push_back(obj.getValue());
			obj.getValue().id = iter.first;
		}
	}

	for (const Scene::Connection& con : scene.m_connections)
	{
		Object* parent = scene.m_object_map[con.to].object;
		Object* child = scene.m_object_map[con.from].object;
		if (!child) continue;
		if (!parent) continue;

		switch (child.getType())
		{
			case Object::Type::NODE_ATTRIBUTE:
				if (parent.node_attribute)
				{
					Error::s_message = "Invalid node attribute";
					return false;
				}
				parent.node_attribute = (NodeAttribute*)child;
				break;
			case Object::Type::ANIMATION_CURVE_NODE:
				if (parent.isNode())
				{
					AnimationCurveNodeImpl* node = (AnimationCurveNodeImpl*)child;
					node.bone = parent;
					node.bone_link_property = con.property;
				}
				break;
		}

		switch (parent.getType())
		{
			case Object::Type::MESH:
			{
				MeshImpl* mesh = (MeshImpl*)parent;
				switch (child.getType())
				{
					case Object::Type::GEOMETRY:
						if (mesh.geometry)
						{
							Error::s_message = "Invalid mesh";
							return false;
						}
						mesh.geometry = (Geometry*)child;
						break;
					case Object::Type::MATERIAL: mesh.materials.push_back((Material*)child); break;
				}
				break;
			}
			case Object::Type::SKIN:
			{
				SkinImpl* skin = (SkinImpl*)parent;
				if (child.getType() == Object::Type::CLUSTER)
				{
					ClusterImpl* cluster = (ClusterImpl*)child;
					skin.clusters.push_back(cluster);
					if (cluster.skin)
					{
						Error::s_message = "Invalid cluster";
						return false;
					}
					cluster.skin = skin;
				}
				break;
			}
			case Object::Type::MATERIAL:
			{
				MaterialImpl* mat = (MaterialImpl*)parent;
				if (child.getType() == Object::Type::TEXTURE)
				{
					Texture::TextureType type = Texture::COUNT;
					if (con.property == "NormalMap")
						type = Texture::NORMAL;
					else if (con.property == "DiffuseColor")
						type = Texture::DIFFUSE;
					if (type == Texture::COUNT) break;

					if (mat.textures[type])
					{
						break;// This may happen for some models (eg. 2 normal maps in use)
						Error::s_message = "Invalid material";
						return false;
					}

					mat.textures[type] = (Texture*)child;
				}
				break;
			}
			case Object::Type::GEOMETRY:
			{
				GeometryImpl* geom = (GeometryImpl*)parent;
				if (child.getType() == Object::Type::SKIN) geom.skin = (Skin*)child;
				break;
			}
			case Object::Type::CLUSTER:
			{
				ClusterImpl* cluster = (ClusterImpl*)parent;
				if (child.getType() == Object::Type::LIMB_NODE || child.getType() == Object::Type::MESH || child.getType() == Object::Type::NULL_NODE)
				{
					if (cluster.link)
					{
						Error::s_message = "Invalid cluster";
						return false;
					}

					cluster.link = child;
				}
				break;
			}
			case Object::Type::ANIMATION_LAYER:
			{
				if (child.getType() == Object::Type::ANIMATION_CURVE_NODE)
				{
					((AnimationLayerImpl*)parent).curve_nodes.push_back((AnimationCurveNodeImpl*)child);
				}
			}
			break;
			case Object::Type::ANIMATION_CURVE_NODE:
			{
				AnimationCurveNodeImpl* node = (AnimationCurveNodeImpl*)parent;
				if (child.getType() == Object::Type::ANIMATION_CURVE)
				{
					if (!node.curves[0].curve)
					{
						node.curves[0].connection = &con;
						node.curves[0].curve = (AnimationCurve*)child;
					}
					else if (!node.curves[1].curve)
					{
						node.curves[1].connection = &con;
						node.curves[1].curve = (AnimationCurve*)child;
					}
					else if (!node.curves[2].curve)
					{
						node.curves[2].connection = &con;
						node.curves[2].curve = (AnimationCurve*)child;
					}
					else
					{
						Error::s_message = "Invalid animation node";
						return false;
					}
				}
				break;
			}
		}
	}

	for (auto iter : scene.m_object_map)
	{
		Object* obj = iter.second.object;
		if (!obj) continue;
		if(obj.getType() == Object::Type::CLUSTER)
		{
			if (!((ClusterImpl*)iter.second.object).postprocess())
			{
				Error::s_message = "Failed to postprocess cluster";
				return false;
			}
		}
	}

	return true;
}

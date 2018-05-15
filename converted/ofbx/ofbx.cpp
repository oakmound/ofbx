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

template <typename T> static bool parseArrayRaw(const Property& property, T* out, int max_size);
template <typename T> static bool parseBinaryArray(const Property& property, std::vector<T>* out);

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

package ofbx

struct Header {
	magic [21]uint8
	reserved [2]uint8
	version uint32
}

template <typename T> static OptionalError<T> read(Cursor* cursor) {
	if (cursor.current + sizeof(T) > cursor.end) return Error("Reading past the end");
	T value = *(const T*)cursor.current;
	cursor.current += sizeof(T);
	return value;
}

static OptionalError<DataView> readShortString(Cursor* cursor) {
	DataView value;
	OptionalError<uint8> length = read<uint8>(cursor);
	if (length.isError()) return Error();

	if (cursor.current + length.getValue() > cursor.end) return Error("Reading past the end");
	value.begin = cursor.current;
	cursor.current += length.getValue();

	value.end = cursor.current;

	return value;
}

static OptionalError<DataView> readLongString(Cursor* cursor) {
	DataView value;
	OptionalError<uint32> length = read<uint32>(cursor);
	if (length.isError()) return Error();

	if (cursor.current + length.getValue() > cursor.end) return Error("Reading past the end");
	value.begin = cursor.current;
	cursor.current += length.getValue();

	value.end = cursor.current;

	return value;
}

static OptionalError<Property*> readProperty(Cursor* cursor) {
	if (cursor.current == cursor.end) return Error("Reading past the end");

	std::unique_ptr<Property> prop = std::make_unique<Property>();
	prop.next = nullptr;
	prop.type = *cursor.current;
	++cursor.current;
	prop.value.begin = cursor.current;

	switch (prop.type) {
		case 'S': {
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
		case 'R': {
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
		case 'i': {
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

static void deleteElement(Element* el) {
	if (!el) return;

	delete el.first_property;
	deleteElement(el.child);
	Element* iter = el;
	// do not use recursion to avoid stack overflow
	do {
		Element* next = iter.sibling;
		delete iter;
		iter = next;
	} while (iter);
}

static OptionalError<uint64> readElementOffset(Cursor* cursor, uint16 version) {
	if (version >= 7500) {
		OptionalError<uint64> tmp = read<uint64>(cursor);
		if (tmp.isError()) return Error();
		return tmp.getValue();
	}

	OptionalError<uint32> tmp = read<uint32>(cursor);
	if (tmp.isError()) return Error();
	return tmp.getValue();
}

static OptionalError<Element*> readElement(Cursor* cursor, uint32 version) {
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
	for (uint32 i = 0; i < prop_count.getValue(); ++i) {
		OptionalError<Property*> prop = readProperty(cursor);
		if (prop.isError()) {
			deleteElement(element);
			return Error();
		}

		*prop_link = prop.getValue();
		prop_link = &(*prop_link).next;
	}

	if (cursor.current - cursor.begin >= (ptrdiff_t)end_offset.getValue()) return element;

	int BLOCK_SENTINEL_LENGTH = version >= 7500 ? 25 : 13;

	Element** link = &element.child;
	while (cursor.current - cursor.begin < ((ptrdiff_t)end_offset.getValue() - BLOCK_SENTINEL_LENGTH)) {
		OptionalError<Element*> child = readElement(cursor, version);
		if (child.isError()) {
			deleteElement(element);
			return Error();
		}

		*link = child.getValue();
		link = &(*link).sibling;
	}

	if (cursor.current + BLOCK_SENTINEL_LENGTH > cursor.end) {
		deleteElement(element); 
		return Error("Reading past the end");
	}

	cursor.current += BLOCK_SENTINEL_LENGTH;
	return element;
}

static bool isEndLine(const Cursor& cursor) {
	return *cursor.current == '\n';
}

static void skipInsignificantWhitespaces(Cursor* cursor) {
	while (cursor.current < cursor.end && isspace(*cursor.current) && *cursor.current != '\n') {
		++cursor.current;
	}
}

static void skipLine(Cursor* cursor) {
	while (cursor.current < cursor.end && !isEndLine(*cursor)) {
		++cursor.current;
	}
	if (cursor.current < cursor.end) ++cursor.current;
	skipInsignificantWhitespaces(cursor);
}

static void skipWhitespaces(Cursor* cursor) {
	while (cursor.current < cursor.end && isspace(*cursor.current)) {
		++cursor.current;
	}
	while (cursor.current < cursor.end && *cursor.current == ';') skipLine(cursor);
}

static bool isTextTokenChar(char c) {
	return isalnum(c) || c == '_';
}

static DataView readTextToken(Cursor* cursor) {
	DataView ret;
	ret.begin = cursor.current;
	while (cursor.current < cursor.end && isTextTokenChar(*cursor.current)) {
		++cursor.current;
	}
	ret.end = cursor.current;
	return ret;
}

static OptionalError<Property*> readTextProperty(Cursor* cursor) {
	std::unique_ptr<Property> prop = std::make_unique<Property>();
	prop.value.is_binary = false;
	prop.next = nullptr;
	if (*cursor.current == '"') {
		prop.type = 'S';
		++cursor.current;
		prop.value.begin = cursor.current;
		while (cursor.current < cursor.end && *cursor.current != '"') {
			++cursor.current;
		}
		prop.value.end = cursor.current;
		if (cursor.current < cursor.end) ++cursor.current; // skip '"'
		return prop.release();
	}
	
	if (isdigit(*cursor.current) || *cursor.current == '-') {
		prop.type = 'L';
		prop.value.begin = cursor.current;
		if (*cursor.current == '-') ++cursor.current;
		while (cursor.current < cursor.end && isdigit(*cursor.current)) {
			++cursor.current;
		}
		prop.value.end = cursor.current;

		if (cursor.current < cursor.end && *cursor.current == '.') {
			prop.type = 'D';
			++cursor.current;
			while (cursor.current < cursor.end && isdigit(*cursor.current)) {
				++cursor.current;
			}
			if (cursor.current < cursor.end && (*cursor.current == 'e' || *cursor.current == 'E')) {
				// 10.5e-013
				++cursor.current;
				if (cursor.current < cursor.end && *cursor.current == '-') ++cursor.current;
				while (cursor.current < cursor.end && isdigit(*cursor.current)) ++cursor.current;
			}

			prop.value.end = cursor.current;
		}
		return prop.release();
	}
	
	if (*cursor.current == 'T' || *cursor.current == 'Y') {
		// WTF is this
		prop.type = *cursor.current;
		prop.value.begin = cursor.current;
		++cursor.current;
		prop.value.end = cursor.current;
		return prop.release();
	}

	if (*cursor.current == '*') {
		prop.type = 'l';
		++cursor.current;
		// Vertices: *10740 { a: 14.2760353088379,... }
		while (cursor.current < cursor.end && *cursor.current != ':') {
			++cursor.current;
		}
		if (cursor.current < cursor.end) ++cursor.current; // skip ':'
		skipInsignificantWhitespaces(cursor);
		prop.value.begin = cursor.current;
		prop.count = 0;
		bool is_any = false;
		while (cursor.current < cursor.end && *cursor.current != '}') {
			if (*cursor.current == ',') {
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

static OptionalError<Element*> readTextElement(Cursor* cursor) {
	DataView id = readTextToken(cursor);
	if (cursor.current == cursor.end) return Error("Unexpected end of file");
	if(*cursor.current != ':') return Error("Unexpected end of file");
	++cursor.current;

	skipWhitespaces(cursor);
	if (cursor.current == cursor.end) return Error("Unexpected end of file");

	Element* element = new Element;
	element.id = id;

	Property** prop_link = &element.first_property;
	while (cursor.current < cursor.end && *cursor.current != '\n' && *cursor.current != '{') {
		OptionalError<Property*> prop = readTextProperty(cursor);
		if (prop.isError()) {
			deleteElement(element);
			return Error();
		}
		if (cursor.current < cursor.end && *cursor.current == ',') {
			++cursor.current;
			skipWhitespaces(cursor);
		}
		skipInsignificantWhitespaces(cursor);

		*prop_link = prop.getValue();
		prop_link = &(*prop_link).next;
	}
	
	Element** link = &element.child;
	if (*cursor.current == '{') {
		++cursor.current;
		skipWhitespaces(cursor);
		while (cursor.current < cursor.end && *cursor.current != '}') {
			OptionalError<Element*> child = readTextElement(cursor);
			if (child.isError()) {
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

static OptionalError<Element*> tokenizeText(const uint8* data, size_t size) {
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
	while (cursor.current < cursor.end) {
		if (*cursor.current == ';' || *cursor.current == '\r' || *cursor.current == '\n') {
			skipLine(&cursor);
		}
		else {
			OptionalError<Element*> child = readTextElement(&cursor);
			if (child.isError()) {
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

static OptionalError<Element*> tokenize(const uint8* data, size_t size) {
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
	for (;;) {
		OptionalError<Element*> child = readElement(&cursor, header.version);
		if (child.isError()) {
			deleteElement(root);
			return Error();
		}
		*element = child.getValue();
		if (!*element) return root;
		element = &(*element).sibling;
	}
}
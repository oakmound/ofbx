package ofbx

import (
	"encoding/binary"
)

type Header struct {
	magic [21]int
	reserved [2]int
	version int
}

// type Cursor struct{
// 	current *int
// 	begin *int 
// 	end *int
// }


type Cursor io.Reader

const(

	UINT8_BYTES = 4
	UINT32_BYTES = 32

)


func (c *Cursor ) readShortString() []byte{
	
	c.cur += UINT8_BYTES
	if c.cur > len(data){
		errors.NewError("Reading past the end")
	}
	return c.data[c.cur-UINT8_BYTES: c.cur]
}
func (c *Cursor ) readLongString() []byte{
	
	c.cur += UINT32_BYTES
	if c.cur > len(data){
		errors.NewError("Reading past the end")
	}
	return c.data[c.cur-UINT32_BYTES: c.cur]
}


func deleteElement(e *Element){
	return
}

func (c *Cursor) isEndLine() bool {
	return c.data[c.cur] == '\n'
}
//TODO: Make a isspace for bytes not unicode
func (c *Cursor) skipInsignificantWhitespaces(){
	for (c.cur < len(c.data) && unicode.IsSpace(c.data[c.cur]) && c.isEndLine()){
		c.cur++
	}
}

func (c *Cursor) skipLine(){
	for c.cur < len(c.data) && !c.isEndLine{
		c.cur++
	}
	if (c.cur < len(c.data)){
		c.cur++
	}
	c.skipInsignificantWhitespaces()
}

func (c *Cursor) skipWhitespaces(){
	for c.cur < len(c.data) && unicode.IsSpace(c.data[c.cur]){
		c.cur++
	}
	for c.cur < len(c.data) && c.data[c.cur] == ';'{
		c.skipLine()
	}
}

func (c *Cursor) isTextToken(){
	isTextTokenChar(c.data[c.cur])
}

func isTextTokenChar(char c) {
	return unicode.IsDigit(c) || unicode.IsLetter(c) ||  c == '_'
}


func (c *Cursor) readTextToken(){
	start := c.cur
	for c.cur < len(c.data) && c.isTextToken(){
		c.cur++
	}
	return c.data[start: c.cur]
}


func (c *Cursor) readElementOffset(version int) int64{
	if version >= 7500{
			return 64
	}
	return 32
}

func (c *Cursor) readProperty() *Property{
	if c.cur > len(c.data){
		return errors.NewError("Reading Past End")
	}
		prop := Property{}



		prop.typ = c.Read(


		prop.typ = c.data[c.cur]
		c.cur++

		switch(prop.typ){
		case 'S':
			val = c.readLongString()
			if val != nil {return errors.NewError("")}
			prop.value = val.getValue()//TODO: get value from the thing
		

		case 'Y': 
		c.cur += 2
		case 'C': 
		c.cur  += 1
		case 'I': 
		c.cur  += 4
		case 'F': 
		c.cur  += 4
		case 'D': 
		c.cur  += 8
		case 'L': 
		c.cur  += 8
		case 'R':
		 length := UINT32_BYTES
	
		if c.cur  + length  > len(c.data){
			errors.NewError("Reading past the end")
		} 
		c.cur  += len.getValue();
		case 'b'||'f'||'d'||'l'||'i':
			length := UINT32_BYTES
			OptionalError<uint32> length = read<uint32>(cursor);
			OptionalError<uint32> encoding = read<uint32>(cursor);
			OptionalError<uint32> comp_len = read<uint32>(cursor);
			if (length.isError() | encoding.isError() | comp_len.isError()) return Error();
			if (cursor.current + comp_len.getValue() > cursor.end) return Error("Reading past the end");
			cursor.current += comp_len.getValue();
			break;
		default:
			errors.NewError("Did not know this property")
		}
	
}

static OptionalError<Property*> readProperty(Cursor* cursor) {
	if (cursor.current == cursor.end) return Error("Reading past the end");

	std::unique_ptr<Property> prop = std::make_unique<Property>();
	prop.next = nullptr;
	prop.type = *cursor.current;
	++cursor.current;
	prop.value.begin = cursor.current;

	switch (prop.type) {
		
		case 'R': {
			OptionalError<uint32> len = read<uint32>(cursor);
			if (len.isError()) return Error();
			if (cursor.current + len.getValue() > cursor.end) return Error("Reading past the end");
			cursor.current += len.getValue();
			break;
		}
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






func tokenize(data []byte) *Element, errors.Error{
	cursor := NewCursor(data )

	//Get header here for the current thing

	root := &Element{}
	element := &root.child
	for true {
		child, err  := readElement(&cursor, header.version)
		if err != nil {
			deleteElement(root)
			return err
		}
		element = child
		if element == nil{
			return root
		}
		element = element.sibling
	}
}



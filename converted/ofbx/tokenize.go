package ofbx

import (
	"encoding/binary"
	"bufio"
)

type Header struct {
	magic [21]int
	reserved [2]int
	version int
}

type Cursor *bufio.Reader

const(
	UINT8_BYTES = 1
	UINT32_BYTES = 4
	HEADER_BYTES = (21 + 2 + 1) * 4
)

func (c *Cursor) readShortString() (string, error) {
	length, err := c.ReadByte()
	if err != nil {
		return nil, err
	}
	byt := make([]byte, int(length))
	_, err = c.Read(byt)
	if err != nil {
		return nil, err
	} 
	return string(byt), nil
}
func (c *Cursor ) readLongString() (string, error) {
	var length uint32
	err := binary.Read(c, binary.BigEndian, &length)
	if err != nil {
		return "", err
	}
	byt := make([]byte, length)
	_, err = c.Read(byt)
	if err != nil {
		return "", err
	} 
	return string(byt), nil
}

func (c *Cursor) isEndLine() bool {
	by, err := c.Peek(1)
	if err != nil {
		fmt.Println(err)
	}
	return by[0] == '\n'
}
//TODO: Make a isspace for bytes not unicode
func (c *Cursor) skipInsignificantWhitespaces() error {
	for {
		by, _, err := c.ReadRune()
		if err != nil {
			return err
		}
		if unicode.IsSpace(by[0]) && by[0] != '\n' {
			continue
		}
		c.UnreadRune()
		break
	}
}

func (c *Cursor) skipLine(){
	c.ReadLine()
}

func (c *Cursor) skipWhitespaces() error {
	for {
		by, _, err := c.ReadRune()
		if err != nil {
			return err
		}
		if unicode.IsSpace(by[0]) {
			continue
		}
		c.UnreadRune()
		break
	}
	for {
		by, _, err := c.ReadRune()
		if err != nil {
			return err
		}
		if by[0] == ';' {
			c.skipLine()
			continue
		}
		c.UnreadRune()
		break
	}
}

func isTextTokenChar(c rune) {
	return unicode.IsDigit(c) || unicode.IsLetter(c) ||  c == '_'
}


func (c *Cursor) readTextToken() (DataView, error) {
	out := bytes.NewBuffer([]byte{})
	for {
		r, _, err := c.ReadRune()
		if err != nil {
			return nil, err
		}
		if isTextTokenChar(r) {
			out.WriteRune(r)
			continue
		}
		c.UnreadRune()
	}
	return DataView(out), nil
}


func (c *Cursor) readElementOffset(version uint16) (uint64, error) {
	if version >= 7500{
		var i uint64 
		err := binary.Read(c, binary.BigEndian, &i)
		return i, err
	}
	var i uint32 
	err := binary.Read(c, binary.BigEndian, &i)
	return uint64(i), err
}

func (c *Curesor) readBytes(len int) []byte{
	tempArr := make([]byte, len)
	_, err := c.Read(tempArr)
	return tempArr
}

func (c *Cursor) readProperty() *Property{
	if c.cur > len(c.data){
		return errors.NewError("Reading Past End")
	}
		prop := Property{}
		typ, _, err := c.ReadRune()
		if err !=nil{
			fmt.Println(err)
		}
		prop.typ = typ
		switch(prop.typ){
		case 'S':
			prop.value, err = c.readLongString()
		case 'Y': 
			prop.value = c.readBytes(2)
		case 'C': 
		prop.value =c.readBytes(1)
		case 'I': 
		prop.value =c.readBytes(4)
		case 'F': 
		prop.value =c.readBytes(4)
		case 'D': 
		prop.value =c.readBytes(8)
		case 'L': 
		prop.value =c.readBytes(8)
		case 'R':
		 	tmp := c.readBytes(4)
			length :=  binary.BigEndian.Uint32(tmp)
			prop.value = append(tmp,c.readBytes(length)...)
		case 'b'||'f'||'d'||'l'||'i':
			temp := c.readBytes(8)
			tempArr := c.readBytes(4)
			length :=  binary.BigEndian.Uint32(tempArr)
			prop.value = append(append(temp, tempArr...),c.readBytes(length)...)
		default:
			errors.NewError("Did not know this property")
		}
		if err != nil{
			fmt.Println(err)
		}
		return &prop
}

func (c *Cursor) readElement(version int){
	initial
	end_offset := c.readElementOffset(version)
	if end_offset == 0{
		return nil
	}
	prop_count := c.readElementOffset(version)
	prop_length := c.readElementOffset(version)
	id := c.readShortString()

	element := Element{}
	element.id = id

	oldProp := *element.first_property
	
	for i := 0 ; i < prop_count; i++{
		prop := c.readProperty()
		oldProp.next = prop
		oldProp = prop
	}

	if initial_size -c.Buffered() >- end_offset{
		return element
	}
	BLOCK_SENTINEL_LENGTH = 13
	if version >= 7500{
		BLOCK_SENTINEL_LENGTH = 25
	}

	link := &element.child

	for initial_size- c.buffered() < end_offset -BLOCK_SENTINEL_LENGTH {
		child := c.readElement(version)
		if child==nil{
			fmt.Println(errors.NewError("ReadingChild element failed"))
		}
		*link = child
		link = &(*link).sibling
	}
	c.Discard(BLOCK_SENTINEL_LENGTH)
	return element
}

func (c *Cursor) readTextProperty() (*Property, error) {
	prop = &Property{}

	r, _, err := cursor.ReadRune()
	if err != nil {
		return nil, err
	}
	if r == '"' {
		prop.type = 'S'
		val := []byte{}
		for {
			
		}
		while (cursor.current < cursor.end && *cursor.current != '"') {
			++cursor.current;
		}
		prop.value.end = cursor.current;
		if (cursor.current < cursor.end) ++cursor.current; // skip '"'
		return prop
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

	return nil, errors.New("TODO")
}

func (c *Cursor) readTextElement() (*Element, error) {
	id, err := cursor.readTextToken()
	if err != nil {
		return nil, err
	}
	r, _, err := cursor.readRune()
	if err != nil {
		return nil, err
	}
	if r != ':' {
		return nil, errors.New("unexpected end of file")
	}

	if err = cursor.skipWhitespaces(); err != nil {
		return nil, err
	}

	element := &Element{}
	element.id = id

	prop_link := &element.first_property
	for {
		by, err := cursor.Peek(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if by[0] == '\n' || by[0] == '{' {
			break
		}
		prop, err = readTextProperty(cursor)
		if err != nil {
			return nil, err
		}
		by, err := cursor.Peek(1)
		if err != io.EOF {
			if err != nil {
				return nil, err
			}
			if by[0] == ',' {
				cursor.Discard(1)
				cursor.skipWhitespaces()
			}
		}
		cursor.skipInsignificantWhitespaces()

		*prop_link = prop.getValue()
		prop_link = prop_link.next
	}
	
	link := &element.child
	r, _, err := cursor.ReadRune()
	if err != nil {
		return nil, err
	} 
	if r == '{' {
		cursor.skipWhitespaces()
		for {
			by, err := cursor.Peek()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}
			if by[0] == "}" {
				cursor.Discard(1)
				break
			}
			child, err := readTextElement(cursor)
			if err != nil  {
				return nil, err
			}
			cursor.skipWhitespaces()

			*link = child.getValue()
			link = link.sibling
		}
	} else {
		cursor.UnreadRune()
	}
	return element
}

func tokenizeText(data []byte) (*Element, error) {
	cursor := Cursor(bufio.NewReader(data))
	root := &Element{}
	element := &root.child
	for _, err := cursor.Peek(1); err != io.EOF {
		v, _, err := cursor.ReadRune()
		if err != nil {
			return nil, err
		}
		if (v == ';' || v == '\r' || v == '\n') {
			skipLine(cursor)
		} else {
			child, err := cursor.readTextElement()
			if err != nil {
				return nil, err
			}
			*element = child.getValue()
			if element == nil {
				return root, nil
			}
			element = element.sibling
		}
	}
	return root, nil
}

func tokenize(data []byte) (*Element, error) {
	cursor := Cursor(bufio.NewReader(data))

	var header Header
	err := binary.Read(cursor, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}
	
	root := &Element{}
	element := &root.child
	for true {
		child, err := readElement(&cursor, header.version)
		if err != nil {
			return nil, err
		}
		*element = child
		if element == nil{
			return root, nil
		}
		element = element.sibling
	}
}
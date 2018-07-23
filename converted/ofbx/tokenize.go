package ofbx

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"strconv"
	"unicode"

	"github.com/pkg/errors"
)

// Header is a magic set of ints
type Header struct {
	Magic    [21]uint8
	Reserved [2]uint8
	Version  uint32
}

func (h Header) String() string {
	b := []byte{}
	for _, n := range h.Magic {
		b = append(b, byte(n))
	}
	b = append(b, ' ')
	for _, n := range h.Reserved {
		b = append(b, byte(n))
	}
	s := string(b)
	s += " " + strconv.Itoa(int(h.Version))
	return s
}

// Cursor is a rapper for a reader
type Cursor struct {
	*bufio.Reader
	cr *CountReader
}

// ReadSoFar returns how much of the data has been read
func (c *Cursor) ReadSoFar() int {
	return c.cr.ReadSoFar - c.Reader.Buffered()
}

func (c *Cursor) readShortString() (string, error) {
	length, err := c.ReadByte()
	if err != nil {
		return "", err
	}
	//fmt.Print("slength is ", length)
	byt := make([]byte, int(length))
	_, err = io.ReadFull(c, byt)
	if err != nil {
		return "", err
	}
	return string(byt), nil
}
func (c *Cursor) readLongString() (string, error) {
	var length uint32
	err := binary.Read(c, binary.LittleEndian, &length)
	if err != nil {
		return "", err
	}
	byt := make([]byte, length)
	_, err = io.ReadFull(c, byt)
	if err != nil {
		return "", err
	}
	//fmt.Print(" |S Prop had ", length, "|")
	return string(byt), nil
}

func (c *Cursor) isEndLine() bool {
	by, err := c.Peek(1)
	if err != nil {
		//fmt.Println(err)
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
		if unicode.IsSpace(by) && by != '\n' {
			continue
		}
		c.UnreadRune()
		break
	}
	return nil
}

func (c *Cursor) skipLine() {
	c.ReadBytes('\n')
	c.skipInsignificantWhitespaces()
}

func (c *Cursor) skipWhitespaces() error {
	for {
		by, _, err := c.ReadRune()
		if err != nil {
			return err
		}
		if unicode.IsSpace(by) {
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
		if by == ';' {
			c.skipLine()
			continue
		}
		c.UnreadRune()
		break
	}
	return nil
}

func isTextTokenChar(c rune) bool {
	return unicode.IsDigit(c) || unicode.IsLetter(c) || c == '_'
}

func (c *Cursor) readTextToken() (*DataView, error) {
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
		break
	}
	return BufferDataView(out), nil
}

func (c *Cursor) readElementOffset(version uint16) (uint64, error) {
	if version >= 7500 {
		var i uint64
		err := binary.Read(c, binary.LittleEndian, &i)
		return i, err
	}
	var i uint32
	err := binary.Read(c, binary.LittleEndian, &i)
	return uint64(i), err
}

func (c *Cursor) readBytes(length int) []byte {
	tempArr := make([]byte, length)
	_, err := io.ReadFull(c, tempArr)
	if err != nil {
		//fmt.Println(err)
	}
	return tempArr
}

func (c *Cursor) readProperty() (*Property, error) {
	if _, err := c.Peek(1); err != nil {
		return nil, errors.New("Reading Past End")
	}
	prop := Property{}
	typ, _, err := c.ReadRune()
	if err != nil {
		//fmt.Println(err)
	}
	prop.Type = PropertyType(typ)
	var val string
	//fmt.Println("Got property type:", string(prop.typ))
	switch prop.Type {
	case 'S':
		val, err = c.readLongString()
	case 'Y':
		val = string(c.readBytes(2))
	case 'C':
		val = string(c.readBytes(1))
	case 'I':
		val = string(c.readBytes(4))
	case 'F':
		val = string(c.readBytes(4))
	case 'D':
		val = string(c.readBytes(8))
	case 'L':
		val = string(c.readBytes(8))
	case 'R':
		tmp := c.readBytes(4)
		length := int(binary.LittleEndian.Uint32(tmp))
		val = string(append(tmp, c.readBytes(length)...))
	case 'b', 'f', 'd', 'l', 'i':
		unCompressedLength := c.readBytes(4)
		encoding := c.readBytes(4)
		compressedLength := c.readBytes(4)
		length := int(binary.LittleEndian.Uint32(compressedLength))
		if int(binary.LittleEndian.Uint32(encoding)) == 0 {
			elemCount := int(binary.LittleEndian.Uint32(unCompressedLength))
			switch prop.Type {
			case 'f', 'i':
				length = elemCount * 4
			case 'd', 'l':
				length = elemCount * 8
			}
		}
		prop.Encoding = binary.LittleEndian.Uint32(encoding)
		prop.compressedLength = binary.LittleEndian.Uint32(compressedLength)
		prop.Count = int(binary.LittleEndian.Uint32(unCompressedLength))
		//fmt.Println("prop lengths", unCompressedLength, compressedLength, "props encoding", encoding)
		val = string(c.readBytes(length))
	default:
		return nil, errors.New("Did not know this property:" + string(prop.Type))
	}
	if err != nil {
		//fmt.Println(err)
	}

	prop.value = NewDataView(val)

	return &prop, nil
}

func (c *Cursor) readElement(version uint16) (*Element, error) {
	v, _ := c.Peek(12)
	footer := true
	for _, b := range v {
		if b != 0 {
			footer = false
			break
		}
	}
	if footer {
		// Note we don't actually read the footer contents yet,
		// as far as we know the footer holds no useful information
		//fmt.Println("Returning footer")
		return nil, nil
	}

	end_offset, err := c.readElementOffset(version)
	if err != nil {
		return nil, err
	}
	//fmt.Println("Obtained element end offset", end_offset)
	prop_count, err := c.readElementOffset(version)
	if err != nil {
		return nil, err
	}
	//fmt.Println("Obtained element prop count", prop_count)
	_, err = c.readElementOffset(version)
	if err != nil {
		return nil, err
	}
	//fmt.Println("Obtained property list length", prop_list_length)
	id, err := c.readShortString()
	if err != nil {
		return nil, err
	}
	//fmt.Println("Read short string", id)

	element := Element{}
	element.ID = NewDataView(id)

	element.Properties = make([]*Property, prop_count)
	for i := uint64(0); i < prop_count; i++ {
		element.Properties[i], err = c.readProperty()
	}

	if uint64(c.ReadSoFar()) >= end_offset {
		//fmt.Println("NO Sentinel sizes ", c.ReadSoFar(), end_offset)
		return &element, nil
	}
	blockSentinelLength := 13
	if version >= 7500 {
		blockSentinelLength = 25
	}

	//fmt.Print("sizes pre children ", c.ReadSoFar(), end_offset, uint64(blockSentinelLength))
	for uint64(c.ReadSoFar()) < end_offset-uint64(blockSentinelLength) {
		child, err := c.readElement(version)
		if err != nil {
			return nil, errors.Wrap(err, "ReadingChild element failed")
		}
		element.Children = append(element.Children, child)
		if uint64(c.ReadSoFar()) > end_offset {
			//fmt.Println("Read past where we were supposed to!!", c.ReadSoFar(), end_offset)
		}
	}
	if uint64(c.ReadSoFar()) > end_offset {
		//fmt.Println("Read past where we were supposed to!!", c.ReadSoFar(), end_offset)
	}
	c.Discard(blockSentinelLength)
	//fmt.Println("With Sentinel", uint64(c.ReadSoFar()), "versus", end_offset)
	return &element, nil
}

func (c *Cursor) readTextProperty() (*Property, error) {
	//fmt.Println("Reading text property")
	prop := &Property{}

	r, _, err := c.ReadRune()
	if err != nil {
		return nil, err
	}
	if r == '"' {
		//fmt.Println("Quote start")
		prop.Type = 'S'
		val := bytes.NewBuffer([]byte{})
		for {
			r, _, err := c.ReadRune()
			if err != nil {
				if err == io.EOF {
					break //?
				}
				return nil, err
			}
			if r == '"' {
				break
			}
			val.WriteRune(r)
		}
		prop.value = BufferDataView(val)
		//fmt.Println("Quote end", prop.value.String())
		return prop, nil
	}

	if unicode.IsDigit(r) || r == '-' {
		//fmt.Println("Digit start")
		prop.Type = 'L'
		if r != '-' {
			c.UnreadRune()
		}
		val := bytes.NewBuffer([]byte{})
		for {
			r, _, err := c.ReadRune()
			if err != nil {
				if err == io.EOF {
					break //?
				}
				return nil, err
			}
			if !unicode.IsDigit(r) {
				break
			}
			val.WriteRune(r)
		}

		r, _, err = c.ReadRune()

		if err == nil && r == '.' {
			prop.Type = 'D'
			val.WriteRune(r)
			for {
				r, _, err := c.ReadRune()
				if err != nil {
					if err == io.EOF {
						break //?
					}
					return nil, err
				}
				if !unicode.IsDigit(r) {
					break
				}
				val.WriteRune(r)
			}
			r, _, err = c.ReadRune()
			if err == nil && r == 'e' || r == 'E' {
				// 10.5e-013
				val.WriteRune(r)
				r, _, err = c.ReadRune()
				if r != '-' || !unicode.IsDigit(r) {
					return nil, errors.New("malformed floating point with exponent")
				}
				val.WriteRune(r)
				for {
					r, _, err := c.ReadRune()
					if err != nil {
						if err == io.EOF {
							//fmt.Println("EOF?")
							break //?
						}
						return nil, err
					}
					if !unicode.IsDigit(r) {
						break
					}
					val.WriteRune(r)
				}
			}
		}
		prop.value = BufferDataView(val)
		//fmt.Println("Digits end", prop.value.String())
		return prop, nil
	}

	if r == 'T' || r == 'Y' {
		// WTF is this
		//fmt.Println("WTF start")
		prop.Type = PropertyType(r)
		b, err := c.ReadByte()
		prop.value = NewDataView(string(b))
		//fmt.Println("WTF end", b)
		return prop, err
	}
	if r == '*' {
		//fmt.Println("Asterisk start")
		prop.Type = 'l'
		// Vertices: *10740 { a: 14.2760353088379,... } //Pulled from original...
		pBytes := bytes.NewBuffer([]byte{})
		r2, _, _ := c.ReadRune()
		pBytes.WriteRune(r2)
		_, err := c.Peek(1)
		for err == nil && r2 != ':' {
			r2, _, _ = c.ReadRune()
			pBytes.WriteRune(r2)
			_, err = c.Peek(1)
		}

		c.skipInsignificantWhitespaces() //We assume it is insignificant, so we don't add to buffer

		prop.Count = 0

		is_any := false
		_, err = c.Peek(1)
		for err == nil && r2 != '}' {
			if r2 == ',' {
				if is_any {
					prop.Count++
				}
				is_any = false
			} else if !unicode.IsSpace(r2) && r2 != '\n' {
				is_any = true
			}
			if r2 == '.' {
				prop.Type = 'd'
			}
			r2, _, _ = c.ReadRune()
			pBytes.WriteRune(r2)
			_, err = c.Peek(1)
		}
		if is_any {
			prop.Count++
		}
		prop.value = BufferDataView(pBytes)
		//fmt.Println("Asterisk end", prop.value.String())
		return prop, err
	}
	//fmt.Println("r was", string(r))
	return nil, errors.New("TODO")
}

func (c *Cursor) readTextElement() (*Element, error) {
	//fmt.Println("Read text token start")
	id, err := c.readTextToken()
	if err != nil {
		return nil, err
	}
	//fmt.Println("Read rune start")
	r, _, err := c.ReadRune()
	if err != nil {
		return nil, err
	}
	//fmt.Println("Read rune complete")
	if r != ':' {
		return nil, errors.New("Unexpected end of file")
	}
	//fmt.Println("Skip whitespaces start")
	if err = c.skipWhitespaces(); err != nil {
		return nil, err
	}

	element := &Element{}
	element.ID = id

	//fmt.Println("Looping over properties")
	for {
		//fmt.Println("Property loop")
		by, err := c.Peek(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if by[0] == '\n' || by[0] == '{' {
			break
		}
		prop, err := c.readTextProperty()
		if err != nil {
			return nil, err
		}
		by, err = c.Peek(1)
		if err != io.EOF {
			if err != nil {
				return nil, err
			}
			if by[0] == ',' {
				c.Discard(1)
				c.skipWhitespaces()
			}
		}
		c.skipInsignificantWhitespaces()

		element.Properties = append(element.Properties, prop)
	}

	r, _, err = c.ReadRune()
	if err != nil {
		return nil, err
	}
	if r == '{' {
		c.skipWhitespaces()
		for {
			by, err := c.Peek(1)
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}
			if by[0] == '}' {
				c.Discard(1)
				break
			}
			child, err := c.readTextElement()
			if err != nil {
				return nil, err
			}
			c.skipWhitespaces()

			element.Children = append(element.Children, child)
		}
	} else {
		c.UnreadRune()
	}
	return element, nil
}

//Todo: make this matter
func tokenizeText(r io.Reader) (*Element, error) {
	return nil, nil
	/*
		cr := NewCountReader(r)
		r2 := bufio.NewReader(cr)
		cursor := Cursor{r2, cr}
		root := &Element{}
		element := &root.child
		_, err := cursor.Peek(1)
		fmt.Println("Looping tokenizeText")
		for ; err != io.EOF; _, err = cursor.Peek(1) {
			v, _, err := cursor.ReadRune()
			if err != nil {
				return nil, err
			}
			fmt.Println("Read rune from tokenizeText", v)
			if v == ';' || v == '\r' || v == '\n' {
				fmt.Println("Skipping line")
				cursor.skipLine()
			} else {
				fmt.Println("Reading text element")
				child, err := cursor.readTextElement()
				fmt.Println("Read text element")
				if err != nil {
					fmt.Println("Read text element error", err)
					return nil, err
				}
				*element = child
				if element == nil {
					return root, nil
				}
				element = &(*element).sibling
			}
		}
		return root, nil
	*/
}

func tokenize(r io.Reader) (*Element, error) {
	countReader := NewCountReader(r)
	r2 := bufio.NewReader(countReader)
	cursor := &Cursor{r2, countReader}
	//fmt.Println("initial stats: ", r2.Buffered(), cursor.ReadSoFar())

	var header Header
	err := binary.Read(cursor, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}

	//fmt.Println("Header:", header)

	root := &Element{}

	for {
		//fmt.Println("Reading element")
		child, err := cursor.readElement(uint16(header.Version))
		if err != nil {
			//fmt.Println("Read element failure", err)
			return nil, err
		}

		if child == nil {
			return root, nil
		}
		root.Children = append(root.Children, child)
	}
}

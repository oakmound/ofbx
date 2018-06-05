package ofbx

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unicode"

	"github.com/pkg/errors"
)

type Header struct {
	magic    [21]int
	reserved [2]int
	version  int
}

type Cursor struct {
	bufio.Reader
}

const (
	UINT8_BYTES  = 1
	UINT32_BYTES = 4
	HEADER_BYTES = (21 + 2 + 1) * 4
)

func (c *Cursor) readShortString() (string, error) {
	length, err := c.ReadByte()
	if err != nil {
		return "", err
	}
	byt := make([]byte, int(length))
	_, err = c.Read(byt)
	if err != nil {
		return "", err
	}
	return string(byt), nil
}
func (c *Cursor) readLongString() (string, error) {
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
		if unicode.IsSpace(by) && by != '\n' {
			continue
		}
		c.UnreadRune()
		break
	}
	return nil
}

func (c *Cursor) skipLine() {
	c.ReadLine()
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
	}
	return &DataView{*out}, nil
}

func (c *Cursor) readElementOffset(version uint16) (uint64, error) {
	if version >= 7500 {
		var i uint64
		err := binary.Read(c, binary.BigEndian, &i)
		return i, err
	}
	var i uint32
	err := binary.Read(c, binary.BigEndian, &i)
	return uint64(i), err
}

func (c *Cursor) readBytes(len int) []byte {
	tempArr := make([]byte, len)
	_, err := c.Read(tempArr)
	if err != nil {
		fmt.Println(err)
	}
	return tempArr
}

func (c *Cursor) readProperty() (*Property, error) {
	if c.Buffered() == 0 {
		return nil, errors.New("Reading Past End")
	}
	prop := Property{}
	typ, _, err := c.ReadRune()
	if err != nil {
		fmt.Println(err)
	}
	prop.typ = PropertyType(typ)
	var val string
	switch prop.typ {
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
		length := int(binary.BigEndian.Uint32(tmp))
		val = string(append(tmp, c.readBytes(length)...))
	case 'b', 'f', 'd', 'l', 'i':
		temp := c.readBytes(8)
		tempArr := c.readBytes(4)
		length := int(binary.BigEndian.Uint32(tempArr))
		val = string(append(append(temp, tempArr...), c.readBytes(length)...))
	default:
		return nil, errors.New("Did not know this property")
	}
	if err != nil {
		fmt.Println(err)
	}

	//convert to prop
	prop.value = NewDataView(val)

	return &prop, nil
}

func (c *Cursor) readElement(version uint16) (*Element, error) {
	initialSize := c.Buffered()
	end_offset, err := c.readElementOffset(version)
	if err != nil {
		return nil, err
	}
	prop_count, err := c.readElementOffset(version)
	if err != nil {
		return nil, err
	}
	_, err = c.readElementOffset(version)
	if err != nil {
		return nil, err
	}
	id, err := c.readShortString()
	if err != nil {
		return nil, err
	}

	element := Element{}
	element.id = NewDataView(id)

	oldProp := element.first_property

	for i := uint64(0); i < prop_count; i++ {
		prop, err := c.readProperty()
		if err != nil {
			return nil, err
		}
		oldProp.next = prop
		oldProp = prop
	}

	if uint64(initialSize-c.Buffered()) > -end_offset {
		return &element, nil
	}
	blockSentinelLength := 13
	if version >= 7500 {
		blockSentinelLength = 25
	}

	link := &element.child

	for uint64(initialSize-c.Buffered()) < end_offset-uint64(blockSentinelLength) {
		child, err := c.readElement(version)
		if err != nil {
			return nil, errors.Wrap(err, "ReadingChild element failed")
		}
		*link = child
		link = &(*link).sibling
	}
	c.Discard(blockSentinelLength)
	return &element, nil
}

func (c *Cursor) readTextProperty() (*Property, error) {
	prop := &Property{}

	r, _, err := c.ReadRune()
	if err != nil {
		return nil, err
	}
	if r == '"' {
		prop.typ = 'S'
		prop.value = NewDataView("")
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
			prop.value.WriteRune(r)
		}
		return prop, nil
	}

	if unicode.IsDigit(r) || r == '-' {
		prop.typ = 'L'
		if r != '-' {
			c.UnreadRune()
		}
		prop.value = NewDataView("")
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
			prop.value.WriteRune(r)
		}

		r, _, err = c.ReadRune()

		if err == nil && r == '.' {
			prop.typ = 'D'
			prop.value.WriteRune(r)
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
				prop.value.WriteRune(r)
			}
			r, _, err = c.ReadRune()
			if err == nil && r == 'e' || r == 'E' {
				// 10.5e-013
				prop.value.WriteRune(r)
				r, _, err = c.ReadRune()
				if r != '-' || !unicode.IsDigit(r) {
					return nil, errors.New("malformed floating point with exponent")
				}
				prop.value.WriteRune(r)
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
					prop.value.WriteRune(r)
				}
			}
		}
		return prop, nil
	}

	if r == 'T' || r == 'Y' {
		// WTF is this
		prop.typ = PropertyType(r)
		b, err := c.ReadByte()
		prop.value = NewDataView(string(b))
		return prop, err
	}
	if r == '*' {
		prop.typ = 'l'
		// Vertices: *10740 { a: 14.2760353088379,... } //Pulled from original...
		pBytes := NewDataView("")
		r2, _, _ := c.ReadRune()
		pBytes.WriteRune(r2)
		for c.Buffered() > 0 && r2 != ':' {
			r2, _, _ = c.ReadRune()
			pBytes.WriteRune(r2)
		}

		c.skipInsignificantWhitespaces() //We assume it is insignificat so dont add to buff

		prop.count = 0

		is_any := false
		for c.Buffered() > 0 && r2 != '}' {
			if r2 == ',' {
				if is_any {
					prop.count++
				}
				is_any = false
			} else if !unicode.IsSpace(r2) && r2 != '\n' {
				is_any = true
			}
			if r2 == '.' {
				prop.typ = 'd'
			}
			r2, _, _ = c.ReadRune()
			pBytes.WriteRune(r2)
		}
		if is_any {
			prop.count++
		}
		prop.value = pBytes
		return prop, err
	}
	return nil, errors.New("TODO")
}

func (c *Cursor) readTextElement() (*Element, error) {
	id, err := c.readTextToken()
	if err != nil {
		return nil, err
	}
	r, _, err := c.ReadRune()
	if err != nil {
		return nil, err
	}
	if r != ':' {
		return nil, errors.New("unexpected end of file")
	}

	if err = c.skipWhitespaces(); err != nil {
		return nil, err
	}

	element := &Element{}
	element.id = id

	prop_link := &element.first_property
	for {
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

		*prop_link = prop
		prop_link = &(*prop_link).next
	}

	link := &element.child
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

			*link = child
			link = &(*link).sibling
		}
	} else {
		c.UnreadRune()
	}
	return element, nil
}

func tokenizeText(data []byte) (*Element, error) {
	r := bufio.NewReader(bytes.NewReader(data))
	cursor := Cursor{*r}
	root := &Element{}
	element := &root.child
	_, err := cursor.Peek(1)
	for ; err != io.EOF; _, err = cursor.Peek(1) {
		v, _, err := cursor.ReadRune()
		if err != nil {
			return nil, err
		}
		if v == ';' || v == '\r' || v == '\n' {
			cursor.skipLine()
		} else {
			child, err := cursor.readTextElement()
			if err != nil {
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
}

func tokenize(data []byte) (*Element, error) {
	r := bufio.NewReader(bytes.NewReader(data))
	cursor := &Cursor{*r}

	var header Header
	err := binary.Read(cursor, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}

	root := &Element{}
	element := &root.child
	for true {
		child, err := cursor.readElement(uint16(header.version))
		if err != nil {
			return nil, err
		}
		*element = child
		if element == nil {
			return root, nil
		}
		element = &(*element).sibling
	}
	return *element, nil
}

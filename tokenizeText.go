package ofbx

import (
	"bytes"
	"errors"
	"io"
	"unicode"
)

func (c *Cursor) isEndLine() bool {
	by, err := c.Peek(1)
	if err != nil {
		//fmt.Println(err)
	}
	return by[0] == '\n'
}

// TODO: Make a isspace for bytes not unicode
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

		isAny := false
		_, err = c.Peek(1)
		for err == nil && r2 != '}' {
			if r2 == ',' {
				if isAny {
					prop.Count++
				}
				isAny = false
			} else if !unicode.IsSpace(r2) && r2 != '\n' {
				isAny = true
			}
			if r2 == '.' {
				prop.Type = 'd'
			}
			r2, _, _ = c.ReadRune()
			pBytes.WriteRune(r2)
			_, err = c.Peek(1)
		}
		if isAny {
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

// Todo: make this matter
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

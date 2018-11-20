package ofbx

import (
	"bufio"
	"encoding/binary"
	"io"
	"strconv"

	"github.com/pkg/errors"
)

// Header is a magic set of ints
type Header struct {
	Reserved [2]uint8
	Version  uint32
}

func (h Header) String() string {
	b := []byte{}
	for _, n := range h.Reserved {
		b = append(b, byte(n))
	}
	s := string(b)
	s += " " + strconv.Itoa(int(h.Version))
	return s
}

// Cursor is a wrapper for a reader
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

	endOffset, err := c.readElementOffset(version)
	if err != nil {
		return nil, err
	}
	//fmt.Println("Obtained element end offset", endOffset)
	propCt, err := c.readElementOffset(version)
	if err != nil {
		return nil, err
	}
	//fmt.Println("Obtained element prop count", propCt)
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

	//fmt.Println("Read A:", c.ReadSoFar(), endOffset)

	element := Element{}
	element.ID = NewDataView(id)

	element.Properties = make([]*Property, propCt)
	for i := uint64(0); i < propCt; i++ {
		element.Properties[i], err = c.readProperty()
	}

	//fmt.Println("Read B:", c.ReadSoFar(), endOffset)

	if uint64(c.ReadSoFar()) >= endOffset {
		//fmt.Println("NO Sentinel; sizes: ", c.ReadSoFar(), endOffset)
		return &element, nil
	}
	blockSentinelLength := 13
	if version >= 7500 {
		blockSentinelLength = 25
	}

	//fmt.Print("sizes pre children ", c.ReadSoFar(), endOffset, uint64(blockSentinelLength))
	for uint64(c.ReadSoFar()) < endOffset-uint64(blockSentinelLength) {
		child, err := c.readElement(version)
		if err != nil {
			return nil, errors.Wrap(err, "ReadingChild element failed")
		}
		element.Children = append(element.Children, child)
		if uint64(c.ReadSoFar()) > endOffset {
			//fmt.Println("Read past where we were supposed to!!", c.ReadSoFar(), endOffset)
		}
	}
	if uint64(c.ReadSoFar()) > endOffset {
		//fmt.Println("Read past where we were supposed to!!", c.ReadSoFar(), endOffset)
	}
	//fmt.Println("About to discard", c.ReadSoFar(), blockSentinelLength)
	c.Discard(blockSentinelLength)
	//fmt.Println("With Sentinel", uint64(c.ReadSoFar()), "versus", endOffset)
	//fmt.Println("Read C:", c.ReadSoFar(), endOffset)
	return &element, nil
}

func tokenize(r io.Reader) (*Element, error) {
	countReader := NewCountReader(r)
	r2 := bufio.NewReader(countReader)
	cursor := &Cursor{r2, countReader}
	//fmt.Println("initial stats: ", r2.Buffered(), cursor.ReadSoFar())

	ok := IsBinary(cursor)
	if !ok {
		return nil, errors.New("Non-binary FBX")
	}

	var header Header
	err := binary.Read(cursor, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}
	//fmt.Println(header)

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

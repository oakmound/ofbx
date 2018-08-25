package threefbx

import (
	"bufio"
	"encoding/binary"
	"io"
)

// CountReader an io reader that knows how many bytes it has currently read
type CountReader struct {
	io.Reader
	ReadSoFar int
}

// Read reads in some number of bytes and returns the result from the underlying io.Reader
func (c *CountReader) Read(p []byte) (n int, err error) {
	n, err = c.Reader.Read(p)
	c.ReadSoFar += n
	return n, err
}

// NewCountReader creates a new wrapper around an io.Reader with a count of 0
func NewCountReader(r io.Reader) *CountReader {
	return &CountReader{
		r, 0,
	}
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

func (c *Cursor) readBytes(length int) []byte {
	tempArr := make([]byte, length)
	_, err := io.ReadFull(c, tempArr)
	if err != nil {
		//fmt.Println(err)
	}
	return tempArr
}

type BinaryReader struct {
	r      CountReader
	Offset int
	order  binary.ByteOrder
}

func NewBinaryReader(r io.Reader, littleEndian bool) *BinaryReader {
	countReader := NewCountReader(r)
	reader := &Cursor{bufio.NewReader(countReader), countReader}
	br := &BinaryReader{}
	br.r = reader
	br.littleEndian = littleEndian
	return br
}

// seems like true/false representation depends on exporter.
// true: 1 or 'Y'(=0x59), false: 0 or 'T'(=0x54)
// then sees LSB.
func (br *BinaryReader) getBoolean() bool {
	return br.getUint8()&1 != 0
}
func (br *BinaryReader) getBooleanArray(size int) []bool {
	bs := make([]bool, size)
	for i := 0; i < size; i++ {
		bs[i] = this.getBoolean()
	}
	return bs
}
func (br *BinaryReader) getUint8() {
	return br.ReadByte()
}
func (br *BinaryReader) getInt16() {
	var i int16
	binary.Read(br, br.order, &i)
	return i
}
func (br *BinaryReader) getInt32() {
	var i int32
	binary.Read(br, br.order, &i)
	return i
}
func (br *BinaryReader) getUint32() {
	var i uint32
	binary.Read(br, br.order, &i)
	return i
}
func (br *BinaryReader) getInt64() {
	var i int64
	binary.Read(br, br.order, &i)
	return i
}
func (br *BinaryReader) getUint64() {
	var i uint64
	binary.Read(br, br.order, &i)
	return i
}
func (br *BinaryReader) getFloat32() {
	var i float32
	binary.Read(br, br.order, &i)
	return i
}
func (br *BinaryReader) getFloat64() {
	var i float64
	binary.Read(br, br.order, &i)
	return i
}

func (br *BinaryReader) getInt32Array(size int) []int32 {
	is := make([]int32, size)
	for i := 0; i < size; i++ {
		is[i] = this.getInt32()
	}
	return is
}

func (br *BinaryReader) getInt64Array(size int) {
	is := make([]int64, size)
	for i := 0; i < size; i++ {
		is[i] = this.getInt64()
	}
	return is
}

func (br *BinaryReader) getFloat32Array(size int) {
	is := make([]float32, size)
	for i := 0; i < size; i++ {
		is[i] = this.getFloat32()
	}
	return is
}

func (br *BinaryReader) getFloat64Array(size int) {
	is := make([]float64, size)
	for i := 0; i < size; i++ {
		is[i] = this.getFloat64()
	}
	return is
}
func (br *BinaryReader) getArrayBuffer(size int) {
	bs := make([]byte, size)
	io.ReadFull(r, bs)
	return bs
}

func (br *BinaryReader) getString(size int) {
	byt := make([]byte, int(length))
	_, err = io.ReadFull(c, byt)
	if err != nil {
		return "", err
	}
	return string(byt), nil
}

func (br *BinaryReader) getShortString() string {
	length, _ := br.ReadByte()
	return br.getString(int(length))
}
func (br *BinaryReader) getLongString() string {
	return br.getString(int(br.getUint32()))
}

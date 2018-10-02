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
	r            *Cursor
	order        binary.ByteOrder
	littleEndian bool
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
func (br *BinaryReader) getBooleanArray(size uint32) []bool {
	bs := make([]bool, size)
	for i := uint32(0); i < size; i++ {
		bs[i] = br.getBoolean()
	}
	return bs
}
func (br *BinaryReader) getUint8() uint8 {
	b, _ := br.r.ReadByte()
	return b
}
func (br *BinaryReader) getInt16() int16 {
	var i int16
	binary.Read(br.r, br.order, &i)
	return i
}
func (br *BinaryReader) getInt32() int32 {
	var i int32
	binary.Read(br.r, br.order, &i)
	return i
}
func (br *BinaryReader) getUint32() uint32 {
	var i uint32
	binary.Read(br.r, br.order, &i)
	return i
}
func (br *BinaryReader) getInt64() int64 {
	var i int64
	binary.Read(br.r, br.order, &i)
	return i
}
func (br *BinaryReader) getUint64() uint64 {
	var i uint64
	binary.Read(br.r, br.order, &i)
	return i
}
func (br *BinaryReader) getFloat32() float32 {
	var i float32
	binary.Read(br.r, br.order, &i)
	return i
}
func (br *BinaryReader) getFloat64() float64 {
	var i float64
	binary.Read(br.r, br.order, &i)
	return i
}

func (br *BinaryReader) getInt32Array(size uint32) []int32 {
	is := make([]int32, size)
	for i := uint32(0); i < size; i++ {
		is[i] = br.getInt32()
	}
	return is
}

func (br *BinaryReader) getInt64Array(size uint32) []int64 {
	is := make([]int64, size)
	for i := uint32(0); i < size; i++ {
		is[i] = br.getInt64()
	}
	return is
}

func (br *BinaryReader) getFloat32Array(size uint32) []float32 {
	is := make([]float32, size)
	for i := uint32(0); i < size; i++ {
		is[i] = br.getFloat32()
	}
	return is
}

func (br *BinaryReader) getFloat64Array(size uint32) []float64 {
	is := make([]float64, size)
	for i := uint32(0); i < size; i++ {
		is[i] = br.getFloat64()
	}
	return is
}
func (br *BinaryReader) getArrayBuffer(size uint32) []byte {
	bs := make([]byte, size)
	io.ReadFull(br.r, bs)
	return bs
}

func (br *BinaryReader) getString(size uint32) string {
	byt := make([]byte, size)
	_, err := io.ReadFull(br.r, byt)
	if err != nil {
		return ""
	}
	return string(byt)
}

func (br *BinaryReader) getShortString() string {
	length, _ := br.r.ReadByte()
	return br.getString(uint32(length))
}
func (br *BinaryReader) getLongString() string {
	return br.getString(br.getUint32())
}

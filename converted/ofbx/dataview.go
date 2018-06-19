package ofbx

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type DataView struct {
	bytes.Reader
}

func NewDataView(s string) *DataView {
	return &DataView{
		*bytes.NewReader([]byte(s)),
	}
}

func BufferDataView(buff *bytes.Buffer) *DataView {
	return &DataView{
		*bytes.NewReader(buff.Bytes()),
	}
}

func (dv *DataView) String() string {
	ln := dv.Len()
	data := make([]byte, ln)
	_, err := dv.Read(data)
	if err != nil {
		fmt.Println(err)
	}
	// Todo: maybe don't do this?
	dv.Seek(0, io.SeekStart)
	return string(data)
}

func (dv *DataView) touint64() uint64 {
	var i uint64
	err := binary.Read(dv, binary.LittleEndian, &i)
	if err != nil {
		fmt.Println("binary read failure:", err)
	}
	return i
}

func (dv *DataView) toint64() int64 {
	var i int64
	err := binary.Read(dv, binary.LittleEndian, &i)
	if err != nil {
		fmt.Println("binary read failure:", err)
	}
	return i
}

func (dv *DataView) toInt32() int32 {
	var i int32
	err := binary.Read(dv, binary.LittleEndian, &i)
	if err != nil {
		fmt.Println("binary read failure:", err)
	}
	return i
}

func (dv *DataView) touint32() uint32 {
	var i uint32
	err := binary.Read(dv, binary.LittleEndian, &i)
	if err != nil {
		fmt.Println("binary read failure:", err)
	}
	return i
}

func (dv *DataView) toDouble() float64 {
	var i float64
	err := binary.Read(dv, binary.LittleEndian, &i)
	if err != nil {
		fmt.Println("binary read failure:", err)
	}
	return i
}

func (dv *DataView) toFloat() float32 {
	var i float32
	err := binary.Read(dv, binary.LittleEndian, &i)
	if err != nil {
		fmt.Println("binary read failure:", err)
	}
	return i
}

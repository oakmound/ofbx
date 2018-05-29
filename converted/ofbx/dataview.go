package ofbx

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type DataView bytes.Buffer

func (dv *DataView) touint64() uint64 {
	var i uint64
	err := binary.Read(dv, binary.BigEndian, &i)
	if err != nil {
		fmt.Println("binary read failure:", err)
	}
	return i
}

func (dv *DataView) toint64() int64 {
	var i int64
	err := binary.Read(dv, binary.BigEndian, &i)
	if err != nil {
		fmt.Println("binary read failure:", err)
	}
	return i
}

func (dv *DataView) toInt() int {
	var i int
	err := binary.Read(dv, binary.BigEndian, &i)
	if err != nil {
		fmt.Println("binary read failure:", err)
	}
	return i
}

func (dv *DataView) touint32() uint32 {
	var i uint32
	err := binary.Read(dv, binary.BigEndian, &i)
	if err != nil {
		fmt.Println("binary read failure:", err)
	}
	return i
}

func (dv *DataView) toDouble() float64 {
	var i float64
	err := binary.Read(dv, binary.BigEndian, &i)
	if err != nil {
		fmt.Println("binary read failure:", err)
	}
	return i
}

func (dv *DataView) toFloat() float32 {
	var i float32
	err := binary.Read(dv, binary.BigEndian, &i)
	if err != nil {
		fmt.Println("binary read failure:", err)
	}
	return i
}

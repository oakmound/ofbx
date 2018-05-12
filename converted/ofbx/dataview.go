package ofbx

import (
	"encoding/binary"
)

type DataView *bytes.Buffer

func (dv DataView) touint64() uint64 {
	i, _ := binary.ReadUvarint(dv)
	return i
}


func (dv DataView) toint64() int64 {
	i, _ := binary.ReadVarint(dv)
	return i
}


func (dv DataView) toInt() int {
	i, _ := binary.ReadVarint(dv)
	return int(i)
}


func (dv DataView) touint32() uint32 {
	if (is_binary)
	{
		assert(end - begin == sizeof(uint32))
		return *(uint32*)begin
	}
	return (uint32)atoll(( string)begin)
}


func (dv DataView) toDouble() float64 {
	if (is_binary)
	{
		assert(end - begin == sizeof(double))
		return *(double*)begin
	}
	return atof(( string)begin)
}


func (dv DataView) toFloat() float32 {
	if (is_binary)
	{
		assert(end - begin == sizeof(float))
		return *(float*)begin
	}
	return (float)atof(( string)begin)
}
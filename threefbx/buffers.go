package threefbx

import "fmt"

type bufferDefinition struct {
	dataSize      int32
	indices       []int32
	mappingType   string
	referenceType string
}

type intBuffer struct {
	bufferDefinition
	buffer []int32
}

type floatBuffer struct {
	bufferDefinition
	buffer []float64
}

func getDataSlicePos(bfd bufferDefinition, polygonVertexIndex, polygonIndex, vertexIndex int32) (from, to int32) {
	var index int32
	switch bfd.mappingType {
	case "ByPolygonVertex":
		index = polygonVertexIndex
	case "ByPolygon":
		index = polygonIndex
	case "ByVertice":
		index = vertexIndex
	case "AllSame":
		index = bfd.indices[0]
	default:
		fmt.Println("THREE.FBXLoader: unknown attribute mapping type " + bfd.mappingType)
	}
	if bfd.referenceType == "IndexToDirect" {
		index = bfd.indices[index]
	}
	from = index * bfd.dataSize
	to = from + bfd.dataSize
	return
}

// extracts the data from the correct position in the FBX array based on indexing type
func (info intBuffer) getData(polygonVertexIndex int, polygonIndex, vertexIndex int32) []int32 {
	from, to := getDataSlicePos(info.bufferDefinition, int32(polygonVertexIndex), polygonIndex, vertexIndex)
	out := make([]int32, info.dataSize)
	copy(out, info.buffer[from:to])
	return out
}

// extracts the data from the correct position in the FBX array based on indexing type
func (info floatBuffer) getData(polygonVertexIndex int, polygonIndex, vertexIndex int32) []float64 {
	from, to := getDataSlicePos(info.bufferDefinition, int32(polygonVertexIndex), polygonIndex, vertexIndex)
	out := make([]float64, info.dataSize)
	copy(out, info.buffer[from:to])
	return out
}

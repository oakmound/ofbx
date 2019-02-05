package threefbx

import (
	"fmt"
	"sort"
)

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

// ordereNormals ordrees our normals correctly
func (info floatBuffer) orderByVertice(indicies []int32) []float64 {
	ordered := make([]float64, 0)

	switch info.mappingType { //https://help.autodesk.com/view/FBX/2016/ENU/?guid=__cpp_ref_class_fbx_layer_element_html

	case "ByPolygonVertex": // Each vertex has as many entries as polygons it is a part of
		indexMapping := map[int32][]float64{}
		fmt.Println("Len of index ", len(indicies))
		for i, j := range indicies {
			if j < 0 {
				j = j ^ -1
				indicies[i] = j
			}
			indexMapping[j] = info.getData(i, 0, 0)
		}

		sort.Slice(indicies, func(x, y int) bool { return indicies[x] < indicies[y] })

		for _, j := range indicies {
			if val, ok := indexMapping[j]; ok {
				ordered = append(ordered, val...)
				delete(indexMapping, j)
			}
		}

	case "ByPolygon": // Each polygon can only have one mapping

	case "ByVertice": // Each vertex has one entry
		for i := range indicies {
			ordered = append(ordered, info.getData(0, 0, int32(i))...)
		}

	case "AllSame": //One mapping for entire surface
		for range indicies {
			ordered = append(ordered, info.getData(0, 0, 0)...)
		}

	default:
		fmt.Println("THREE.FBXLoader: unknown attribute mapping type " + info.mappingType)
	}

	return ordered
}

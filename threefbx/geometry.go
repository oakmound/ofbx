package threefbx

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/oakmound/oak/alg"
	"github.com/oakmound/oak/alg/floatgeom"
)

//TODO consider how they use transformdata and how it creates transforms
//can we skip transform data and just jam things on as we get them?

func generateTransform(td TransformData) mgl64.Mat4 {
	order := ZYXOrder
	if td.eulerOrder != nil {
		order = *td.eulerOrder
	}

	translation := floatgeom.Point3{}
	if td.translation != nil {
		translation = *td.translation
	}

	if td.rotationOffset != nil {
		translation = translation.Add(*td.rotationOffset)
	}

	rotation := mgl64.Ident4()
	if td.rotation != nil {
		rot := td.rotation.Mul(alg.DegToRad)
		rotation = makeRotationFromEuler(rot, order)
	}

	if td.preRotation != nil {
		rot := td.preRotation.Mul(alg.DegToRad)
		mat := makeRotationFromEuler(rot, order)
		rotation = mat.Mul4(rotation)
	}
	if td.postRotation != nil {
		rot := td.postRotation.Mul(alg.DegToRad)
		mat := makeRotationFromEuler(rot, order)
		mat = mat.Inverse()
		rotation = rotation.Mul4(mat)
	}

	transform := mgl64.Ident4()

	if td.scale != nil {
		transform = transform.Mul(*td.scale)
	}
	transform = transform.WithPosition(translation)
	transform = transform.Mul4(rotation)
	return transform
}

type InfoObject struct {
	MappingType   string
	ReferenceType string
	DataSize      int
	Indices       []int
	Buffer        []byte
}

// extracts the data from the correct position in the FBX array based on indexing type
func getData(polygonVertexIndex, polygonIndex, vertexIndex int, info InfoObject) {
	var index int
	switch info.MappingType {
	case "ByPolygonVertex":
		index = polygonVertexIndex
	case "ByPolygon":
		index = polygonIndex
	case "ByVertice":
		index = vertexIndex
	case "AllSame":
		index = infoObject.indices[0]
	default:
		fmt.Println("THREE.FBXLoader: unknown attribute mapping type " + info.MappingType)
	}
	if info.ReferenceType == "IndexToDirect" {
		index = info.Indices[index]
	}
	from := index * info.DataSize
	to := from + info.DataSize
	out := make([]byte, info.DataSize)
	copy(out, info.Buffer[from:to])
	return out
}

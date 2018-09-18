package threefbx

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/oakmound/oak/alg"
	"github.com/oakmound/oak/alg/floatgeom"
)

type EulerOrder int

const (
	ZYXOrder       EulerOrder = iota // -> XYZ extrinsic
	YZXOrder       EulerOrder = iota // -> XZY extrinsic
	XZYOrder       EulerOrder = iota // -> YZX extrinsic
	ZXYOrder       EulerOrder = iota // -> YXZ extrinsic
	YXZOrder       EulerOrder = iota // -> ZXY extrinsic
	XYZOrder       EulerOrder = iota // -> ZYX extrinsic
	LastEulerOrder EulerOrder = iota
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

	rotation := floatgeom.IdentityMatrix4()
	if td.rotation != nil {
		rot := td.rotation.Scale(alg.DegToRad)
		rotation = makeRotationFromEuler(rot, order)
	}

	if td.preRotation != nil {
		rot := td.preRotation.Scale(alg.DegToRad)
		mat := makeRotationFromEuler(rot, order)
		rotation = rotation.PreMultiply(mat)
	}
	if td.postRotation != nil {
		rot := td.postRotation.Scale(alg.DegToRad)
		mat := makeRotationFromEuler(rot, order)
		mat = mat.Inverse()
		rotation = rotation.multiply(mat)
	}

	transform := floatgeom.IdentityMatrix4()

	if td.scale != nil {
		transform = transform.Scale(*td.scale)
	}
	transform = transform.WithPosition(translation)
	transform = transform.Multiply(rotation)
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

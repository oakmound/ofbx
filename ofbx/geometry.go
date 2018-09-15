package threefbx

import (
	"github.com/oakmound/oak/alg"
)

type EulerOrder int

var (
	ZYXOrder EulerOrder = iota // -> XYZ extrinsic
	YZXOrder EulerOrder = iota // -> XZY extrinsic
	XZYOrder EulerOrder = iota // -> YZX extrinsic
	ZXYOrder EulerOrder = iota // -> YXZ extrinsic
	YXZOrder EulerOrder = iota // -> ZXY extrinsic
	XYZOrder EulerOrder = iota // -> ZYX extrinsic
	LastEulerOrder EulerOrder = iota
)

//TODO consider how they use transformdata and how it creates transforms
//can we skip transform data and just jam things on as we get them?

var tempMat = new THREE.Matrix4();
	var tempEuler = new THREE.Euler();
	var tempVec = new THREE.Vector3();
	var translation = new THREE.Vector3();
	var rotation = new THREE.Matrix4();
	// generate transformation from FBX transform data
	// ref: https://help.autodesk.com/view/FBX/2017/ENU/?guid=__files_GUID_10CDD63C_79C1_4F2D_BB28_AD2BE65A02ED_htm
	// transformData = {
	//	 eulerOrder: int,
	//	 translation: [],
	//   rotationOffset: [],
	//	 preRotation
	//	 rotation
	//	 postRotation
	//   scale
	// }
	// all entries are optional

func generateTransform(td TransformData) floatgeom.Matrix4 {	
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
	MappingType string
	ReferenceType string
	DataSize int
	Indices []int
	Buffer []byte
}

var dataArray = [];
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
			index = infoObject.indices[ 0 ]
		default:
			fmt.Println("THREE.FBXLoader: unknown attribute mapping type " + info.MappingType )
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
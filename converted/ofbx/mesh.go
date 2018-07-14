package ofbx

import "github.com/oakmound/oak/alg/floatgeom"

type Mesh struct {
	Object
	Geometry  *Geometry
	scene     *Scene
	Materials []*Material
}

func NewMesh(scene *Scene, element *Element) *Mesh {
	m := &Mesh{}
	m.Object = *NewObject(scene, element)
	m.Object.is_node = true
	return m
}

func (m *Mesh) Type() Type {
	return MESH
}

func (m *Mesh) getGeometricMatrix() Matrix {
	translation := resolveVec3Property(m, "GeometricTranslation", floatgeom.Point3{0, 0, 0})
	rotation := resolveVec3Property(m, "GeometricRotation", floatgeom.Point3{0, 0, 0})
	scale := resolveVec3Property(m, "GeometricScaling", floatgeom.Point3{1, 1, 1})

	scale_mtx := makeIdentity()
	scale_mtx.m[0] = scale.X()
	scale_mtx.m[5] = scale.Y()
	scale_mtx.m[10] = scale.Z()
	mtx := getRotationMatrix(rotation, EULER_XYZ)
	setTranslation(translation, &mtx)

	return scale_mtx.Mul(mtx)
}

func (m *Mesh) String() string {
	return m.stringPrefix("")
}

func (m *Mesh) stringPrefix(prefix string) string {
	s := prefix + "Mesh:\n"
	s += m.Geometry.stringPrefix(prefix+"\t") + "\n"
	for _, mat := range m.Materials {
		s += mat.stringPrefix(prefix+"\t") + "\n"
	}
	return s
}

package ofbx

import (
	"fmt"

	"github.com/oakmound/oak/alg/floatgeom"
)

// Mesh is a geometry made of polygon
// https://help.autodesk.com/view/FBX/2017/ENU/?guid=__cpp_ref_class_fbx_mesh_html
type Mesh struct {
	Object
	Geometry  *Geometry
	scene     *Scene
	Materials []*Material
}

// NewMesh cretes a new stub Object
func NewMesh(scene *Scene, element *Element) *Mesh {
	m := &Mesh{}
	m.Object = *NewObject(scene, element)
	m.Object.is_node = true
	return m
}

// Type returns MESH
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
	s := prefix + "Mesh:" + fmt.Sprintf("%v", m.ID()) + "\n"
	s += m.Geometry.stringPrefix(prefix + "\t")
	for _, mat := range m.Materials {
		s += "\n"
		s += mat.stringPrefix(prefix + "\t")
	}
	return s
}

package ofbx

type Mesh struct {
	Object
	geometry  *Geometry
	scene     *Scene
	materials []*Material
}

func NewMesh(scene *Scene, element *Element) *Mesh {
	m := &Mesh{}
	m.Object = NewObject(scene, element)
	m.Object.is_node = true
	return m
}

func (m *Mesh) Type() Type {
	return MESH
}

func (m *Mesh) getGeometry() *Geometry {
	return m.geometry
}

func (m *Mesh) getGeometricMatrix() Matrix {
	translation := resolveVec3Property(*m, "GeometricTranslation", &Vec3{0, 0, 0})
	rotation := resolveVec3Property(*m, "GeometricRotation", &Vec3{0, 0, 0})
	scale := resolveVec3Property(*m, "GeometricScaling", &Vec3{1, 1, 1})

	scale_mtx := m.makeIdentity()
	scale_mtx.m[0] = float32(scale.x)
	scale_mtx.m[5] = float32(scale.y)
	scale_mtx.m[10] = float32(scale.z)
	mtx := m.getRotationMatrix(rotation, EULER_XYZ)
	setTranslation(translation, &mtx)

	return scale_mtx.Mul(mtx)
}

func (m *Mesh) getMaterial(idx int) []Material {
	return m.materials[idx]
}

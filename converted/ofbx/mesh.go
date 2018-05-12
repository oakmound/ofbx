package ofbx

type Mesh struct {
	Object
}

func NewMesh(scene *Scene, element *IElement) *Mesh {
	return nil
}

func (m *Mesh) Type() Type {
	return MESH
}

func (m *Mesh) getGeometry() *Geometry {
	return nil
}

func (m *Mesh) getGeometricMatrix() Matrix {
	return Matrix{}
}

func (m *Mesh) getMaterial(idx int) []Material {
	return nil
}

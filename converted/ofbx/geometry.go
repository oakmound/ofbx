package ofbx

type Geometry struct {
	Object
}

func NewGeometry(scene *Scene, element *IElement) *Geometry {
	return nil
}

func (g *Geometry) Type() Type {
	return GEOMETRY
}

func (g *Geometry) UVSMax() int {
	return 4
}

func (g *Geometry) getVertices() []Vec3 {
	return nil
}

func (g *Geometry) getNormals() *Vec3 {
	return nil
}

func (g *Geometry) getUVs(index int) *Vec2 {
	return nil
}

func (g *Geometry) getColors() *Vec4 {
	return nil
}

func (g *Geometry) getTangents() *Vec3 {
	return nil
}

func (g *Geometry) getSkin() *Skin {
	return nil
}

func (g *Geometry) getMaterials() *int {
	return nil
}

package ofbx

type Material struct {
	Object
	diffuse_color Color
	textures      [TextureCOUNT]*Texture
}

func NewMaterial(scene *Scene, element *Element) *Material {
	m := &Material{}
	m.Object = NewObject(scene, element)
	return m
}

func (m *Material) Type() Type {
	return MATERIAL
}

func (m *Material) getDiffuseColor() Color {
	return m.diffuse_color
}

func (m *Material) getTexture(typ TextureType) *Texture {
	return m.textures[typ]
}

package ofbx

type Material struct {
	Object
}

func NewMaterial(scene *Scene, element *IElement) *Material {
	return nil
}

func (m *Material) Type() Type {
	return MATERIAL
}

func (m *Material) getDiffuseColor() Color {
	return Color{}
}

func (m *Material) getTexture(typ TextureType) *Texture {
	return nil
}

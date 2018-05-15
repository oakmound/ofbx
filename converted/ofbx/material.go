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

Material::Material(const Scene& _scene, const IElement& _element)
	: Object(_scene, _element) {
}

struct MaterialImpl : Material {
	MaterialImpl(const Scene& _scene, const IElement& _element)
		: Material(_scene, _element) {
		for (const Texture*& tex : textures) tex = nullptr;
	}

	Type getType() const override { return Type::MATERIAL; }

	const Texture* getTexture(Texture::TextureType type) const override { return textures[type]; }
	Color getDiffuseColor() const override { return diffuse_color; }

	const Texture* textures[Texture::TextureType::COUNT];
	Color diffuse_color;
};
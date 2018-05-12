package ofbx

type TextureType int

const (
	DIFFUSE TextureType = iota
	NORMAL  TextureType = iota
	COUNT   TextureType = iota
)

type Texture struct {
	Object
}

func NewTexture(scene *Scene, element *IElement) *Texture {
	return nil
}

func (t *Texture) Type() Type {
	return TEXTURE
}

func (t *Texture) getFileName() DataView {
	return DataView{}
}

func (t *Texture) getRelativeFileName() DataView {
	return DataView{}
}

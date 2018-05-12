package ofbx

type TextureType int

const (
	DIFFUSE TextureType = iota
	NORMAL  TextureType = iota
	COUNT   TextureType = iota
)

type Texture struct {
	Object
	filename          DataView
	relative_filename DataView
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

// Keeping for conversion sake, will update and remove later TODO:Clean out references so it can be removed

func (t *Texture) getType() Type {
	return t.Type()
}

package ofbx

type TextureType int

const (
	DIFFUSE      TextureType = iota
	NORMAL       TextureType = iota
	TextureCOUNT TextureType = iota
)

type Texture struct {
	Object
	filename          *DataView
	relative_filename *DataView
}

func NewTexture(scene *Scene, element *Element) *Texture {
	//TODO: Shoulduse NewObject here
	t := &Texture{
		Object: *NewObject(scene, element),
	}
	return t
}

func (t *Texture) Type() Type {
	return TEXTURE
}

func (t *Texture) getFileName() *DataView {
	return t.filename
}

func (t *Texture) getRelativeFileName() *DataView {
	return t.relative_filename
}

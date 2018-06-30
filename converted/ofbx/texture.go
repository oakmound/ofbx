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

func (t *Texture) String() string {
	s := "Texture: " + Object.String()
	s += ", filename: " + t.filename.String()
	s += ", relative_filename: " + t.relative_filename.String()
	return s
}

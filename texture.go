package ofbx

//TextureType determines how a texture be used
type TextureType int

// Texture type block
const (
	DIFFUSE      TextureType = iota
	NORMAL       TextureType = iota
	TextureCOUNT TextureType = iota
)

//Texture is a texture file on an object
type Texture struct {
	Object
	filename         *DataView
	relativeFilename *DataView
}

// NewTexture creates a texture
func NewTexture(scene *Scene, element *Element) *Texture {
	t := &Texture{
		Object: *NewObject(scene, element),
	}
	return t
}

// Type returns Texture
func (t *Texture) Type() Type {
	return TEXTURE
}

func (t *Texture) getFileName() *DataView {
	return t.filename
}

func (t *Texture) getRelativeFileName() *DataView {
	return t.relativeFilename
}

func (t *Texture) String() string {
	s := "Texture: " + t.Object.String()
	s += ", filename: " + t.filename.String()
	s += ", relativeFilename: " + t.relativeFilename.String()
	return s
}

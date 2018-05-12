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
	//TODO: Shoulduse NewObject here
	t := {
		NewObject(scene, element),
		nil,
		nil
	}
	return t
}

func (t *Texture) Type() Type {
	return TEXTURE
}

func (t *Texture) getFileName() DataView {
	return filename
}

func (t *Texture) getRelativeFileName() DataView {
	return relative_filename
}

// Keeping for conversion sake, will update and remove later TODO:Clean out references so it can be removed

func (t *Texture) getType() Type {
	return t.Type()
}

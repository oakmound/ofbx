package ofbx

import "fmt"

// Material stores texture pointers and how to apply them
type Material struct {
	Object
	EmissiveColor     Color
	EmissiveFactor    float64
	AmbientColor      Color
	DiffuseColor      Color
	DiffuseFactor     float64
	TransparentColor  Color
	SpecularColor     Color
	SpecularFactor    float64
	Shininess         float64
	ShininessExponent float64
	ReflectionColor   Color
	ReflectionFactor  float64
	Textures          [TextureCOUNT]*Texture
}

// NewMaterial makes a stub Material
func NewMaterial(scene *Scene, element *Element) *Material {
	m := &Material{}
	m.Object = *NewObject(scene, element)
	return m
}

// Type returns MATERIAl
func (m *Material) Type() Type {
	return MATERIAL
}

func (m *Material) String() string {
	return m.stringPrefix("")
}

func (m *Material) stringPrefix(prefix string) string {
	s := prefix + "Material: " + "\n"
	s += prefix + "EmissiveColor" + m.EmissiveColor.String() + "\n"
	s += prefix + fmt.Sprintf("EmissiveFactor: %f", m.EmissiveFactor) + "\n"
	s += prefix + "AmbientColor" + m.AmbientColor.String() + "\n"
	s += prefix + "DiffuseColor" + m.DiffuseColor.String() + "\n"
	s += prefix + fmt.Sprintf("DiffuseFactor: %f", m.DiffuseFactor) + "\n"
	s += prefix + "TransparentColor" + m.TransparentColor.String() + "\n"
	s += prefix + "SpecularColor" + m.SpecularColor.String() + "\n"
	s += prefix + fmt.Sprintf("SpecularFactor: %f", m.SpecularFactor) + "\n"
	s += prefix + fmt.Sprintf("Shininess: %f", m.Shininess) + "\n"
	s += prefix + fmt.Sprintf("ShininessExponent: %f", m.ShininessExponent) + "\n"
	s += prefix + "ReflectionColor" + m.ReflectionColor.String() + "\n"
	s += prefix + fmt.Sprintf("ReflectionFactor: %f", m.ReflectionFactor) + "\n"
	if m.Textures[DIFFUSE] != nil {
		s += "Diffuse Texture: " + m.Textures[DIFFUSE].String() + "\n"
	}
	if m.Textures[NORMAL] != nil {
		s += "Normal Texture: " + m.Textures[NORMAL].String() + "\n"
	}
	return s
}

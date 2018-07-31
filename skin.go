package ofbx

// Skin is a mapping for textures that denotes the control points to act on
type Skin struct {
	Object
	Clusters []*Cluster
}

// NewSkin creates a new skin
func NewSkin(scene *Scene, element *Element) *Skin {
	s := Skin{}
	s.Object = *NewObject(scene, element)
	return &s
}

// Type returns skin as type
func (s *Skin) Type() Type {
	return SKIN
}

func (s *Skin) String() string {
	return s.stringPrefix("")
}

func (s *Skin) stringPrefix(prefix string) string {
	str := prefix + "Skin: \n" + s.Object.stringPrefix(prefix+"\t")
	for _, cluster := range s.Clusters {
		str += cluster.stringPrefix(prefix+"\t") + "\n"
	}
	return str
}

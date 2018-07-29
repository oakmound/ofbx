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
	str := "Skin: " + s.Object.String()
	for _, cluster := range s.clusters {
		str += "\t" + cluster.String() + "\n"
	}
	return str
}

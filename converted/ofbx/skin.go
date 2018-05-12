package ofbx

type Skin struct {
	Object
}

func NewSkin(scene *Scene, element *IElement) *Skin {
	return nil
}

func (s *Skin) Type() Type {
	return SKIN
}

func (s *Skin) getCluster() []Cluster {
	return nil
}

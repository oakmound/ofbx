package ofbx

type Skin struct {
	Object
	clusters []*Cluster
}

func NewSkin(scene *Scene, element *Element) *Skin {
	s := Skin{}
	s.Object = *NewObject(scene, element)
	return &s
}

func (s *Skin) Type() Type {
	return SKIN
}

func (s *Skin) getCluster(idx int) *Cluster {
	return s.clusters[idx]
}

func (s *Skin) getClusterCount() int {
	return len(s.clusters)
}

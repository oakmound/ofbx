package ofbx

type Skin struct {
	Object
	clusters []*Cluster
}

func NewSkin(scene *Scene, element *IElement) *Skin {
	//TODO: Shoulduse NewObject here
	s := NewObject(scene, element).(*Skin)
	return s
}

func (s *Skin) Type() Type {
	return SKIN
}

func (s *Skin) getCluster(idx int) *Cluster {
	return clusters[idx]
}

func (s *Skin) getClusterCount() int {
	return len(s.clusters)
}

/// FROM CPP

func (s *Skin) getType() Type {
	return s.Type()
}

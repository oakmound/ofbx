package threefbx

type Object3d struct {
	children []Object3d
	id       int
	name     string
}

func (o *Object3d) getObjectByName(name string) *Object3d {
	for _, c := range o.children {
		if c.name == name {
			return &c
		}
	}
	return nil
}

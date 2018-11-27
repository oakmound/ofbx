package threefbx

type Curve struct {
	*baseModel
	geometry *Geometry
	material Material
}

func NewLine(geometry *Geometry, material Material) *Curve {
	return &Curve{
		geometry: geometry,
		material: material,
	}
}

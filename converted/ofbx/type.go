package ofbx

type Type int

const (
	ROOT                 Type = iota
	GEOMETRY             Type = iota
	MATERIAL             Type = iota
	MESH                 Type = iota
	TEXTURE              Type = iota
	LIMB_NODE            Type = iota
	NULL_NODE            Type = iota
	NODE_ATTRIBUTE       Type = iota
	CLUSTER              Type = iota
	SKIN                 Type = iota
	ANIMATION_STACK      Type = iota
	ANIMATION_LAYER      Type = iota
	ANIMATION_CURVE      Type = iota
	ANIMATION_CURVE_NODE Type = iota
	NOTYPE Type = iota
)

var (
	typeStrings = map[Type]string{
		ROOT:                 "root",
		GEOMETRY:             "geometry",
		MATERIAL:             "material",
		MESH:                 "mesh",
		TEXTURE:              "texture",
		LIMB_NODE:            "limb node",
		NULL_NODE:            "null node",
		NODE_ATTRIBUTE:       "node attribute",
		CLUSTER:              "cluster",
		SKIN:                 "skin",
		ANIMATION_STACK:      "animation stack",
		ANIMATION_LAYER:      "animation layer",
		ANIMATION_CURVE:      "animation curve",
		ANIMATION_CURVE_NODE: "animation curve node",
	}
)

func (t Type) String() string {
	return typeStrings[t]
}

package threefbx

import (
	"github.com/oakmound/ofbx"
)

type Property struct {
	Typ     ofbx.PropertyType
	Payload interface{}
}

func (p Property) IsArray() bool {
	switch p.Typ {
	case ofbx.ArrayBOOL, ofbx.ArrayBYTE, ofbx.ArrayDOUBLE,
		ofbx.ArrayFLOAT, ofbx.ArrayINT, ofbx.ArrayLONG:
		return true
	default:
		return false
	}
}

func NodeProperty(n *Node) Property {
	// Todo: node property type?
	return Property{Payload:n}
}
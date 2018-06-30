package ofbx

import "fmt"

type PropertyType rune

const (
	LONG         PropertyType = 'L'
	INTEGER      PropertyType = 'I'
	STRING       PropertyType = 'S'
	FLOAT        PropertyType = 'F'
	DOUBLE       PropertyType = 'D'
	ARRAY_DOUBLE PropertyType = 'd'
	ARRAY_INT    PropertyType = 'i'
	ARRAY_LONG   PropertyType = 'l'
	ARRAY_FLOAT  PropertyType = 'f'
)

var (
	propertyTypeSizes = map[PropertyType]int{
		DOUBLE:       8,
		INTEGER:      4,
		LONG:         8,
		FLOAT:        4,
		ARRAY_DOUBLE: 8,
		ARRAY_INT:    4,
		ARRAY_LONG:   8,
		ARRAY_FLOAT:  4,
	}
)

func (pt PropertyType) Size() int {
	return propertyTypeSizes[pt]
}

func (pt PropertyType) IsArray() bool {
	switch pt {
	case ARRAY_DOUBLE, ARRAY_FLOAT, ARRAY_INT, ARRAY_LONG:
		return true
	}
	return false
}

type Property struct {
	count            int
	typ              PropertyType
	value            *DataView
	next             *Property
	encoding         uint32
	compressedLength uint32
}

func (p *Property) Type() PropertyType {
	return p.typ
}
func (p *Property) getNext() *Property {
	return p.next
}

func (p *Property) getValue() *DataView {
	return p.value
}

func (p *Property) getCount() int {
	return p.count
}

func (p *Property) getEncoding() uint32 {
	return p.encoding
}

func (p *Property) getValuesF32() ([]float32, error) {
	return parseArrayRawFloat32(p)
}

func (p *Property) getValuesInt64() ([]int64, error) {
	return parseArrayRawInt64(p)
}

func findChildren(element *Element, id string) []*Element {
	iterables := element.children
	for idx, val := range iterables {
		if val.id.String() == id {
			return iterables[idx:]
		}
	}
	return []*Element{}
}

func findChildProperty(element *Element, id string) *Property {
	iterables := element.children
	for idx, val := range iterables {
		if val.id.String() == id {
			return iterables[idx].first_property
		}
	}
	return nil
}

func resolveProperty(obj Obj, name string) *Element {
	props := findChildren(obj.Element(), "Properties70")
	if props == nil {
		return nil
	}

	props = props[0].children
	for _, prop := range props {
		if prop.first_property != nil && prop.first_property.value.String() == name {
			return prop
		}
	}
	return nil
}

func isString(prop *Property) bool {
	if prop == nil {
		return false
	}
	return prop.Type() == STRING
}

func isLong(prop *Property) bool {
	if prop == nil {
		return false
	}
	return prop.Type() == LONG
}

func (p *Property) String() string {
	s := "Property: count=" + fmt.Sprintf("%e", p.count)
	s += ", PropType= " + fmt.Sprintf("%e", p.typ)
	s += ", value= " + p.value.String()
	// TODO: However we reimplement next
	s += ", encoding=" + fmt.Sprintf("%e", p.encoding)
	s += ", compressedLen=" + fmt.Sprintf("%e", p.compressedLength)
	return s
}

// count            int
// 	typ              PropertyType
// 	value            *DataView
// 	next             *Property
// 	encoding         uint32
// 	compressedLength uint32

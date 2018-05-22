package ofbx

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

type Property struct {
	count int
	typ   uint8
	value DataView
	next  *Property
}

func (p *Property) getType() Type {
	return Type(p.typ)
}

func (p *Property) getNext() *Property {
	return p.next
}

func (p *Property) getValue() DataView {
	return p.value
}

func (p *Property) getCount() int {
	return p.count
}

func (p *Property) getValuesF64(values []float64, max_size int) bool {
	return parseArrayRaw(*this, values, max_size)
}

func (p *Property) getValuesInt(values []int, max_size int) bool {
	return parseArrayRaw(*this, values, max_size)
}

func (p *Property) getValuesF32(values []float32, max_size int) bool {
	return parseArrayRaw(*this, values, max_size)
}

func (p *Property) getValuesUInt64(values []uint64, max_size int) bool {
	return parseArrayRaw(*this, values, max_size)
}

func (p *Property) getValuesInt64(values []int64, max_size int) bool {
	return parseArrayRaw(*this, values, max_size)
}

func findChild(element *Element, id string) *Element {
	iter := element.child
	for iter != nil {
		if iter.id == id {
			return iter
		}
		iter = iter.sibling
	}
	return nil
}

func resolveProperty(obj *Object, name string) *Element {
	props := findChild(obj.element, "Properties70")
	if props == nil {
		return nil
	}

	prop := props.child
	for prop != nil {
		if prop.first_property && prop.first_property.value == name {
			return prop
		}
		prop = prop.sibling
	}
	return nil
}

func isString(prop *Property) bool {
	if prop == nil {
		return false
	}
	return prop.getType() == STRING
}

func isLong(prop *Property) bool {
	if prop == nil {
		return false
	}
	return prop.getType() == LONG
}

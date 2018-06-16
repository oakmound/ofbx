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
	count              int
	typ                PropertyType
	value              *DataView
	next               *Property
	encoding           uint32
	compressedLength   uint32
	unCompressedLength uint32
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

func (p *Property) getValuesF32(maxSize int) ([]float32, error) {
	return parseArrayRawFloat32(p, maxSize)
}

func (p *Property) getValuesInt64(maxSize int) ([]int64, error) {
	return parseArrayRawInt64(p, maxSize)
}

func findChild(element *Element, id string) *Element {
	iter := element.child
	for iter != nil {
		if iter.id.String() == id {
			return iter
		}
		iter = iter.sibling
	}
	return nil
}

func resolveProperty(obj Obj, name string) *Element {
	props := findChild(obj.Element(), "Properties70")
	if props == nil {
		return nil
	}

	prop := props.child
	for prop != nil {
		if prop.first_property != nil && prop.first_property.value.String() == name {
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
	return prop.Type() == STRING
}

func isLong(prop *Property) bool {
	if prop == nil {
		return false
	}
	return prop.Type() == LONG
}

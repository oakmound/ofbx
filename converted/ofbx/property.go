package ofbx

import "fmt"

type PropertyType rune

const (
	BOOL         PropertyType = 'C'
	INT16        PropertyType = 'Y'
	LONG         PropertyType = 'L'
	INTEGER      PropertyType = 'I'
	STRING       PropertyType = 'S'
	RAWSTRING    PropertyType = 'R'
	FLOAT        PropertyType = 'F'
	DOUBLE       PropertyType = 'D'
	ARRAY_DOUBLE PropertyType = 'd'
	ARRAY_INT    PropertyType = 'i'
	ARRAY_LONG   PropertyType = 'l'
	ARRAY_FLOAT  PropertyType = 'f'
	ARRAY_BOOL   PropertyType = 'b'
	ARRAY_BYTE   PropertyType = 'c'
)

var (
	propertyTypeSizes = map[PropertyType]int{
		BOOL:         1,
		INT16:        2,
		DOUBLE:       8,
		INTEGER:      4,
		LONG:         8,
		FLOAT:        4,
		ARRAY_DOUBLE: 8,
		ARRAY_INT:    4,
		ARRAY_LONG:   8,
		ARRAY_FLOAT:  4,
		ARRAY_BOOL:   1,
		ARRAY_BYTE:   1,
	}
)

func (p *Property) stringValue() string {

	switch p.typ {
	case BOOL:
		return fmt.Sprintf("%v", p.value.toBool())
	case LONG:
		return fmt.Sprintf("%d", p.value.toint64())
	case INTEGER:
		return fmt.Sprintf("%d", p.value.toInt32())
	case STRING:
		return p.value.String()
	case RAWSTRING:
		return p.value.String()
	case FLOAT:
		return fmt.Sprintf("%f", p.value.toFloat())
	case DOUBLE:
		return fmt.Sprintf("%f", p.value.toDouble())
	case ARRAY_DOUBLE:
		sli, err := parseArrayRawFloat64(p)
		if err != nil {
			return "Bad Format" + err.Error()
		}
		return fmt.Sprintf("%v", sli)
	case ARRAY_INT:
		sli, err := parseArrayRawInt(p)
		if err != nil {
			return "Bad Format" + err.Error()
		}
		return fmt.Sprintf("%v", sli)
	case ARRAY_LONG:
		sli, err := parseArrayRawInt64(p)
		if err != nil {
			return "Bad Format" + err.Error()
		}
		return fmt.Sprintf("%v", sli)
	case ARRAY_FLOAT:
		sli, err := parseArrayRawFloat32(p)
		if err != nil {
			return "Bad Format" + err.Error()
		}
		return fmt.Sprintf("%v", sli)
	case ARRAY_BOOL:
		return "Bool array not implemented"
	case ARRAY_BYTE:
		return "Byte array not implemented"
	}

	return "Error: Not a known property Type " + string(p.typ)
}

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
	encoding         uint32
	compressedLength uint32
}

func (p *Property) Type() PropertyType {
	return p.typ
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

func findSingleChildProperty(element *Element, id string) *Property {
	iterables := element.children
	for idx, val := range iterables {
		if val.id.String() == id {
			if len(iterables[idx].properties) > 0 {
				return iterables[idx].properties[0]
			}
		}
	}
	return nil
}

func findChildProperty(element *Element, id string) []*Property {
	iterables := element.children
	for idx, val := range iterables {
		if val.id.String() == id {
			return iterables[idx].properties
		}
	}
	return nil
}

func resolveProperty(obj Obj, name string) *Element {
	elems := findChildren(obj.Element(), "Properties70")
	if elems == nil {
		return nil
	}

	elems = elems[0].children
	for _, elem := range elems {
		if prop := elem.getProperty(0); prop != nil && prop.value.String() == name {
			return elem
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
	return p.stringPrefix("")
}

func (p *Property) stringPrefix(prefix string) string {
	s := prefix + p.stringValue()
	// s += ", proptype= " + fmt.Sprintf("%q", p.typ)
	// s += "count=" + fmt.Sprintf("%d", p.count)
	// s += ", encoding=" + fmt.Sprintf("%d", p.encoding)
	// s += ", compressedLen=" + fmt.Sprintf("%d", p.compressedLength)
	return s
}

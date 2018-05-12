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

type IElementProperty struct {
}

func (iep *IElementProperty) getType() Type {
	return 0
}

func (iep *IElementProperty) getNext() *IElementProperty {
	return nil
}

func (iep *IElementProperty) getValue() DataView {
	return DataView{}
}

func (iep *IElementProperty) getCount() int {
	return 0
}

func (iep *IElementProperty) getValuesF64(values []float64, max_size int) bool {
	return false
}

func (iep *IElementProperty) getValuesInt(values []int, max_size int) bool {
	return false
}

func (iep *IElementProperty) getValuesF32(values []float32, max_size int) bool {
	return false
}

func (iep *IElementProperty) getValuesUInt64(values []uint64, max_size int) bool {
	return false
}

func (iep *IElementProperty) getValuesInt64(values []int64, max_size int) bool {
	return false
}

type IElement struct{}

func (ie *IElement) getFirstChild() *IElement {
	return nil
}
func (ie *IElement) getSibling() *IElement {
	return nil
}
func (ie *IElement) getID() DataView {
	return DataView{}
}
func (ie *IElement) getFirstProperty() *IElementProperty {
	return nil
}

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

type ElementProperty struct {
}

func (ep *ElementProperty) getType() Type {
	return 0
}

func (ep *ElementProperty) getNext() *ElementProperty {
	return nil
}

func (ep *ElementProperty) getValue() DataView {
	return DataView{}
}

func (ep *ElementProperty) getCount() int {
	return 0
}

func (ep *ElementProperty) getValuesF64(values []float64, max_size int) bool {
	return false
}

func (ep *ElementProperty) getValuesInt(values []int, max_size int) bool {
	return false
}

func (ep *ElementProperty) getValuesF32(values []float32, max_size int) bool {
	return false
}

func (ep *ElementProperty) getValuesUInt64(values []uint64, max_size int) bool {
	return false
}

func (ep *ElementProperty) getValuesInt64(values []int64, max_size int) bool {
	return false
}


struct Property : IElementProperty
{
	~Property() { delete next; }
	Type getType() const override { return (Type)type; }
	IElementProperty* getNext() const override { return next; }
	DataView getValue() const override { return value; }
	int getCount() const override
	{
		assert(type == ARRAY_DOUBLE || type == ARRAY_INT || type == ARRAY_FLOAT || type == ARRAY_LONG);
		if (value.is_binary)
		{
			return int(*(uint32*)value.begin);
		}
		return count;
	}

	bool getValues(double* values, int max_size) const override { return parseArrayRaw(*this, values, max_size); }

	bool getValues(float* values, int max_size) const override { return parseArrayRaw(*this, values, max_size); }

	bool getValues(uint64* values, int max_size) const override { return parseArrayRaw(*this, values, max_size); }
	
	bool getValues(int64* values, int max_size) const override { return parseArrayRaw(*this, values, max_size); }

	bool getValues(int* values, int max_size) const override { return parseArrayRaw(*this, values, max_size); }

	int count;
	uint8 type;
	DataView value;
	Property* next = nullptr;
};

static const Element* findChild(const Element& element, const char* id)
{
	Element* const* iter = &element.child;
	while (*iter)
	{
		if ((*iter).id == id) return *iter;
		iter = &(*iter).sibling;
	}
	return nullptr;
}


static IElement* resolveProperty(const Object& obj, const char* name)
{
	const Element* props = findChild((const Element&)obj.element, "Properties70");
	if (!props) return nullptr;

	Element* prop = props.child;
	while (prop)
	{
		if (prop.first_property && prop.first_property.value == name)
		{
			return prop;
		}
		prop = prop.sibling;
	}
	return nullptr;
}

type Element struct {
	id             DataView
	child          *Element
	sibling        *Element
	first_property *Property
}

func (e *Element) getFirstChild() *Element {
	return e.child
}
func (e *Element) getSibling() *Element {
	return e.sibling
}
func (e *Element) getID() DataView {
	return e.id
}
func (e *Element) getFirstProperty() *ElementProperty {
	return e.first_property
}

func (e *Element) getProperty(idx int) *ElementProperty {
	prop := e.first_property
	for i := 0; i < idx; i++ {
		if prop == nil {
			return nil
		}
		prop = prop.getNext()
	}
	return prop
}

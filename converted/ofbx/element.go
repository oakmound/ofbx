package ofbx

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
func (e *Element) getFirstProperty() *Property {
	return e.first_property
}

func (e *Element) getProperty(idx int) *Property {
	prop := e.first_property
	for i := 0; i < idx; i++ {
		if prop == nil {
			return nil
		}
		prop = prop.getNext()
	}
	return prop
}

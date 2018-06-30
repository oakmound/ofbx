package ofbx

type Element struct {
	id             *DataView
	children       []*Element
	first_property *Property
}

func (e *Element) getChildren() []*Element {
	return e.children
}

func (e *Element) getID() *DataView {
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

func (e *Element) String() string {
	return e.stringPrefix("")
}

func (e *Element) stringPrefix(prefix string) string {
	s := prefix + "Element: " + e.id.String()
	if len(e.children) != 0 {
		s += ", children: "
		for _, c := range e.children {
			s += c.stringPrefix("\t" + prefix)
		}
	}
	return s + "\n"
}

package ofbx

type Element struct {
	id         *DataView
	children   []*Element
	properties []*Property
}

func (e *Element) getChildren() []*Element {
	return e.children
}

func (e *Element) getID() *DataView {
	return e.id
}

func (e *Element) getProperty(idx int) *Property {
	if len(e.properties) <= idx {
		return nil
	}
	return e.properties[idx]
}

func (e *Element) String() string {
	return e.stringPrefix("")
}

func (e *Element) stringPrefix(prefix string) string {
	s := prefix + "Element: "
	if e.id != nil {
		s += e.id.String()
	}
	if len(e.children) != 0 {
		s += " children: " + "\n"
		for _, c := range e.children {
			s += c.stringPrefix("\t" + prefix)
		}
		return s
	}
	return s + "\n"
}

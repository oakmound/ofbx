package ofbx

import "strconv"

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
	if len(e.properties) > 1 {
		if fmter, ok := propFormats[e.properties[1].String()]; ok {
			s += " " + fmter(e.properties)
		} else {
			for idx, p := range e.properties {
				v := p.stringPrefix("\t" + prefix + "prop" + strconv.Itoa(idx) + "=")
				if v == "" {
					continue
				}
				s += "\n" + v
			}
		}
	} else if len(e.properties) == 1 {
		s += e.properties[0].stringPrefix(", prop=")
	}
	if len(e.children) != 0 {
		s += "\n"
		//s += "\n" + prefix + "children: " + "\n"
		for _, c := range e.children {
			s += c.stringPrefix(prefix + "\t")
		}
		return s
	}
	return s + "\n"
}

type propertyFormat func([]*Property) string

func numberPropFormat(props []*Property) string {
	if len(props) != 5 {
		return "bad number"
	}
	return props[0].String() + "=" + props[4].String()
}

func colorPropFormat(props []*Property) string {
	if len(props) != 7 {
		return "bad color"
	}
	s := props[0].String()
	s += ":"
	s += " R=" + props[4].String()
	s += " G=" + props[5].String()
	s += " B=" + props[6].String()
	return s
}

func vectorPropFormat(props []*Property) string {
	if len(props) != 7 {
		return "bad vector"
	}
	s := props[0].String()
	s += ":"
	s += " X=" + props[4].String()
	s += " Y=" + props[5].String()
	s += " Z=" + props[6].String()
	return s
}

var (
	propFormats = map[string]propertyFormat{
		"Number":          numberPropFormat,
		"int":             numberPropFormat,
		"enum":            numberPropFormat,
		"KTime":           numberPropFormat,
		"Color":           colorPropFormat,
		"Lcl Scaling":     vectorPropFormat,
		"Lcl Translation": vectorPropFormat,
		"Lcl Rotation":    vectorPropFormat,
		// "d|X":              numberPropFormat,
		// "d|Y":              numberPropFormat,
		// "d|Z":              numberPropFormat,
		// "LocalStop":        numberPropFormat,
		// "ReferenceStop":    numberPropFormat,
		// "EmissiveColor":    colorPropFormat,
		// "AmbientColor":     colorPropFormat,
		// "DiffuseColor":     colorPropFormat,
		// "TransparentColor": colorPropFormat,
		// "SpecularColor":    colorPropFormat,
	}
)

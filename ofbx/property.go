package threefbx

type Property interface {
	IsArray() bool
	Payload() interface{}
}

type SimpleProperty struct {
	payload interface{}
}

func (sp *SimpleProperty) Payload() interface{} {
	return sp.payload
}

func (sp *SimpleProperty) IsArray() bool {
	return false
}

type ArrayProperty struct {
	payload interface{}
}

func (ap *ArrayProperty) IsArray() bool {
	return true
}

func (ap *ArrayProperty) Payload() interface{} {
	return ap.payload
}

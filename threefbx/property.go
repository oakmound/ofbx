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

type MapProperty struct {
	m map[string]Property
}

func (mp *MapProperty) IsArray() bool {
	return false
}

func (mp *MapProperty) Payload() interface{} {
	return mp.m
}

type IDMapProperty struct {
	m map[int64]Property
}

func (mp *IDMapProperty) IsArray() bool {
	return false
}

func (mp *IDMapProperty) Payload() interface{} {
	return mp.m
}

package threefbx

type IDType = string

// Node is a typed object
type Node struct {
	ID       IDType
	attrName string
	attrType string
	name     string

	singleProperty bool

	a            Property
	props        map[string]Property
	propertyList []Property
}

// NewNode creates a new node
func NewNode(name string) *Node {
	n := &Node{}
	n.name = name
	n.props = make(map[string]Property)
	return n
}
func (n *Node) IsArray() bool {
	return false
}
func (n *Node) Payload() interface{} {
	//Maybe return property map?
	return n
}

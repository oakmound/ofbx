package threefbx

// Node is a typed object
type Node struct {
	ID       int
	attrName string
	attrType string
	name     string

	singleProperty bool

	a            Property
	props        map[string]Property
	poseNode     []Property
	propertyList []Property
	connections  []Property
}

// NewNode creates a new node
func NewNode(name string) *Node {
	n := &Node{}
	n.name = name
	return n
}
func (n *Node) IsArray() bool {
	return false
}
func (n *Node) Payload() interface{} {
	//Maybe return property map?
	return n
}

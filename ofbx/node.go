package threefbx

// Node is a typed object
type Node struct {
	ID       int64
	attrName string
	attrType string
	name     string

	singleProperty bool

	props    map[string]Property
	poseNode []Property
}

// NewNode creates a new node
func NewNode(name string, t, typ Type) *Node {
	n := &Node{}
	n.Name = name
	n.isNode = true
	n.typ = typ

	return n
}

func (n *Node) String() string {
	return n.stringPrefix("")
}
func (n *Node) stringPrefix(prefix string) string {
	return prefix + n.typ.String() + ":\n" + n.Object.stringPrefix("\t"+prefix)
}

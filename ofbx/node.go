package threefbx

// Node is a typed object
type Node struct {
	id    int64
	name  string
	typ   Type
	props map[string]Property
}

// NewNode creates a new node
func NewNode(name string, t, typ Type) *Node {
	n := &Node{}
	n.Name = name
	n.isNode = true
	n.typ = typ

	return n
}

// Type returns a nodes type
func (n *Node) Type() Type {
	return n.typ
}

func (n *Node) String() string {
	return n.stringPrefix("")
}
func (n *Node) stringPrefix(prefix string) string {
	return prefix + n.typ.String() + ":\n" + n.Object.stringPrefix("\t"+prefix)
}

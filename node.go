package ofbx

// Node is a typed object
type Node struct {
	Object
	typ Type
}

// NewNode creates a new node
func NewNode(scene *Scene, element *Element, typ Type) *Node {
	n := &Node{}
	n.Object = *NewObject(scene, element)
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

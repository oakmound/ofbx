package threefbx

import (
	"errors"
)

// Connection is a connection from an Object to either another Object or a Property
type Connection struct {
	Typ      ConnectionType
	From, To IDType
	Property string
}

// ConnectionType dictates what the Object is connecting to
type ConnectionType int

// Connection Types
const (
	// ObjectConn is a connection to another Object
	ObjectConn ConnectionType = iota
	// PropConn is a connection to a proprety
	PropConn ConnectionType = iota
)

type ParsedConnections map[IDType]ConnectionSet

type ConnectionSet struct {
	parents  []Relationship
	children []Relationship
}

type Relationship struct {
	ID       IDType
	Property string
}

func NewParsedConnections() ParsedConnections {
	return ParsedConnections(make(map[IDType]ConnectionSet))
}

func NewConnection(n *Node) (Connection, error) {
	cn := Connection{}
	var ok bool
	cn.To, ok = n.props["To"].Payload.(IDType)
	if !ok {
		return cn, errors.New("Node lacking Connection properties")
	}
	cn.From, ok = n.props["From"].Payload.(IDType)
	if !ok {
		return cn, errors.New("Node lacking Connection properties")
	}
	cn.Property, ok = n.props["Property"].Payload.(string)
	if !ok {
		return cn, errors.New("Node lacking Connection properties")
	}
	return cn, nil
}

func (l *Loader) parseConnections() (ParsedConnections, error) {
	cns := NewParsedConnections()
	for _, n := range l.tree.Objects["Connections"] {
		cn, err := NewConnection(n)
		if err != nil {
			return nil, err
		}
		cf := cns[cn.From]
		ct := cns[cn.To]
		pr := Relationship{ID: cn.To, Property: cn.Property}
		cr := Relationship{ID: cn.From, Property: cn.Property}
		cf.parents = append(cf.parents, pr)
		ct.children = append(ct.children, cr)
		cns[cn.From] = cf
		cns[cn.To] = ct
	}
	return cns, nil
}

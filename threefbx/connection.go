package threefbx

import (
	"errors"
	"fmt"
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

func NewConnection(props []Property) (Connection, error) {
	cn := Connection{}
	var ok bool
	cn.To, ok = props[0].Payload.(IDType)
	if !ok {
		return cn, errors.New("Node lacking Connection properties: To")
	}
	cn.From, ok = props[1].Payload.(IDType)
	if !ok {
		return cn, errors.New("Node lacking Connection properties: From")
	}
	cn.Property, ok = props[2].Payload.(string)
	if !ok {
		return cn, errors.New("Node lacking Connection properties: Property")
	}
	return cn, nil
}

func (l *Loader) parseConnections() (ParsedConnections, error) {
	cns := NewParsedConnections()
	fmt.Println("Raw connections", l.rawConnections)
	for _, cn := range l.rawConnections {
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

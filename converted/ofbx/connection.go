package ofbx

import "fmt"

// Connection is a connection from an Object to either another Object or a Property
type Connection struct {
	typ      ConnectionType
	from, to uint64
	property string
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

func (ct ConnectionType) String() string {
	if ct == ObjectConn {
		return "OBJ->OBJ"
	}
	return "OBJ->PROP"
}

func (c *Connection) String() string {
	s := "Connection: " + c.typ.String()
	s += " from=" + fmt.Sprintf("%d", c.from)
	s += " to=" + fmt.Sprintf("%d", c.to)
	s += " property=" + c.property
	return s
}

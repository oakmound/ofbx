package ofbx

import "fmt"

type Connection struct {
	typ      ConnectionType
	from, to uint64
	property string
}

type ConnectionType int

// Connection Types
const (
	OBJECT_OBJECT   ConnectionType = iota
	OBJECT_PROPERTY ConnectionType = iota
)

func (ct ConnectionType) String() string {
	if ct == OBJECT_OBJECT {
		return "OBJ->OBJ"
	}
	return "OBJ->PROP"
}

func (c *Connection) String() string {
	s := "Connection: " + c.typ.String()
	s += " from=" + fmt.Sprintf("%e", c.from)
	s += " to=" + fmt.Sprintf("%e", c.to)
	s += " property=" + c.property
	return s
}

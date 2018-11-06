package threefbx

import "errors"

type ParsedConnections map[int]ConnectionSet

type ConnectionSet struct {
	parents  []Connection
	children []Connection
}

func NewParsedConnections() ParsedConnections {
	return ParsedConnections(make(map[int]ConnectionSet))
}

type Connection struct {
	ID           int
	To, From     int
	Relationship string
}

func NewConnection(n *Node) (Connection, error) {
	cn := Connection{}
	var ok bool
	cn.ID, ok = n.props["ID"].Payload().(int)
	if !ok {
		return cn, errors.New("Node lacking Connection properties")
	}
	cn.To, ok = n.props["To"].Payload().(int)
	if !ok {
		return cn, errors.New("Node lacking Connection properties")
	}
	cn.From, ok = n.props["From"].Payload().(int)
	if !ok {
		return cn, errors.New("Node lacking Connection properties")
	}
	cn.Relationship, ok = n.props["Relationship"].Payload().(string)
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
		cf.parents = append(cf.parents, cn)
		ct.children = append(ct.children, cn)
		cns[cn.From] = cf
		cns[cn.To] = ct
	}
	return cns, nil
}

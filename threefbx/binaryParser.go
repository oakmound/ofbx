package threefbx

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/oakmound/ofbx"
)

func (l *Loader) ParseBinary(r io.Reader) (*Tree, error) {
	reader := NewBinaryReader(r, true)
	// We already read first 21 bytes
	reader.r.Discard(2) // skip reserved bytes

	var version = reader.getUint32()
	fmt.Println("FBX binary version: ", version)
	var allNodes = &Tree{}
	for {
		node, err := l.parseBinaryNode(reader, int(version))
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if node != nil {
			allNodes.Objects[node.name][node.ID] = node
		}
	}
	return allNodes, nil
}

// recursively parse nodes until the end of the file is reached
func (l *Loader) parseBinaryNode(r *BinaryReader, version int) (*Node, error) {
	n := NewNode("")
	// The first three data sizes depends on version.
	var err error
	var nodeEnd uint64
	var numProperties uint64
	if version >= 7500 {
		nodeEnd = r.getUint64()
		numProperties = r.getUint64()
		// propertiesListLen
		r.getUint64()
	} else {
		nodeEnd = uint64(r.getUint32())
		numProperties = uint64(r.getUint32())
		// propertiesListLen
		r.getUint32()
	}
	name := r.getShortString()
	// Regards this node as NULL-record if nodeEnd is zero
	if nodeEnd == 0 {
		return nil, nil
	}

	fmt.Println("Properties:")
	propertyList := make([]Property, numProperties)
	for i := uint64(0); i < numProperties; i++ {
		propertyList[i], err = l.parseBinaryProperty(r)
		if err != nil {
			return nil, err
		}
		fmt.Println("	", propertyList[i])
	}

	// check if this node represents just a single property
	// like (name, 0) set or (name2, [0, 1, 2]) set of {name: 0, name2: [0, 1, 2]}
	if numProperties == 1 && uint64(r.r.ReadSoFar()) == nodeEnd {
		n.singleProperty = true
	}
	for nodeEnd > uint64(r.r.ReadSoFar()) {
		subNode, err := l.parseBinaryNode(r, version)
		if err != nil {
			return nil, err
		}
		if subNode != nil {
			l.parseBinarySubNode(name, n, subNode)
		}
	}
	n.propertyList = propertyList // raw property list used by parent
	n.name = name
	// Regards the first three elements in propertyList as id, attrName, and attrType
	if len(propertyList) == 0 {
		return n, nil
	}
	if s, ok := propertyList[0].Payload.(string); ok {
		n.ID = s
	} else if i, ok := propertyList[0].Payload.(int); ok {
		n.ID = strconv.Itoa(i)
	} else {
		return nil, fmt.Errorf("Expected IDType for ID %T:%v", propertyList[0].Payload, propertyList[0].Payload)
	}
	if len(propertyList) == 1 {
		return n, nil
	}
	if s, ok := propertyList[1].Payload.(string); ok {
		n.attrName = s
	} else {
		return nil, errors.New("Expected string type for attrName")
	}
	if len(propertyList) == 2 {
		return n, nil
	}
	if s, ok := propertyList[2].Payload.(string); ok {
		n.attrType = s
	} else {
		return nil, errors.New("Expected string type for attrType")
	}
	return n, nil
}

func (l *Loader) parseBinarySubNode(name string, root, child *Node) error {
	// special case: child node is single property
	if child.singleProperty {
		value := child.propertyList[0]
		if value.IsArray() {
			root.props[child.name] = NodeProperty(child)
			child.a = value
		} else {
			root.props[child.name] = value
		}
	} else if name == "Connections" && child.name == "C" {
		props := child.propertyList
		conn := ofbx.Connection{}
		connType, ok := props[0].Payload.(string)
		if !ok {
			return errors.New("Expected string for connection type")
		}
		switch connType {
		case "OO":
			conn.Typ = ofbx.ObjectConn
		case "OP":
			conn.Typ = ofbx.PropConn
			if len(props) > 3 {
				conn.Property, ok = props[3].Payload.(string)
				if !ok {
					return errors.New("Expected string for connection property")
				}
			}
		default:
			return errors.New("Unknown connection type " + connType)
		}
		conn.From, ok = props[1].Payload.(uint64)
		if !ok {
			return errors.New("Expected uint64 for conn.From")
		}
		conn.To, ok = props[2].Payload.(uint64)
		if !ok {
			return errors.New("Expected uint64 for conn.To")
		}
		// Javascript discards FBX connection type, we keep it
		l.rawConnections = append(l.rawConnections, props)
	} else if child.name == "Properties70" {
		for k, v := range child.props {
			root.props[k] = v
		}
	} else if name == "Properties70" && child.name == "P" {
		innerPropName, ok := child.propertyList[0].Payload.(string)
		if !ok {
			return errors.New("Expected string inner property name")
		}
		inPropType, ok := child.propertyList[1].Payload.(string)
		if !ok {
			return errors.New("Expected string inner property type")
		}
		inPropType2 := child.propertyList[2]
		innerPropFlag := child.propertyList[3]
		var innerPropValue Property

		if strings.HasPrefix(innerPropName, "Lcl ") {
			innerPropName = "Lcl_" + innerPropName[3:]
		}
		if strings.HasPrefix(inPropType, "Lcl ") {
			inPropType = "Lcl_" + inPropType[3:]
		}

		if inPropType == "Color" || inPropType == "ColorRGB" || inPropType == "Vector" || inPropType == "Vector3D" || strings.HasPrefix(inPropType, "Lcl_") {
			innerPropValue = Property{
				Payload: []interface{}{
					child.propertyList[4].Payload,
					child.propertyList[5].Payload,
					child.propertyList[6].Payload,
				},
			}
		} else {
			innerPropValue = child.propertyList[4]
		}
		// this will be copied to parent, see above
		root.props[innerPropName] = Property{Payload: map[string]Property{
			"type":  Property{Payload: inPropType},
			"type2": inPropType2,
			"flag":  innerPropFlag,
			"value": innerPropValue,
		}}
	} else if _, ok := root.props[child.name]; !ok {
		root.props[child.name] = Property{Payload: map[IDType]Property{child.ID: NodeProperty(child)}}
	} else {
		if child.name == "PoseNode" {
			if !root.props[child.name].IsArray() {
				// Patrick: Ugh??
				root.props[child.name] = Property{Payload: []interface{}{root.props[child.name], child}}
				return nil
			}
			pay := root.props[child.name].Payload
			root.props[child.name] = Property{Payload: append(pay.([]interface{}), child)}
		} else {
			prop, ok := root.props[child.name]
			if !ok {
				return nil
			}
			m, ok := prop.Payload.(map[IDType]Property)
			if !ok {
				return nil
			}
			_, ok = m[child.ID]
			if ok {
				return nil
			}
			m[child.ID] = NodeProperty(child)
			root.props[child.name] = Property{Payload: m}
		}
	}
	return nil
}

func (l *Loader) parseBinaryProperty(r *BinaryReader) (Property, error) {
	ty := r.getString(1)
	switch ty {
	case "C":
		return Property{ofbx.PropertyType(ty[0]), r.getBoolean()}, nil
	case "D":
		return Property{ofbx.PropertyType(ty[0]), r.getFloat64()}, nil
	case "F":
		return Property{ofbx.PropertyType(ty[0]), r.getFloat32()}, nil
	case "I":
		return Property{ofbx.PropertyType(ty[0]), r.getInt32()}, nil
	case "L":
		return Property{ofbx.PropertyType(ty[0]), r.getInt64()}, nil
	case "R":
		return Property{ofbx.PropertyType(ty[0]), r.getArrayBuffer(r.getUint32())}, nil
	case "S":
		return Property{ofbx.PropertyType(ty[0]), r.getString(r.getUint32())}, nil
	case "Y":
		return Property{ofbx.PropertyType(ty[0]), r.getInt16()}, nil
	case "b":
	case "c":
	case "d":
	case "f":
	case "i":
	case "l":
		arrayLength := r.getUint32()
		encoding := r.getUint32() // 0: non-compressed, 1: compressed
		compressedLength := r.getUint32()
		if encoding == 0 {
			switch ty {
			case "b":
			case "c":
				return Property{ofbx.PropertyType(ty[0]), r.getBooleanArray(arrayLength)}, nil
			case "d":
				return Property{ofbx.PropertyType(ty[0]), r.getFloat64Array(arrayLength)}, nil
			case "f":
				return Property{ofbx.PropertyType(ty[0]), r.getFloat32Array(arrayLength)}, nil
			case "i":
				return Property{ofbx.PropertyType(ty[0]), r.getInt32Array(arrayLength)}, nil
			case "l":
				return Property{ofbx.PropertyType(ty[0]), r.getInt64Array(arrayLength)}, nil
			}
		}
		buff := r.getArrayBuffer(compressedLength)
		r2, err := zlib.NewReader(bytes.NewReader(buff))
		if err != nil {
			return Property{}, err
		}
		defer r2.Close()
		r3 := NewBinaryReader(r2, false)
		switch ty {
		case "b":
		case "c":
			return Property{ofbx.PropertyType(ty[0]), r3.getBooleanArray(arrayLength)}, nil
		case "d":
			return Property{ofbx.PropertyType(ty[0]), r3.getFloat64Array(arrayLength)}, nil
		case "f":
			return Property{ofbx.PropertyType(ty[0]), r3.getFloat32Array(arrayLength)}, nil
		case "i":
			return Property{ofbx.PropertyType(ty[0]), r3.getInt32Array(arrayLength)}, nil
		case "l":
			return Property{ofbx.PropertyType(ty[0]), r3.getInt64Array(arrayLength)}, nil
		}
	}
	return Property{}, errors.New("Undefined property type: " + ty)
}

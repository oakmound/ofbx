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
	reader.r.cr.ReadSoFar += 21
	reader.r.Discard(2) // skip reserved bytes

	var version = reader.getUint32()
	fmt.Println("FBX binary version: ", version)
	var allNodes = NewTree()
	for {

		node, err := l.parseBinaryNode(reader, int(version))
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if node == nil {
			break
		}
		if _, ok := allNodes.Objects[node.name]; !ok {
			allNodes.Objects[node.name] = make(map[IDType]*Node)
		}
		allNodes.Objects[node.name][node.ID] = node

	}
	return allNodes, nil
}

// recursively parse nodes until the end of the file is reached
func (l *Loader) parseBinaryNode(r *BinaryReader, version int) (*Node, error) {
	v, _ := r.r.Peek(12)
	footer := true
	for _, b := range v {
		if b != 0 {
			footer = false
			break
		}
	}
	if footer {
		// Note we don't actually read the footer contents yet,
		// as far as we know the footer holds no useful information
		//fmt.Println("Returning footer")
		return nil, nil
	}
	n := NewNode("")
	// The first three data sizes depends on version.
	var err error
	var nodeEnd uint64
	var numProperties uint64
	blockSentinelLength := 13

	if version >= 7500 {
		nodeEnd = r.getUint64()
		numProperties = r.getUint64()
		// propertiesListLen
		r.getUint64()
		blockSentinelLength = 25
	} else {
		nodeEnd = uint64(r.getUint32())
		numProperties = uint64(r.getUint32())
		// propertiesListLen
		r.getUint32()
	}
	n.name = r.getShortString()

	// Regards this node as NULL-record if nodeEnd is zero
	if nodeEnd == 0 {
		return nil, nil
	}

	fmt.Println("Read A:", r.r.ReadSoFar(), nodeEnd)

	fmt.Println("Properties(", n.name, " ", numProperties, nodeEnd, "):")
	propertyList := make([]Property, numProperties)
	for i := uint64(0); i < numProperties; i++ {
		propertyList[i], err = l.parseBinaryProperty(r)
		if err != nil {
			return nil, err
		}
		fmt.Println("	", propertyList[i])
	}
	n.propertyList = propertyList // raw property list used by parent

	// check if this node represents just a single property
	// like (name, 0) set or (name2, [0, 1, 2]) set of {name: 0, name2: [0, 1, 2]}
	if numProperties == 1 && uint64(r.r.ReadSoFar()) == nodeEnd-uint64(blockSentinelLength) {
		n.singleProperty = true
	}

	fmt.Println("Read B:", r.r.ReadSoFar(), nodeEnd)

	if uint64(r.r.ReadSoFar()) < nodeEnd {
		for uint64(r.r.ReadSoFar()) < nodeEnd-uint64(blockSentinelLength) {
			subNode, err := l.parseBinaryNode(r, version)
			if err != nil {
				return nil, err
			}
			if subNode == nil {
				break
			}
			err = l.parseBinarySubNode(n.name, n, subNode)
			if err != nil {
				return nil, err
			}
		}
		fmt.Println("About to discard", r.r.ReadSoFar(), blockSentinelLength)
		r.r.Discard(blockSentinelLength)
	} else {
		fmt.Println("No sentinel??")
	}

	fmt.Println("Read C:", r.r.ReadSoFar(), nodeEnd)

	// Regards the first three elements in propertyList as id, attrName, and attrType
	if len(propertyList) == 0 {
		return n, nil
	}
	if s, ok := propertyList[0].Payload.(string); ok {
		n.ID = s
	} else if i, ok := propertyList[0].Payload.(int32); ok {
		n.ID = strconv.Itoa(int(i))
	} else if i, ok := propertyList[0].Payload.(int64); ok {
		n.ID = strconv.Itoa(int(i))
	}
	if len(propertyList) == 1 {
		return n, nil
	}
	if s, ok := propertyList[1].Payload.(string); ok {
		n.attrName = s
	}
	if len(propertyList) == 2 {
		return n, nil
	}
	if s, ok := propertyList[2].Payload.(string); ok {
		n.attrType = s
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
		fmt.Println("Found connections sub node")
		props := child.propertyList
		conn := Connection{}
		connType, ok := props[0].Payload.(string)
		if !ok {
			return errors.New("Expected string for connection type")
		}
		switch connType {
		case "OO":
			conn.Typ = ObjectConn
		case "OP":
			conn.Typ = PropConn
			if len(props) > 3 {
				conn.Property, ok = props[3].Payload.(string)
				if !ok {
					return errors.New("Expected string for connection property")
				}
			}
		default:
			return errors.New("Unknown connection type " + connType)
		}
		from, ok := props[1].Payload.(int64)
		if !ok {
			return fmt.Errorf("Expected int64 for conn.From %t:%v", props[1].Payload, props[1].Payload)
		}
		to, ok := props[2].Payload.(int64)
		if !ok {
			return errors.New("Expected int64 for conn.To")
		}
		conn.From = strconv.FormatInt(from, 10)
		conn.To = strconv.FormatInt(to, 10)
		// Javascript discards FBX connection type, we keep it
		l.rawConnections = append(l.rawConnections, conn)
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
		} else if len(child.propertyList) > 4 {

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
	case "b", "c", "d", "f", "i", "l":
		arrayLength := r.getUint32()
		encoding := r.getUint32() // 0: non-compressed, 1: compressed
		compressedLength := r.getUint32()
		if encoding == 0 {
			switch ty {
			case "b", "c":
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
		case "b", "c":
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

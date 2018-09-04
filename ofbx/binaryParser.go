package threefbx

import (
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"strings"
)

func (l *Loader) parseBinary(r io.Reader) (FBXTree, error) {
	reader := NewBinaryReader(r, true)
	// We already read first 21 bytes
	reader.Discard(2) // skip reserved bytes

	var version = reader.getUint32()
	fmt.Println("FBX binary version: " + version)
	var allNodes = NewFBXTree()
	for {
		node, err := l.parseBinaryNode(reader, version)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if node != nil {
			allNodes[node.name] = node
		}
	}
	return allNodes, nil
}

// recursively parse nodes until the end of the file is reached
func (l *Loader) parseBinaryNode(r *BinaryReader, version int) (*Node, error) {
	var n Node
	// The first three data sizes depends on version.
	var endOffset uint64
	var numProperties uint64
	var propertiesListLen uint64
	if version >= 7500 {
		nodeEnd = r.getUint64()
		numProperties = r.getUint64()
		propertiesListLen = r.getUint64()
	} else {
		nodeEnd = uint64(r.getUint32())
		numProperties = uint64(r.getUint32())
		propertiesListLen = uint64(r.getUint32())
	}
	name := r.getShortString()
	// Regards this node as NULL-record if nodeEnd is zero
	if nodeEnd == 0 {
		return nil, nil
	}

	propertyList := make([]Property, numProperties)
	for i := 0; i < numProperties; i++ {
		propertyList[i] = l.parseBinaryProperty(reader)
	}
	// check if this node represents just a single property
	// like (name, 0) set or (name2, [0, 1, 2]) set of {name: 0, name2: [0, 1, 2]}
	if numProperties == 1 && r.ReadSoFar() == nodeEnd {
		node.singleProperty = true
	}
	for nodeEnd > r.ReadSoFar() {
		subNode := r.parseBinaryNode(reader, version)
		if subNode != nil {
			this.parseBinarySubNode(name, node, subNode)
		}
	}
	node.propertyList = propertyList // raw property list used by parent
	node.name = name
	// Regards the first three elements in propertyList as id, attrName, and attrType
	if len(propertyList) == 0 {
		return nil, node
	}
	if i, ok := propertyList[0].Payload().(int64); ok {
		node.id = i
	} else {
		return nil, errors.New("Expected int64 type for ID")
	}
	if len(propertyList) == 1 {
		return nil, node
	}
	if s, ok := propertyList[1].Payload().(string); ok {
		node.attrName = s
	} else {
		return nil, errors.New("Expected string type for attrName")
	}
	if len(propertyList) == 2 {
		return nil, node
	}
	if s, ok := propertyList[2].Payload().(string); ok {
		node.attrType = s
	} else {
		return nil, errors.New("Expected string type for attrType")
	}
	return nil, node
}

func (l *Loader) parseBinarySubNode(name string, node, subNode Node) {
	// special case: child node is single property
	if subNode.singleProperty {
		value := subNode.propertyList[0]
		if value.IsArray() {
			node[subNode.name] = subNode
			subNode.a = value
		} else {
			node[subNode.name] = value
		}
	} else if name == "Connections" && subNode.name == "C" {
		props := make([]interface{}, len(propertyList)-1)
		for i := 1; i < len(propertyList); i++ {
			props[i-1] = propertyList[i]
		}
		node.connections = append(node.connections, props)
	} else if subNode.name == "Properties70" {
		for k, v := range subNode {
			node[k] = v
		}
	} else if name == "Properties70" && subNode.name == "P" {
		innerPropName, ok := subNode.propertyList[0].(string)
		if !ok {
			return nil, errors.New("Expected string inner property name")
		}
		inPropType, ok := subNode.propertyList[1].(string)
		if !ok {
			return nil, errors.New("Expected string inner property type")
		}
		inPropType2 := subNode.propertyList[2]
		innerPropFlag := subNode.propertyList[3]
		var innerPropValue Property

		if strings.HasPrefix(innerPropName, "Lcl ") {
			innerPropName = "Lcl_" + innerPropName[3:]
		}
		if strings.HasPrefix(inPropType, "Lcl ") {
			inPropType = "Lcl_" + inPropType[3:]
		}

		if inPropType == "Color" || inPropType == "ColorRGB" || inPropType == "Vector" || inPropType == "Vector3D" || strings.HasPrefix(inPropType, "Lcl_") {
			innerPropValue = &ArrayProperty{
				[]interface{}{
					subNode.propertyList[4].Payload(),
					subNode.propertyList[5].Payload(),
					subNode.propertyList[6].Payload(),
				},
			}
		} else {
			innerPropValue = subNode.propertyList[4]
		}
		// this will be copied to parent, see above
		node[innerPropName] = map[string]interface{}{
			"type":  inPropType,
			"type2": inPropType2,
			"flag":  innerPropFlag,
			"value": innerPropValue,
		}
	} else if _, ok := node[subNode.name]; !ok {
		if i, ok := subNode[id].(int64); ok {
			node[subNode.name] = make(map[int64]Node)
			node[subNode.name][i] = subNode
		} else {
			node[subNode.name] = subNode
		}
	} else {
		if subNode.name == "PoseNode" {
			if !node[subNode.name].IsArray() {
				// Patrick; Ugh??
				node[subNode.name] = []interface{}{node[subNode.name]}
			}
			node[subNode.name].push(subNode)
		} else if _, ok := node[subNode.name][subNode.id]; !ok {
			node[subNode.name][subNode.id] = subNode
		}
	}
}

type Property interface {
	IsArray() bool
	Payload() interface{}
}

type SimpleProperty struct {
	payload interface{}
}

func (sp *SimpleProperty) Payload() interface{} {
	return sp.payload
}

func (sp *SimpleProperty) IsArray() bool {
	return false
}

type ArrayProperty struct {
	payload interface{}
}

func (ap *ArrayProperty) IsArray() bool {
	return true
}

func (ap *ArrayProperty) Payload() interface{} {
	return ap.payload
}

func (l *Loader) parseBinaryProperty(r *BinaryReader) interface{} {
	ty := r.getString(1)
	switch ty {
	case 'C':
		return &SimpleProperty{r.getBoolean()}, nil
	case 'D':
		return &SimpleProperty{r.getFloat64()}, nil
	case 'F':
		return &SimpleProperty{r.getFloat32()}, nil
	case 'I':
		return &SimpleProperty{r.getInt32()}, nil
	case 'L':
		return &SimpleProperty{r.getInt64()}, nil
	case 'R':
		return &ArrayProperty{r.getArrayBuffer(r.getUint32())}, nil
	case 'S':
		return &SimpleProperty{r.getString(r.getUint32())}, nil
	case 'Y':
		return &SimpleProperty{r.getInt16()}, nil
	case 'b':
	case 'c':
	case 'd':
	case 'f':
	case 'i':
	case 'l':
		arrayLength := r.getUint32()
		encoding := r.getUint32() // 0: non-compressed, 1: compressed
		compressedLength := r.getUint32()
		if encoding == 0 {
			switch ty {
			case 'b':
			case 'c':
				return &ArrayProperty{r.getBooleanArray(arrayLength)}, nil
			case 'd':
				return &ArrayProperty{r.getFloat64Array(arrayLength)}, nil
			case 'f':
				return &ArrayProperty{r.getFloat32Array(arrayLength)}, nil
			case 'i':
				return &ArrayProperty{r.getInt32Array(arrayLength)}, nil
			case 'l':
				return &ArrayProperty{r.getInt64Array(arrayLength)}, nil
			}
		}
		r2 := zlib.NewReader(r.getArrayBuffer(compressedLength))
		defer r2.Close()
		switch ty {
		case 'b':
		case 'c':
			return &ArrayProperty{r2.getBooleanArray(arrayLength)}, nil
		case 'd':
			return &ArrayProperty{r2.getFloat64Array(arrayLength)}, nil
		case 'f':
			return &ArrayProperty{r2.getFloat32Array(arrayLength)}, nil
		case 'i':
			return &ArrayProperty{r2.getInt32Array(arrayLength)}, nil
		case 'l':
			return &ArrayProperty{r2.getInt64Array(arrayLength)}, nil
		}
	}
	return nil, errors.New("Undefined property type: " + ty)
}

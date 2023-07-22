package threefbx

import (
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

//enforce id is always int

//TODO: currently working with NODE as a fbx tree node we should see if that can be merged with our object type from the previous parser
// using newTreeNode(string)

func (l *Loader) ParseASCII(r io.Reader) (Tree, error) {
	tp := NewTextParser()
	byt, err := ioutil.ReadAll(r)
	if err != nil {
		return Tree{}, err
	}
	return tp.parse(string(byt)), nil
}

type TextParser struct {
	nodeStack     []Node
	currentProp   Property
	allNodes      Tree
	currentIndent int
}

var (
	firstQuote          = regexp.MustCompile("^\"")
	encapsulatingQuotes = regexp.MustCompile("^\"|\"$")
	quotes              = regexp.MustCompile("\"")
	lastComma           = regexp.MustCompile(",$")
	whiteSpace          = regexp.MustCompile("/s")
	split               = regexp.MustCompile("[\r\n]+")
	commentRegex        = regexp.MustCompile("^[\\s\\t]*;'")
	emptyRegex          = regexp.MustCompile("^[\\s\\t]*$")
)

//CurrentIndent iswhere the next thing is

func NewTextParser() *TextParser {
	return &TextParser{nodeStack: []Node{}}
}

// parse takes in ascii formatted text and parses it into a node structure for the FBX tree
func (tp *TextParser) parse(text string) Tree {
	tp.allNodes = Tree{}
	splitText := split.Split(text, -1)

	for lineNum, line := range splitText {
		if commentRegex.FindStringIndex(line) != nil || emptyRegex.FindStringIndex(line) != nil {
			break
		}
		if vals := regexp.MustCompile("^\\t{"+strconv.Itoa(tp.currentIndent)+"}(\\w+):(.*){").FindAllString(line, -1); len(vals) > 0 {
			tp.parseNodeBegin(line, vals)
		} else if vals := regexp.MustCompile("^\\t{"+strconv.Itoa(tp.currentIndent)+"}(\\w+):[\\s\\t\\r\\n](.*)").FindAllString(line, -1); len(vals) > 0 {
			tp.parseNodeProperty(line, vals, splitText[lineNum+1])
		} else if vals := regexp.MustCompile("^\\t{"+strconv.Itoa(tp.currentIndent-1)+"}}").FindAllString(line, -1); len(vals) > 0 {
			//nodestack pop.
			tp.nodeStack = tp.nodeStack[0 : len(tp.nodeStack)-1]
			tp.currentIndent--
			// tp.popStack()
		} else if matched, _ := regexp.MatchString("^[^\\s\\t}]", line); matched {
			tp.parseNodePropertyContinued(line)
		}
	}
	return tp.allNodes
}

// unwrapProperty is a helper due to its common use in the textparser on properties
func unwrapProperty(toUnwrap string) string {
	return encapsulatingQuotes.ReplaceAllString(strings.TrimSpace(toUnwrap), "")
}

type nodeAttr struct {
	ID   IDType
	Name string
	Typ  string
}

//TODO what is a node attr
//REMINDER: bring up at meeting: original id format doesnt make sense  assign and reassign if not int unless rollback?
func (tp *TextParser) parseNodeAttr(attrs []string) (nodeAttr, error) {
	id := attrs[0]
	name := ""
	typ := ""
	if len(attrs) > 2 {
		name = attrs[1]
		typ = attrs[2]
	} else if len(attrs) == 1 {
		fmt.Println("Unexpected value according to original as it checked length was greater than 1 but then also hit the index 2")
	}

	return nodeAttr{
		ID:   id,
		Name: name,
		Typ:  typ,
	}, nil
}

// getCurrentNode returns the current node given the nodestack on tp
func (tp *TextParser) getCurrentNode() Node {
	return tp.nodeStack[tp.currentIndent-1]
}

func (tp *TextParser) addNode(nodeName string, attrID IDType, node *Node) {
	if _, ok := tp.allNodes.Objects[nodeName]; !ok {
		tp.allNodes.Objects[nodeName] = make(map[IDType]*Node)
	}

	tp.allNodes.Objects[nodeName][attrID] = node
}

func (tp *TextParser) parseNodeBegin(line string, property []string) error {
	nodeName := unwrapProperty(property[1])
	nodeAttrs := strings.Split(property[2], ",")
	for i := 0; i < len(nodeAttrs); i++ {
		nodeAttrs[i] = unwrapProperty(nodeAttrs[i])
	}

	node := NewNode(nodeName)
	//TODO: Remove need for these

	//attrs can return without an integer id... when?
	attrs, err := tp.parseNodeAttr(nodeAttrs)
	if err != nil {
		return err
	}
	currentNode := tp.getCurrentNode()

	if tp.currentIndent == 0 {
		tp.addNode(nodeName, attrs.ID, node)
		// tp.allNodes.Objects[nodeName][attrs.ID] = node
		// tp.allNodes.append(nodeName, node) //this adds to the overall nodes that dont exist yet
	} else {
		//This is a subnode
		eProp, ok := currentNode.props[nodeName]
		if ok {
			if nodeName == "PoseNode" {
				poseNodes := currentNode.props["PoseNode"].Payload.([]*Node)
				poseNodes = append(poseNodes, node)
				currentNode.props["PoseNode"] = Property{Payload: poseNodes}
			} else if _, ok := currentNode.props[nodeName]; ok {
				// currentNode.props[nodeName]
				//TODO: Figure this out looks like it may be the mechanism they used to go from single prop to multi?
				// currentNode[ nodeName ] = {};
				//         currentNode[ nodeName ][ currentNode[ nodeName ].id ] = currentNode[ nodeName ];
				currentNode.props[nodeName] = Property{Payload: map[IDType]Property{attrs.ID: eProp}}
			}
			if attrs.ID != "" {
				m, ok := eProp.Payload.(map[IDType]Property)
				if !ok {
					return nil
				}
				_, ok = m[attrs.ID]
				if ok {
					return nil
				}
				m[attrs.ID] = NodeProperty(node)
				currentNode.props[nodeName] = Property{Payload: m}
			}
		} else if nodeName != "Properties70" {
			if nodeName == "PoseNode" {
				currentNode.props[nodeName] = Property{Payload: []*Node{node}}
			} else {
				currentNode.props[nodeName] = Property{Payload: map[IDType]Property{attrs.ID: NodeProperty(node)}}
			}
		} else {
			currentNode.props[nodeName] = Property{Payload: map[IDType]Property{attrs.ID: currentNode.props[nodeName]}}
		}
		node.ID = attrs.ID

		if attrs.Name != "" {
			node.attrName = attrs.Name
		}
		if attrs.Typ != "" {
			node.attrType = attrs.Typ
		}
		tp.nodeStack = append(tp.nodeStack, *node)
	}
	return nil
}

//parseNodeProperty takes the current line a prop and the next line
func (tp *TextParser) parseNodeProperty(line string, property []string, contentLine string) {

	var propValue interface{}

	propName := unwrapProperty(property[1])
	propValue = unwrapProperty(property[2])
	// for special case: base64 image data follows "Content: ," line
	//	Content: ,
	//	 "/9j/4RDaRXhpZgAA  TU0A..."
	if propName == "Content" && propValue == "," {
		propValue = lastComma.ReplaceAllString(quotes.ReplaceAllString(contentLine, ""), "")
		propValue = strings.TrimSpace(propValue.(string))
	}
	currentNode := tp.getCurrentNode()
	parentName := currentNode.name
	if parentName == "Properties70" {
		tp.parseNodeSpecialProperty(line, propName, propValue.(string))
		return
	}
	_, nodeHasProp := currentNode.props[propName]
	// Connections
	if propName == "C" {
		con := Connection{}
		connProps := strings.Split(propValue.(string), ",")[1:]
		//from, err1 := strconv.Atoi(connProps[0])
		//to, err2 := strconv.Atoi(connProps[1])
		//if err1 != nil || err2 != nil {
		//	return
		//}
		con.From = connProps[0]
		con.To = connProps[1]

		rest := strings.Split(propValue.(string), ",")[3:]
		for i := 0; i < len(rest); i++ {
			rest[i] = strings.TrimSpace(firstQuote.ReplaceAllString(rest[i], ""))
		}
		con.Property = strings.Join(rest, ",")
		propName = "connections"
	} else if propName == "Node" {
		//id, err := strconv.Atoi(propValue.(string))
		//	if err != nil {
		//	return
		//}
		currentNode.ID = propValue.(string)
	}
	if nodeHasProp && currentNode.props[propName].IsArray() {
		pl := currentNode.props[propName].Payload.([]*Node)
		currentNode.props[propName] = Property{Payload: append(pl, propValue.(*Node))}
	} else {
		if propName != "a" {
			currentNode.props[propName] = Property{Payload: propValue}
		} else {
			// currentNode.a = propValue
			currentNode.props[propName] = Property{Payload: propValue}
		}
	}

	ps := propValue.(string)

	// tp.setCurrentProp(currentNode, propName)
	// convert string to array, unless it ends in "," in which case more will be added to it
	if propName == "a" && ps[len(ps)-1] != ',' {
		currentNode.props[propName] = tp.parseNumberArray(propValue.(string))
	}
}

func (tp *TextParser) parseNodeSpecialProperty(line string, propName string, propValue string) {

	props := strings.Split(propValue, "\",")
	for i := 0; i < len(props); i++ {
		props[i] = whiteSpace.ReplaceAllString(firstQuote.ReplaceAllString(strings.TrimSpace(props[i]), ""), "_")
	}
	innerPropName := props[0]
	innerPropType1 := props[1]
	innerPropType2 := props[2]
	innerPropFlag := props[3]
	innerPropValue := props[4]

	var parsedValue interface{}

	switch innerPropType1 {
	case "int", "enum", "bool", "ULongLong",
		"double", "Number", "FieldOfView":
		var err error
		parsedValue, err = strconv.ParseFloat(innerPropValue, 64)
		if err != nil {
			fmt.Println("Error parsing property as float", err)
		}
	case "Color", "ColorRGB", "Vector3D", "Lcl_Translation",
		"Lcl_Rotation", "Lcl_Scaling":
		parsedValue = tp.parseNumberArray(innerPropValue)
	default:
		parsedValue = innerPropValue
	}
	tp.nodeStack[tp.currentIndent-2].props[innerPropName] = Property{Payload: map[string]Property{
		"type":  Property{Payload: innerPropType1},
		"type2": Property{Payload: innerPropType2},
		"flag":  Property{Payload: innerPropFlag},
		"value": Property{Payload: parsedValue},
	}}
}

func (tp *TextParser) parseNumberArray(s string) Property {
	ar := strings.Split(s, ",")[3:]
	floatArr := make([]float64, len(ar))
	var err error
	for i := 0; i < len(ar); i++ {
		floatArr[i], err = strconv.ParseFloat(
			strings.TrimSpace(firstQuote.ReplaceAllString(ar[i], "")),
			64,
		)
		if err != nil {
			fmt.Println("Error parsing float", ar[i])
			return Property{}
		}
	}
	return Property{Payload: floatArr}
}

// parseNodePropertyContinued appends lines to the property on .a until it is finished and then parses itas a number array
func (tp *TextParser) parseNodePropertyContinued(line string) {
	currentNode := tp.getCurrentNode()
	if line[len(line)-1] != ',' {
		new := tp.parseNumberArray(line)
		old := currentNode.a.Payload.([]float64)
		currentNode.a = Property{Payload: append(old, new.Payload.([]float64)...)}
	} else {
		old := currentNode.a.Payload.([]string)
		currentNode.a = Property{Payload: append(old, line)}
	}
}
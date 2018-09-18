package threefbx

import (
	"strconv"
	"strings"
)

//enforce id is always int

//TODO: currently working with NODE as a fbx tree node we should see if that can be merged with our object type from the previous parser
// using newTreeNode(string)


type TextParser struct{
	nodeStack []Node
	currentProp Property
	allNodes FBXTree
	currentIndent int
}

//CurrentIndent iswhere the next thing is

func NewTextParser() TextParser{
	
	tp := TextParser{nodeStack:[]Node{},currentProp}
    
    firstQuote := regexp.MustCompile("^\"")
    encapsulatingQuotes := regexp.MustCompile("^\"|\"$")
    quotes := regexp.MustCompile("\"")
    lastComma := regexp.MustCompile(",$")
    whiteSpace := regexp.MustCompile("/s")
    

}

// parse takes in ascii formatted text and parses it into a node structure for the FBX tree
func (tp *TextParser) parse(text String){
	//TODO: how does this know bout the fbxTree
	fmt.Println("FBXTree: ", FBxTree)
	tp.allNodes = NewFBXTree()
	split := regexp.MustCompile("[\r\n]+").Split(text,-1)

	commentRegex := regexp.MustCompile("^[\\s\\t]*;'")
	emptyRegex := regexp.MustCompile("^[\\s\\t]*$")


	for lineNum, line := range split{
		if commentRegex.FindStringIndex(split)!=nil || emptyRegex.FindStringIndex(split) !=nil{
			return
		}
		if val, ok := regexp.MustCompile("^\\t{" + strconv.Atoi(tp.currentIndent) + "}(\\w+):(.*){").FindAllString(line, -1); ok{
			tp.parseNodeBegin(line, val)
		}else if val, ok := regexp.MustCompile("^\\t{" + strconv.Atoi( self.currentIndent ) + "}(\\w+):[\\s\\t\\r\\n](.*)").FindAllString(line, -1); ok {
			tp.parseNodeProperty(line, val, split[lineNum+1])
		}else if val, ok := regexp.MustCompile("^\\t{" + strconv.Atoi( self.currentIndent - 1 ) + "}}").FindAllString(line, -1); ok {
			tp.popStack()
		}else if regexp.MatchString("^[^\\s\\t}]"){
            tp.parseNodePropertyContinued(line)
        }
	} 
	return tp.allNodes
}

// unwrapProperty is a helper due to its common use in the textparser on properties
func unwrapProperty(toUnwrap string) string{
    return  encapsulatingQuotes.ReplaceAllString(strings.Trim(toUnwrap,""))
}

type nodeAttr struct {
    ID int64
    Name string
    Typ string
}

//TODO what is a node attr
//REMINDER: bring up at meeting: original id format doesnt make sense  assign and reassign if not int unless rollback?
func (tp *TextParser) parseNodeAttr(attrs []string) (nodeAttr, error) {
    id, err := strconv.ParseInt(attrs[0], 10, 64)
    if err != nil {
        return nodeAttr{}, err
    }
    name := ""
    typ := ""
    if len(attrs) > 2{
        name = attrs[1]
        typ = attrs[2]
    } else if len(attrs) == 1 {
        fmt.Println("Unexpected value according to original as it checked length was greater than 1 but then also hit the index 2")
    }

    return nodeAttr{ 
        ID: id,
        Name: name, 
        Typ: typ,
    }, nil
}
    
func (tp *TextParser) parseNodeBegin(line string, property []string){
    nodeName := unwrapProperty(property[1])
    nodeAttrs :=  strings.Split(property[2], ",")
    for i:=0; i < len(nodeAttrs); i++ {
       nodeAttrs[i] = unwrapProperty(nodeAttrs[i])
     }

     node := newTreeNode(nodeName)
     //TODO: Remove need for these

     //attrs can return without an integer id... when?
     attrs := tp.parseNodeAttr(nodeAttrs)
     currentNode := tp.getCurrentNode()

    atrId, attrErr := strconv.Atoi(attrs.id)

     if(tp.currentIndent==0){
         tp.allNodes.add(nodeName,node) //this adds to the overall nodes that dont exist yet
     }else {
          //This is a subnode
          eProp, ok = currentNode.props[nodeName];
          if  ok {
            if nodeName =="PoseNode"{
                currentNode.PoseNodes = append(currentNode.poseNodes, node)
            }else if currentNode.props[nodeName] {
                // currentNode.props[nodeName]
                //TODO: Figure this out looks like it may be the mechanism they used to go from single prop to multi?
                // currentNode[ nodeName ] = {};
                //         currentNode[ nodeName ][ currentNode[ nodeName ].id ] = currentNode[ nodeName ];
            }
            if  atrId != 0 {  
                currentNode.props[ nodeName ][ attrs.id ] = node
            }
            
          }  else if !attrErr  {
                currentNode.props[ nodeName ] = []property{}
                currentNode.props[ nodeName ][ attrs.id ] = node;
        } else if nodeName != "Properties70" {
                if  nodeName == "PoseNode"	{
                    currentNode.Props[ nodeName ] = ArrayProperty{[]Node{node}}
                } else{
                    currentNode.Props[ nodeName ] = SimpleProperty{node}
                }
        }
        if !attrErr{
            node.ID = atrId
        }
        if attrs.Name != ""{
            node.attrName = attrs.name
        }
        if attrs.Typ != "" {
            node.attrType = attrs.Typ
        }
        tp.nodeStack = append(tp.nodeStack, node)
    }
}

//parseNodeProperty takes the current line a prop and the next line
func (tp *TextParser) parseNodeProperty(line, property, contentLine string){
    propName := unwrapProperty(property[1])
    propValue := unwrapProperty(property[2])
        // for special case: base64 image data follows "Content: ," line
        //	Content: ,
        //	 "/9j/4RDaRXhpZgAA  TU0A..."
        if  propName == "Content" && propValue == "," {                                                                                                                                                                                                                                                                                       
            propValue = lastComma.ReplaceAllString(quotes.ReplaceAllString(contentLine,""), "")
            propValue =  strings.Trim(propValue)
        }
         currentNode := tp.getCurrentNode();
         parentName := tp.currentNode.name;
        if  parentName == "Properties70"  {
            tp.parseNodeSpecialProperty( line, propName, propValue )
        return;                                                          
        }
        _ , nodeHasProp :=  currentNode.props[ propName ] 
         // Connections
         if  propName == "C"  {
             connProps := strings.Split(propValue,",")[1:]
             from :=   Atoi(connProps[0])
             to :=  Atoi(connProps[1])

             rest := strings.Split(propValue, "," )[3:]
            for i := 0; i < len(rest) ; i++ {
                rest[i] = strings.Trim(firstQuote.ReplaceAllString(rest[i], ""))
            }
            propName = "connections"
            propValue = []int{from, to}

            propValue = append( propValue, rest );
            if !nodeHasProp {
                currentNode[ propName ] = []int{}
            }   
        }
        
        // Node
        if  propName == "Node" {
            currentNode.ID = propValue
        }
        // connections
        if nodeHasProp && currentNode.props[ propName ].IsArray()   {
            currentNode.props[ propName ] = append(currentNode.props[ propName ],propValue )
        } else {
            if propName != "a" {
                currentNode[ propName ] = propValue
            } else{
                currentNode.a = propValue
            }
        }
        tp.setCurrentProp( currentNode, propName )
        // convert string to array, unless it ends in "," in which case more will be added to it
        if  propName == "a" && propValue[- 1] != "," {
            currentNode.a = parseNumberArray( propValue )
        }
}


func (tp *TextParser) parseNodeSpecialProperty(line string, propName string, propValue string){

    props := strings.Split(propValue, "\",")
    for i := 0 ; i < len(props); i ++{
        props[i] = whiteSpace.ReplaceAllString(firstQuote.ReplaceAllString(strings.Trim(props[i]),""), "_")
        innerPropName := props[0]
        innerPropType1 := props[ 1 ]
        innerPropType2 := props[ 2 ]
        innerPropFlag := props[ 3 ]
        innerPropValue := props[ 4 ]

        switch innerPropType1{
        case "int"||"enum"|| "bool" || "ULongLong" ||
         "double" || "Number" || "FieldOfView" :
            innerPropValue = strings.ParseFloat(innerPropValue)
        case "Color" || "ColorRGB" || "Vector3D" || "Lcl_Translation" || 
            "Lcl_Rotation" || "Lcl_Scaling":
         innerPropValue = parseNumberArray( innerPropValue )
        }
    }
    tp.getPrevNode()[innerPropName] = map[string]string{
        "type": innerPropType1,
        "type2": innerPropType2,
        "flag": innerPropFlag,
        "value": innerPropValue,
    }
    tp.setCurrentProp(tp.getPrevNode(), innerPropName)
}

// parseNodePropertyContinued appends lines to the property on .a until it is finished and then parses itas a number array
func (tp *TextParser)  parseNodePropertyContinued(line string ){
    currentNode := tp.getCurrentNode()
    currentNode.a  = append(currentNode.a, line)
    if line[-1] != ","{
        currentNode.a = parseNubmerArray(currentNode.a)
    }
}


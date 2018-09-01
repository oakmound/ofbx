package threefbx

import (
	"strings"
)


//TODO: currently working with NODE as a fbx tree node we should see if that can be merged with our object type from the previous parser
// using newTreeNode(string)


type TextParser struct{
	nodeStack []Node
	currentProp ???
	allNodes FBXTree
	currentIndent int
}

//CurrentIndent iswhere the next thing is

func NewTextParser() TextParser{
	
	tp := TextParser{nodeStack:[]Node{},currentProp}
    
    
    encapsulatingQuotes := regexp.MustCompile("^\"|\"$")

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
		}else if regexp.MatchString("^[^\s\t}]"){
            tp.parseNodePropertyContinued(line)
        }
	} 
	return tp.allNodes
}

// unwrapProperty is a helper due to its common use in the textparser on properties
func unwrapProperty(toUnwrap string) string{
    return  encapsulatingQuotes.ReplaceAllString(strings.Trim(toUnwrap,""))
}

//TODO what is a node attr
//REMINDER: bring up at meeting: original id format doesnt make sense  assign and reassign if not int unless rollback?
func (tp *TextParser) parseNodeAttr(attrs []string) nodeAttr{
    id := attrs[0]
    if (id != ''){

    }
    name := ''
    typ := ''
    if len(attrs) > 2{
        name = attrs[1]
        typ = attrs[2]
    }else if len(attrs)==1{
        fmt.Println("Unexpected value according to original as it checked length was greater than 1 but then also hit the index 2")
    }

    return nodeAttr{ id: id, name: name, type: type }    
}
    
func (tp *TextParser) parseNodeBegin(line string, property []string){
    nodeName := unwrapProperty(property[1])
    nodeAttrs :=  strings.Split(property[2], ",")
     for i:=0; i < len(nodeAttrs); i++ ){
       nodeAttrs[i] = unwrapProperty(nodeAttrs[i])
     }

     node := newTreeNode(nodeName)
     //TODO: Remove need for these
     attrs := tp.parseNodeAttr(nodeAttrs)
     currentNode := tp.getCurrentNode()
     if(tp.currentIndent==0){
         tp.allNodes.add(nodeName,node)
     }else{
         if(currentNode.Nodes)
     }

-----------------------------------


}

//parseNodeProperty takes the current line a prop and the next line
func (tp *TextParser) parseNodeProperty(line string ,property string,contentLine string){
    propName := unwrapProperty(property[1])
   propValue := unwrapProperty(property[2])
        // for special case: base64 image data follows "Content: ," line
        //	Content: ,
        //	 "/9j/4RDaRXhpZgAATU0A..."
        if ( propName === 'Content' && propValue === ',' ) {
            propValue = contentLine.replace( /"/g, '' ).replace( /,$/, '' ).trim();
        }
}




// parse an FBX file in ASCII format
function TextParser() {}
TextParser.prototype = {
    constructor: TextParser,
 
    parseNodeBegin: function ( line, property ) {
        var nodeName = property[ 1 ].trim().replace( /^"/, '' ).replace( /"$/, '' );
        var nodeAttrs = property[ 2 ].split( ',' ).map( function ( attr ) {
            return attr.trim().replace( /^"/, '' ).replace( /"$/, '' );
        } );
        var node = { name: nodeName };
        var attrs = this.parseNodeAttr( nodeAttrs );
        var currentNode = this.getCurrentNode();
        // a top node
        if ( this.currentIndent === 0 ) {
            this.allNodes.add( nodeName, node );
        } else { // a subnode
            // if the subnode already exists, append it
            if ( nodeName in currentNode ) {
            // special case Pose needs PoseNodes as an array
                if ( nodeName === 'PoseNode' ) {
                    currentNode.PoseNode.push( node );
                } else if ( currentNode[ nodeName ].id !== undefined ) {
                    currentNode[ nodeName ] = {};
                    currentNode[ nodeName ][ currentNode[ nodeName ].id ] = currentNode[ nodeName ];
                }
                if ( attrs.id !== '' ) currentNode[ nodeName ][ attrs.id ] = node;
            } else if ( typeof attrs.id === 'number' ) {
                currentNode[ nodeName ] = {};
                currentNode[ nodeName ][ attrs.id ] = node;
            } else if ( nodeName !== 'Properties70' ) {
                if ( nodeName === 'PoseNode' )	currentNode[ nodeName ] = [ node ];
                else currentNode[ nodeName ] = node;
            }
        }
        if ( typeof attrs.id === 'number' ) node.id = attrs.id;
        if ( attrs.name !== '' ) node.attrName = attrs.name;
        if ( attrs.type !== '' ) node.attrType = attrs.type;
        this.pushStack( node );
    },
    // parseNodeAttr: function ( attrs ) {
    //     var id = attrs[ 0 ];
    //     if ( attrs[ 0 ] !== '' ) {
    //         id = parseInt( attrs[ 0 ] );
    //         if ( isNaN( id ) ) {
    //             id = attrs[ 0 ];
    //         }
    //     }
    //     var name = '', type = '';
    //     if ( attrs.length > 1 ) {
    //         name = attrs[ 1 ].replace( /^(\w+)::/, '' );
    //         type = attrs[ 2 ];
    //     }
    //     return { id: id, name: name, type: type };
    // },
    parseNodeProperty: function ( line, property, contentLine ) {
        var propName = property[ 1 ].replace( /^"/, '' ).replace( /"$/, '' ).trim();
        var propValue = property[ 2 ].replace( /^"/, '' ).replace( /"$/, '' ).trim();
        // for special case: base64 image data follows "Content: ," line
        //	Content: ,
        //	 "/9j/4RDaRXhpZgAATU0A..."
        if ( propName === 'Content' && propValue === ',' ) {
            propValue = contentLine.replace( /"/g, '' ).replace( /,$/, '' ).trim();
        }
        var currentNode = this.getCurrentNode();
        var parentName = currentNode.name;
        if ( parentName === 'Properties70' ) {
            this.parseNodeSpecialProperty( line, propName, propValue );
            return;
        }
        // Connections
        if ( propName === 'C' ) {
            var connProps = propValue.split( ',' ).slice( 1 );
            var from = parseInt( connProps[ 0 ] );
            var to = parseInt( connProps[ 1 ] );
            var rest = propValue.split( ',' ).slice( 3 );
            rest = rest.map( function ( elem ) {
                return elem.trim().replace( /^"/, '' );
            } );
            propName = 'connections';
            propValue = [ from, to ];
            append( propValue, rest );
            if ( currentNode[ propName ] === undefined ) {
                currentNode[ propName ] = [];
            }
        }
        // Node
        if ( propName === 'Node' ) currentNode.id = propValue;
        // connections
        if ( propName in currentNode && Array.isArray( currentNode[ propName ] ) ) {
            currentNode[ propName ].push( propValue );
        } else {
            if ( propName !== 'a' ) currentNode[ propName ] = propValue;
            else currentNode.a = propValue;
        }
        this.setCurrentProp( currentNode, propName );
        // convert string to array, unless it ends in ',' in which case more will be added to it
        if ( propName === 'a' && propValue.slice( - 1 ) !== ',' ) {
            currentNode.a = parseNumberArray( propValue );
        }
    },
    parseNodePropertyContinued: function ( line ) {
        var currentNode = this.getCurrentNode();
        currentNode.a += line;
        // if the line doesn't end in ',' we have reached the end of the property value
        // so convert the string to an array
        if ( line.slice( - 1 ) !== ',' ) {
            currentNode.a = parseNumberArray( currentNode.a );
        }
    },
    // parse "Property70"
    parseNodeSpecialProperty: function ( line, propName, propValue ) {
        // split this
        // P: "Lcl Scaling", "Lcl Scaling", "", "A",1,1,1
        // into array like below
        // ["Lcl Scaling", "Lcl Scaling", "", "A", "1,1,1" ]
        var props = propValue.split( '",' ).map( function ( prop ) {
            return prop.trim().replace( /^\"/, '' ).replace( /\s/, '_' );
        } );
        var innerPropName = props[ 0 ];
        var innerPropType1 = props[ 1 ];
        var innerPropType2 = props[ 2 ];
        var innerPropFlag = props[ 3 ];
        var innerPropValue = props[ 4 ];
        // cast values where needed, otherwise leave as strings
        switch ( innerPropType1 ) {
            case 'int':
            case 'enum':
            case 'bool':
            case 'ULongLong':
            case 'double':
            case 'Number':
            case 'FieldOfView':
                innerPropValue = parseFloat( innerPropValue );
                break;
            case 'Color':
            case 'ColorRGB':
            case 'Vector3D':
            case 'Lcl_Translation':
            case 'Lcl_Rotation':
            case 'Lcl_Scaling':
                innerPropValue = parseNumberArray( innerPropValue );
                break;
        }
        // CAUTION: these props must append to parent's parent
        this.getPrevNode()[ innerPropName ] = {
            'type': innerPropType1,
            'type2': innerPropType2,
            'flag': innerPropFlag,
            'value': innerPropValue
        };
        this.setCurrentProp( this.getPrevNode(), innerPropName );
    },
};
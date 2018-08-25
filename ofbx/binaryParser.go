package threefbx



func (l *Loader) parseBinary(r io.Reader) (FBXTree, error) {
	reader := NewBinaryReader(r, true)
	// We already read first 21 bytes
	reader.Discard(2); // skip reserved bytes

	var version = reader.getUint32();
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
func (l *Loader) parseBinaryNode(r *BinaryReader, version int) {
	var n Node
	// The first three data sizes depends on version.
	var endOffset uint64
	var numProperties uint64
	var propertiesListLen uint64
	if version >= 7500 { 
		endOffset = r.getUint64()
		numProperties = r.getUint64()
		propertiesListLen = r.getUint64()
	} else {
		endOffset = uint64(r.getUint32())
		numProperties = uint64(r.getUint32())
		propertiesListLen = uint64(r.getUint32())
	}
	name := r.getShortString(nameLen)
	// Regards this node as NULL-record if endOffset is zero
	if ( endOffset === 0 ) return null;
	var propertyList = [];
	for ( var i = 0; i < numProperties; i ++ ) {
		propertyList.push( this.parseProperty( reader ) );
	}
	// Regards the first three elements in propertyList as id, attrName, and attrType
	var id = propertyList.length > 0 ? propertyList[ 0 ] : '';
	var attrName = propertyList.length > 1 ? propertyList[ 1 ] : '';
	var attrType = propertyList.length > 2 ? propertyList[ 2 ] : '';
	// check if this node represents just a single property
	// like (name, 0) set or (name2, [0, 1, 2]) set of {name: 0, name2: [0, 1, 2]}
	node.singleProperty = ( numProperties === 1 && reader.getOffset() === endOffset ) ? true : false;
	while ( endOffset > reader.getOffset() ) {
		var subNode = this.parseBinaryNode( reader, version );
		if ( subNode !== null ) this.parseBinarySubNode( name, node, subNode );
	}
	node.propertyList = propertyList; // raw property list used by parent
	if ( typeof id === 'number' ) node.id = id;
	if ( attrName !== '' ) node.attrName = attrName;
	if ( attrType !== '' ) node.attrType = attrType;
	if ( name !== '' ) node.name = name;
	return node;
}
		
func (l *Loader) parseBinarySubNode(name string, node, subNode Node) {
	// special case: child node is single property
	if ( subNode.singleProperty === true ) {
		var value = subNode.propertyList[ 0 ];
		if ( Array.isArray( value ) ) {
			node[ subNode.name ] = subNode;
			subNode.a = value;
		} else {
			node[ subNode.name ] = value;
		}
	} else if ( name === 'Connections' && subNode.name === 'C' ) {
		var array = [];
		subNode.propertyList.forEach( function ( property, i ) {
			// first Connection is FBX type (OO, OP, etc.). We'll discard these
			if ( i !== 0 ) array.push( property );
		} );
		if ( node.connections === undefined ) {
			node.connections = [];
		}
		node.connections.push( array );
	} else if ( subNode.name === 'Properties70' ) {
		var keys = Object.keys( subNode );
		keys.forEach( function ( key ) {
			node[ key ] = subNode[ key ];
		} );
	} else if ( name === 'Properties70' && subNode.name === 'P' ) {
		var innerPropName = subNode.propertyList[ 0 ];
		var innerPropType1 = subNode.propertyList[ 1 ];
		var innerPropType2 = subNode.propertyList[ 2 ];
		var innerPropFlag = subNode.propertyList[ 3 ];
		var innerPropValue;
		if ( innerPropName.indexOf( 'Lcl ' ) === 0 ) innerPropName = innerPropName.replace( 'Lcl ', 'Lcl_' );
		if ( innerPropType1.indexOf( 'Lcl ' ) === 0 ) innerPropType1 = innerPropType1.replace( 'Lcl ', 'Lcl_' );
		if ( innerPropType1 === 'Color' || innerPropType1 === 'ColorRGB' || innerPropType1 === 'Vector' || innerPropType1 === 'Vector3D' || innerPropType1.indexOf( 'Lcl_' ) === 0 ) {
			innerPropValue = [
				subNode.propertyList[ 4 ],
				subNode.propertyList[ 5 ],
				subNode.propertyList[ 6 ]
			];
		} else {
			innerPropValue = subNode.propertyList[ 4 ];
		}
		// this will be copied to parent, see above
		node[ innerPropName ] = {
			'type': innerPropType1,
			'type2': innerPropType2,
			'flag': innerPropFlag,
			'value': innerPropValue
		};
	} else if ( node[ subNode.name ] === undefined ) {
		if ( typeof subNode.id === 'number' ) {
			node[ subNode.name ] = {};
			node[ subNode.name ][ subNode.id ] = subNode;
		} else {
			node[ subNode.name ] = subNode;
		}
	} else {
		if ( subNode.name === 'PoseNode' ) {
			if ( ! Array.isArray( node[ subNode.name ] ) ) {
				node[ subNode.name ] = [ node[ subNode.name ] ];
			}
			node[ subNode.name ].push( subNode );
		} else if ( node[ subNode.name ][ subNode.id ] === undefined ) {
			node[ subNode.name ][ subNode.id ] = subNode;
		}
	}
}
		

func (l *Loader) parseBinaryProperty(r *BinaryReader) {
	var type = reader.getString( 1 );
	switch ( type ) {
		case 'C':
			return reader.getBoolean();
		case 'D':
			return reader.getFloat64();
		case 'F':
			return reader.getFloat32();
		case 'I':
			return reader.getInt32();
		case 'L':
			return reader.getInt64();
		case 'R':
			var length = reader.getUint32();
			return reader.getArrayBuffer( length );
		case 'S':
			var length = reader.getUint32();
			return reader.getString( length );
		case 'Y':
			return reader.getInt16();
		case 'b':
		case 'c':
		case 'd':
		case 'f':
		case 'i':
		case 'l':
			var arrayLength = reader.getUint32();
			var encoding = reader.getUint32(); // 0: non-compressed, 1: compressed
			var compressedLength = reader.getUint32();
			if ( encoding === 0 ) {
				switch ( type ) {
					case 'b':
					case 'c':
						return reader.getBooleanArray( arrayLength );
					case 'd':
						return reader.getFloat64Array( arrayLength );
					case 'f':
						return reader.getFloat32Array( arrayLength );
					case 'i':
						return reader.getInt32Array( arrayLength );
					case 'l':
						return reader.getInt64Array( arrayLength );
				}
			}
			if ( typeof Zlib === 'undefined' ) {
				console.error( 'THREE.FBXLoader: External library Inflate.min.js required, obtain or import from https://github.com/imaya/zlib.js' );
			}
			var inflate = new Zlib.Inflate( new Uint8Array( reader.getArrayBuffer( compressedLength ) ) ); // eslint-disable-line no-undef
			var reader2 = new BinaryReader( inflate.decompress().buffer );
			switch ( type ) {
				case 'b':
				case 'c':
					return reader2.getBooleanArray( arrayLength );
				case 'd':
					return reader2.getFloat64Array( arrayLength );
				case 'f':
					return reader2.getFloat32Array( arrayLength );
				case 'i':
					return reader2.getInt32Array( arrayLength );
				case 'l':
					return reader2.getInt64Array( arrayLength );
			}
		default:
			throw new Error( 'THREE.FBXLoader: Unknown property type ' + type );
	}
}
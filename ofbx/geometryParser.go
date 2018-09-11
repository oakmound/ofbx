package threefbx


//Notes:
// 		* geometries with mutliple models that have different transforms may break this.


// TODO: check assumptions
//Assumptions: parsing a geo probably returns a geometry.


// Geometry tries to replace the need for THREE.BufferGeometry
type Geometry struct{
	name string
	position []Vector3
	color []Color

	skinIndex THREE.Uint16BufferAttribute(4)
	skinWeight THREE.Float32BufferAttribute(4)
	FBX_Deformer ???

	normal ??
	uvs []UV

	groups []Group
}

type UVRaw [2]float32

type Group [3]int

type WeightEntry{ //TODO: see if we can find a better way that doesnt use this (we could do array or split weightable itself out to two props.)
	ID int
	Weight float64
}


// AddGroup was a THREE.js thing start of conversion is here it seems to store a range for which a value is the same
func (g *Geometry) AddGroup(rangeStart, count, groupValue int){
	if g.groups ==null{
		g.groups = {}
	}
	g.groups = append(g.groups, Group{rangeStart, count, groupValue})
}




// parseGeometry converted from parse in the geometry section of the original code
// parse Geometry data from FBXTree and return map of BufferGeometries
// Parse nodes in FBXTree.Objects.Geometry
func (l *Loader) parseGeometry(skeletons map[int64]Skeleton,morphTargets map[int64]MorphTarget ) Geometry, Error{
	geometryMap := make(map[???]???)
	if geoNodes, ok := l.tree.Objects["Geometry"]; ok{
		for _, nodeID := range geoNodes{
			relationships := l.connections.get(nodeID)
			geo := l.parseGeometrySingle(relationships, geoNodes[nodeID], skeletons, morphTargets)
		}
	}
	return geometryMap
}
// parseGeometrySingle parses a single node in FBXTree.Objects.Geometry //updated name due to collisions
func (l *Loader) parseGeometrySingle(relationships ConnectionSet, geoNode Node, skeletons map[int64]Skeleton,morphTargets map[int64]MorphTarget ) Geometry{
	switch geoNode.attrType{
	case "Mesh":
		return l.parseMeshGeometry(relationships, geoNode, skeletons, morphTargets)
	case "NurbsCurve":
		return l.parseNurbsGeometry(geoNode)
	}
	errors.New("Unknown geometry type when parsing" + geoNode.attrType)
	return 
}

// parseMeshGeometry parses a single node mesh geometry in FBXTree.Objects.Geometry
func (l *Loader) parseMeshGeometry( relationships ConnectionSet, geoNode Node,  skeletons map[int64]Skeleton,morphTargets map[int64]MorphTarget) Geometry {

	modelNodes := make([]Node,len(relationships.parents))
	for i, parents := range relationships.parents{
		modelNodes[i] = l.tree.Objects.Model[parent.ID]
	}

	// dont create if geometry has no models
	if len(modelNodes) ==0{
		return nil
	}

	skeleton := {}Skeleton	//TODO: whats this type
	for i := len(relationships.children)-1; i >= 0; i--{
		chID := relationships.children[i].ID
		if skel, ok := skeletons[chID] ; ok{
			skeleton = skel
			break
		}
	}
	morphTarget := {}MorphTarget //TODO: whats this type
	for i := len(relationships.children)-1; i >= 0; i--{
		chID := relationships.children[i].ID
		if morp, ok := morphTargets[chID] ; ok{
			morphTarget = morp
			break
		}
	}
	// TODO: if there is more than one model associated with the geometry, AND the models have
	// different geometric transforms, then this will cause problems
	// if ( modelNodes.length > 1 ) { }
	// For now just assume one model and get the preRotations from that
	modelNode := modelNodes[0]
	transformData := newTransformData() //TODO: figure out type and if this should just be a transform
	if val, ok := modelNode.props["RotationOrder"]; ok{
		transformData.eulerOrder =val
	}
	if val, ok := modelNode.props["GeometricTranslation"]; ok{
		transformData.translation = val
	}
	if val, ok := modelNode.props["GeometricRotation"]; ok{
		transformData.rotation =val
	}
	if val, ok := modelNode.props["GeometricScaling"]; ok{
		transformData.scale = val
	}
	transform = generateTransform( transformData ) //TODO: see above about how this ordering might change
	return l.genGeometry( geoNode, skeleton, morphTarget, transform )

}

// genGeometry generates a THREE.BufferGeometry(ish) from a node in FBXTree.Objects.Geometry
func (l *Loader) genGeometry ( geoNode Node,  skeletons map[int64]Skeleton,morphTargets map[int64]MorphTarget, preTransform ) {
	geo := Geometry{} //https://threejs.org/docs/#api/en/core/BufferGeometry
	geo.name = geoNode.attrName

	geoInfo := l.parseGeoNode(geoNode, skeleton)

	//TODO: unroll buffers into its consituent slices and do away with the buffer construct
	buffers := l.genBuffers(geoInfo)

	positionAttribute :=floatsToVertex3s(buffers.vertex) //https://threejs.org/docs/#api/en/core/BufferAttribute
	


	preTransform.applyToBufferAttribute( positionAttribute )

	geo.position =  positionAttribute
	if len(buffers.color)> 0 {
		geo.color = floatsToVertex3s(buffers.color)
	}


	if skeleton{
		geos.skinIndex =  new THREE.Uint16BufferAttribute( buffers.weightsIndices, 4 ) )
		geo.skinWeight = new THREE.Float32BufferAttribute( buffers.vertexWeights, 4 ) )
		geo.FBX_Deformer = skeleton;
	}

	if len(buffers.normal) > 0{
		normalAttribute := floatsToVertex3s(buffers.normal)
		normalMatrix := new THREE.Matrix3().getNormalMatrix( preTransform ) //TODO: convert out https://threejs.org/docs/#api/en/math/Matrix3.getNormalMatrix
		normalMatrix.applyToBufferAttribute(normalAttribute)
		geo.normal = normalAttribute
	}

	geo.uvs = buffer.uvs //NOTE: pulled back from variadic array of uvs where they progress down uv -> uv1 -> uv2 and so on

	if geoInfo.material && geoInfo.material.mappingType != "AllSame"{
		// Convert the material indices of each vertex into rendering groups on the geometry.
		prevMaterialIndex := buffers.materialIndex[ 0 ]
		startIndex := 0
		for i, matIndex := range buffers.materialIndex{
			if matIndex != prevMaterialIndex{
				geo.AddGroup(startIndex, i-startIndex, prevMaterialIndex) 
				prevMaterialIndex = materialIndex
				startIndex = i
			}
		}
		if len( geo.groups > 0){ //add last group
			lastGroup := geo.groups[ len(geo.groups) - 1 ]
			lastIndex := lastGroup[0] + lastGroup.[1]
			if  lastIndex !== len(buffers.materialIndex) {
				geo.addGroup( lastIndex, len(buffers.materialIndex) - lastIndex, prevMaterialIndex )
			}
		}
		if len(geo.groups) == 0  {
			geo.addGroup( 0, len(buffers.materialIndex), buffers.materialIndex[ 0 ] )
		}
	}

	l.addMorphTargets( geo, geoNode, morphTarget, preTransform )
	return geo
}

// floatsToVertex3s is a helper function to convert flat float arrays into vertex3s 
func floatsToVertex3s(arr []Float32) []Vertex3{
	if len(arr) %3 != 0{
		errors.New("Something went wrong parsing an array of floats to vertices as it was not divisible by 3")
	}
	output := make([]Vertex3, len(arr)/3)
	for i := 0; i < len(arr)/3; i++{
		output[i] = Vertex3{arr[i*3],arr[i*3+1],arr[i*3+2]}
	}
	return output
}

func (l *Loader) parseGeoNode ( geoNode Node , skeleton Skeleton) {
	geoInfo := make(map[string]Property)
 
	geoInfo["vertexPositions"] =  []float32{} //TODO: convert this out to preparse rather than worrying about it later
	if v, ok := geoNode.props["Vertices"]; ok{
		geoInfo["vertexPositions"] = v
	}
	geoInfo["vertexIndices"] =  []int{}
	if v, ok := geoNode.props["PolygonVertexIndex"]; ok{
		geoInfo["vertexIndices"] = v
	}

	if v, ok :=  geoNode.props["LayerElementColor"]; ok{
		geoInfo["color"] = l.parseVertexColors(v[0])
	}
	if v, ok :=  geoNode.props["LayerElementMaterial"]; ok{
		geoInfo["material"] = l.parseMaterialIndices(v[0])
	}
	if v, ok :=  geoNode.props["LayerElementNormal"]; ok{
		geoInfo["normal"] = l.parseNormals(v[0])
	}

	if uvList, ok :=  geoNode.props["LayerElementUV"]; ok{
	//TODO: correct this once we understand all things in a UV object
		geoInfo["uv"] = make([]UVParsed,len(uvList) )
		for i, v := uvList{
			geoInfo["uv"][i] = l.parseUVs(v)
		}
	}

	geoInfo["weightTable"] = ???{}

	if skeleton != null{
		geoInfo["skeleton"] = skeleton
		// loop over the bone's vertex indices and weights
		for i, rawb := range skeleton.rawBones{
			for j,rIndex := range rawb.Indices{
				if _, ok := geoInfo["weightTable"][rIndex]{
					geoInfo["weightTable"][rIndex] = []WeightEntry{}
				}
				geoInfo["weightTable"][rIndex] = append(geoInfo["weightTable"][rIndex],WeightEntry{i,rawb.weights[j]})
			}
		}
	}

	return geoInfo
}

genBuffers: function ( geoInfo ) {
	var buffers = {
		vertex: [],
		normal: [],
		colors: [],
		uvs: [],
		materialIndex: [],
		vertexWeights: [],
		weightsIndices: [],
	};
	var polygonIndex = 0;
	var faceLength = 0;
	var displayedWeightsWarning = false;
	// these will hold data for a single face
	var facePositionIndexes = [];
	var faceNormals = [];
	var faceColors = [];
	var faceUVs = [];
	var faceWeights = [];
	var faceWeightIndices = [];
	var self = this;
	geoInfo.vertexIndices.forEach( function ( vertexIndex, polygonVertexIndex ) {
		var endOfFace = false;
		// Face index and vertex index arrays are combined in a single array
		// A cube with quad faces looks like this:
		// PolygonVertexIndex: *24 {
		//  a: 0, 1, 3, -3, 2, 3, 5, -5, 4, 5, 7, -7, 6, 7, 1, -1, 1, 7, 5, -4, 6, 0, 2, -5
		//  }
		// Negative numbers mark the end of a face - first face here is 0, 1, 3, -3
		// to find index of last vertex bit shift the index: ^ - 1
		if ( vertexIndex < 0 ) {
			vertexIndex = vertexIndex ^ - 1; // equivalent to ( x * -1 ) - 1
			endOfFace = true;
		}
		var weightIndices = [];
		var weights = [];
		facePositionIndexes.push( vertexIndex * 3, vertexIndex * 3 + 1, vertexIndex * 3 + 2 );
		if ( geoInfo.color ) {
			var data = getData( polygonVertexIndex, polygonIndex, vertexIndex, geoInfo.color );
			faceColors.push( data[ 0 ], data[ 1 ], data[ 2 ] );
		}
		if ( geoInfo.skeleton ) {
			if ( geoInfo.weightTable[ vertexIndex ] !== undefined ) {
				geoInfo.weightTable[ vertexIndex ].forEach( function ( wt ) {
					weights.push( wt.Weight );
					weightIndices.push( wt.ID );
				} );
			}
			if ( weights.length > 4 ) {
				if ( ! displayedWeightsWarning ) {
					console.warn( 'THREE.FBXLoader: Vertex has more than 4 skinning weights assigned to vertex. Deleting additional weights.' );
					displayedWeightsWarning = true;
				}
				var wIndex = [ 0, 0, 0, 0 ];
				var Weight = [ 0, 0, 0, 0 ];
				weights.forEach( function ( weight, weightIndex ) {
					var currentWeight = weight;
					var currentIndex = weightIndices[ weightIndex ];
					Weight.forEach( function ( comparedWeight, comparedWeightIndex, comparedWeightArray ) {
						if ( currentWeight > comparedWeight ) {
							comparedWeightArray[ comparedWeightIndex ] = currentWeight;
							currentWeight = comparedWeight;
							var tmp = wIndex[ comparedWeightIndex ];
							wIndex[ comparedWeightIndex ] = currentIndex;
							currentIndex = tmp;
						}
					} );
				} );
				weightIndices = wIndex;
				weights = Weight;
			}
			// if the weight array is shorter than 4 pad with 0s
			while ( weights.length < 4 ) {
				weights.push( 0 );
				weightIndices.push( 0 );
			}
			for ( var i = 0; i < 4; ++ i ) {
				faceWeights.push( weights[ i ] );
				faceWeightIndices.push( weightIndices[ i ] );
			}
		}
		if ( geoInfo.normal ) {
			var data = getData( polygonVertexIndex, polygonIndex, vertexIndex, geoInfo.normal );
			faceNormals.push( data[ 0 ], data[ 1 ], data[ 2 ] );
		}
		if ( geoInfo.material && geoInfo.material.mappingType !== 'AllSame' ) {
			var materialIndex = getData( polygonVertexIndex, polygonIndex, vertexIndex, geoInfo.material )[ 0 ];
		}
		if ( geoInfo.uv ) {
			geoInfo.uv.forEach( function ( uv, i ) {
				var data = getData( polygonVertexIndex, polygonIndex, vertexIndex, uv );
				if ( faceUVs[ i ] === undefined ) {
					faceUVs[ i ] = [];
				}
				faceUVs[ i ].push( data[ 0 ] );
				faceUVs[ i ].push( data[ 1 ] );
			} );
		}
		faceLength ++;
		if ( endOfFace ) {
			self.genFace( buffers, geoInfo, facePositionIndexes, materialIndex, faceNormals, faceColors, faceUVs, faceWeights, faceWeightIndices, faceLength );
			polygonIndex ++;
			faceLength = 0;
			// reset arrays for the next face
			facePositionIndexes = [];
			faceNormals = [];
			faceColors = [];
			faceUVs = [];
			faceWeights = [];
			faceWeightIndices = [];
		}
	} );
	return buffers;
},
// Generate data for a single face in a geometry. If the face is a quad then split it into 2 tris
genFace: function ( buffers, geoInfo, facePositionIndexes, materialIndex, faceNormals, faceColors, faceUVs, faceWeights, faceWeightIndices, faceLength ) {
	for ( var i = 2; i < faceLength; i ++ ) {
		buffers.vertex.push( geoInfo.vertexPositions[ facePositionIndexes[ 0 ] ] );
		buffers.vertex.push( geoInfo.vertexPositions[ facePositionIndexes[ 1 ] ] );
		buffers.vertex.push( geoInfo.vertexPositions[ facePositionIndexes[ 2 ] ] );
		buffers.vertex.push( geoInfo.vertexPositions[ facePositionIndexes[ ( i - 1 ) * 3 ] ] );
		buffers.vertex.push( geoInfo.vertexPositions[ facePositionIndexes[ ( i - 1 ) * 3 + 1 ] ] );
		buffers.vertex.push( geoInfo.vertexPositions[ facePositionIndexes[ ( i - 1 ) * 3 + 2 ] ] );
		buffers.vertex.push( geoInfo.vertexPositions[ facePositionIndexes[ i * 3 ] ] );
		buffers.vertex.push( geoInfo.vertexPositions[ facePositionIndexes[ i * 3 + 1 ] ] );
		buffers.vertex.push( geoInfo.vertexPositions[ facePositionIndexes[ i * 3 + 2 ] ] );
		if ( geoInfo.skeleton ) {
			buffers.vertexWeights.push( faceWeights[ 0 ] );
			buffers.vertexWeights.push( faceWeights[ 1 ] );
			buffers.vertexWeights.push( faceWeights[ 2 ] );
			buffers.vertexWeights.push( faceWeights[ 3 ] );
			buffers.vertexWeights.push( faceWeights[ ( i - 1 ) * 4 ] );
			buffers.vertexWeights.push( faceWeights[ ( i - 1 ) * 4 + 1 ] );
			buffers.vertexWeights.push( faceWeights[ ( i - 1 ) * 4 + 2 ] );
			buffers.vertexWeights.push( faceWeights[ ( i - 1 ) * 4 + 3 ] );
			buffers.vertexWeights.push( faceWeights[ i * 4 ] );
			buffers.vertexWeights.push( faceWeights[ i * 4 + 1 ] );
			buffers.vertexWeights.push( faceWeights[ i * 4 + 2 ] );
			buffers.vertexWeights.push( faceWeights[ i * 4 + 3 ] );
			buffers.weightsIndices.push( faceWeightIndices[ 0 ] );
			buffers.weightsIndices.push( faceWeightIndices[ 1 ] );
			buffers.weightsIndices.push( faceWeightIndices[ 2 ] );
			buffers.weightsIndices.push( faceWeightIndices[ 3 ] );
			buffers.weightsIndices.push( faceWeightIndices[ ( i - 1 ) * 4 ] );
			buffers.weightsIndices.push( faceWeightIndices[ ( i - 1 ) * 4 + 1 ] );
			buffers.weightsIndices.push( faceWeightIndices[ ( i - 1 ) * 4 + 2 ] );
			buffers.weightsIndices.push( faceWeightIndices[ ( i - 1 ) * 4 + 3 ] );
			buffers.weightsIndices.push( faceWeightIndices[ i * 4 ] );
			buffers.weightsIndices.push( faceWeightIndices[ i * 4 + 1 ] );
			buffers.weightsIndices.push( faceWeightIndices[ i * 4 + 2 ] );
			buffers.weightsIndices.push( faceWeightIndices[ i * 4 + 3 ] );
		}
		if ( geoInfo.color ) {
			buffers.colors.push( faceColors[ 0 ] );
			buffers.colors.push( faceColors[ 1 ] );
			buffers.colors.push( faceColors[ 2 ] );
			buffers.colors.push( faceColors[ ( i - 1 ) * 3 ] );
			buffers.colors.push( faceColors[ ( i - 1 ) * 3 + 1 ] );
			buffers.colors.push( faceColors[ ( i - 1 ) * 3 + 2 ] );
			buffers.colors.push( faceColors[ i * 3 ] );
			buffers.colors.push( faceColors[ i * 3 + 1 ] );
			buffers.colors.push( faceColors[ i * 3 + 2 ] );
		}
		if ( geoInfo.material && geoInfo.material.mappingType !== 'AllSame' ) {
			buffers.materialIndex.push( materialIndex );
			buffers.materialIndex.push( materialIndex );
			buffers.materialIndex.push( materialIndex );
		}
		if ( geoInfo.normal ) {
			buffers.normal.push( faceNormals[ 0 ] );
			buffers.normal.push( faceNormals[ 1 ] );
			buffers.normal.push( faceNormals[ 2 ] );
			buffers.normal.push( faceNormals[ ( i - 1 ) * 3 ] );
			buffers.normal.push( faceNormals[ ( i - 1 ) * 3 + 1 ] );
			buffers.normal.push( faceNormals[ ( i - 1 ) * 3 + 2 ] );
			buffers.normal.push( faceNormals[ i * 3 ] );
			buffers.normal.push( faceNormals[ i * 3 + 1 ] );
			buffers.normal.push( faceNormals[ i * 3 + 2 ] );
		}
		if ( geoInfo.uv ) {
			geoInfo.uv.forEach( function ( uv, j ) {
				if ( buffers.uvs[ j ] === undefined ) buffers.uvs[ j ] = [];
				buffers.uvs[ j ].push( faceUVs[ j ][ 0 ] );
				buffers.uvs[ j ].push( faceUVs[ j ][ 1 ] );
				buffers.uvs[ j ].push( faceUVs[ j ][ ( i - 1 ) * 2 ] );
				buffers.uvs[ j ].push( faceUVs[ j ][ ( i - 1 ) * 2 + 1 ] );
				buffers.uvs[ j ].push( faceUVs[ j ][ i * 2 ] );
				buffers.uvs[ j ].push( faceUVs[ j ][ i * 2 + 1 ] );
			} );
		}
	}
},
addMorphTargets: function ( parentGeo, parentGeoNode, morphTarget, preTransform ) {
	if ( morphTarget === null ) return;
	parentGeo.morphAttributes.position = [];
	parentGeo.morphAttributes.normal = [];
	var self = this;
	morphTarget.rawTargets.forEach( function ( rawTarget ) {
		var morphGeoNode = fbxTree.Objects.Geometry[ rawTarget.geoID ];
		if ( morphGeoNode !== undefined ) {
			self.genMorphGeometry( parentGeo, parentGeoNode, morphGeoNode, preTransform );
		}
	} );
},
// a morph geometry node is similar to a standard  node, and the node is also contained
// in FBXTree.Objects.Geometry, however it can only have attributes for position, normal
// and a special attribute Index defining which vertices of the original geometry are affected
// Normal and position attributes only have data for the vertices that are affected by the morph
genMorphGeometry: function ( parentGeo, parentGeoNode, morphGeoNode, preTransform ) {
	var morphGeo = new THREE.BufferGeometry();
	if ( morphGeoNode.attrName ) morphGeo.name = morphGeoNode.attrName;
	var vertexIndices = ( parentGeoNode.PolygonVertexIndex !== undefined ) ? parentGeoNode.PolygonVertexIndex.a : [];
	// make a copy of the parent's vertex positions
	var vertexPositions = ( parentGeoNode.Vertices !== undefined ) ? parentGeoNode.Vertices.a.slice() : [];
	var morphPositions = ( morphGeoNode.Vertices !== undefined ) ? morphGeoNode.Vertices.a : [];
	var indices = ( morphGeoNode.Indexes !== undefined ) ? morphGeoNode.Indexes.a : [];
	for ( var i = 0; i < indices.length; i ++ ) {
		var morphIndex = indices[ i ] * 3;
		// FBX format uses blend shapes rather than morph targets. This can be converted
		// by additively combining the blend shape positions with the original geometry's positions
		vertexPositions[ morphIndex ] += morphPositions[ i * 3 ];
		vertexPositions[ morphIndex + 1 ] += morphPositions[ i * 3 + 1 ];
		vertexPositions[ morphIndex + 2 ] += morphPositions[ i * 3 + 2 ];
	}
	// TODO: add morph normal support
	var morphGeoInfo = {
		vertexIndices: vertexIndices,
		vertexPositions: vertexPositions,
	};
	var morphBuffers = this.genBuffers( morphGeoInfo );
	var positionAttribute = new THREE.Float32BufferAttribute( morphBuffers.vertex, 3 );
	positionAttribute.name = morphGeoNode.attrName;
	preTransform.applyToBufferAttribute( positionAttribute );
	parentGeo.morphAttributes.position.push( positionAttribute );
},
// Parse normal from FBXTree.Objects.Geometry.LayerElementNormal if it exists
parseNormals: function ( NormalNode ) {
	var mappingType = NormalNode.MappingInformationType;
	var referenceType = NormalNode.ReferenceInformationType;
	var buffer = NormalNode.Normals.a;
	var indexBuffer = [];
	if ( referenceType === 'IndexToDirect' ) {
		if ( 'NormalIndex' in NormalNode ) {
			indexBuffer = NormalNode.NormalIndex.a;
		} else if ( 'NormalsIndex' in NormalNode ) {
			indexBuffer = NormalNode.NormalsIndex.a;
		}
	}
	return {
		dataSize: 3,
		buffer: buffer,
		indices: indexBuffer,
		mappingType: mappingType,
		referenceType: referenceType
	};
},
// Parse UVs from FBXTree.Objects.Geometry.LayerElementUV if it exists
parseUVs: function ( UVNode ) {
	var mappingType = UVNode.MappingInformationType;
	var referenceType = UVNode.ReferenceInformationType;
	var buffer = UVNode.UV.a;
	var indexBuffer = [];
	if ( referenceType === 'IndexToDirect' ) {
		indexBuffer = UVNode.UVIndex.a;
	}
	return {
		dataSize: 2,
		buffer: buffer,
		indices: indexBuffer,
		mappingType: mappingType,
		referenceType: referenceType
	};
},
// Parse Vertex Colors from FBXTree.Objects.Geometry.LayerElementColor if it exists
parseVertexColors: function ( ColorNode ) {
	var mappingType = ColorNode.MappingInformationType;
	var referenceType = ColorNode.ReferenceInformationType;
	var buffer = ColorNode.Colors.a;
	var indexBuffer = [];
	if ( referenceType === 'IndexToDirect' ) {
		indexBuffer = ColorNode.ColorIndex.a;
	}
	return {
		dataSize: 4,
		buffer: buffer,
		indices: indexBuffer,
		mappingType: mappingType,
		referenceType: referenceType
	};
},
// Parse mapping and material data in FBXTree.Objects.Geometry.LayerElementMaterial if it exists
parseMaterialIndices: function ( MaterialNode ) {
	var mappingType = MaterialNode.MappingInformationType;
	var referenceType = MaterialNode.ReferenceInformationType;
	if ( mappingType === 'NoMappingInformation' ) {
		return {
			dataSize: 1,
			buffer: [ 0 ],
			indices: [ 0 ],
			mappingType: 'AllSame',
			referenceType: referenceType
		};
	}
	var materialIndexBuffer = MaterialNode.Materials.a;
	// Since materials are stored as indices, there's a bit of a mismatch between FBX and what
	// we expect.So we create an intermediate buffer that points to the index in the buffer,
	// for conforming with the other functions we've written for other data.
	var materialIndices = [];
	for ( var i = 0; i < materialIndexBuffer.length; ++ i ) {
		materialIndices.push( i );
	}
	return {
		dataSize: 1,
		buffer: materialIndexBuffer,
		indices: materialIndices,
		mappingType: mappingType,
		referenceType: referenceType
	};
},
// Generate a NurbGeometry from a node in FBXTree.Objects.Geometry
parseNurbsGeometry: function ( geoNode ) {
	if ( THREE.NURBSCurve === undefined ) {
		console.error( 'THREE.FBXLoader: The loader relies on THREE.NURBSCurve for any nurbs present in the model. Nurbs will show up as empty geometry.' );
		return new THREE.BufferGeometry();
	}
	var order = parseInt( geoNode.Order );
	if ( isNaN( order ) ) {
		console.error( 'THREE.FBXLoader: Invalid Order %s given for geometry ID: %s', geoNode.Order, geoNode.id );
		return new THREE.BufferGeometry();
	}
	var degree = order - 1;
	var knots = geoNode.KnotVector.a;
	var controlPoints = [];
	var pointsValues = geoNode.Points.a;
	for ( var i = 0, l = pointsValues.length; i < l; i += 4 ) {
		controlPoints.push( new THREE.Vector4().fromArray( pointsValues, i ) );
	}
	var startKnot, endKnot;
	if ( geoNode.Form === 'Closed' ) {
		controlPoints.push( controlPoints[ 0 ] );
	} else if ( geoNode.Form === 'Periodic' ) {
		startKnot = degree;
		endKnot = knots.length - 1 - startKnot;
		for ( var i = 0; i < degree; ++ i ) {
			controlPoints.push( controlPoints[ i ] );
		}
	}
	var curve = new THREE.NURBSCurve( degree, knots, controlPoints, startKnot, endKnot );
	var vertices = curve.getPoints( controlPoints.length * 7 );
	var positions = new Float32Array( vertices.length * 3 );
	vertices.forEach( function ( vertex, i ) {
		vertex.toArray( positions, i * 3 );
	} );
	var geometry = new THREE.BufferGeometry();
	geometry.addAttribute( 'position', new THREE.BufferAttribute( positions, 3 ) );
	return geometry;
},
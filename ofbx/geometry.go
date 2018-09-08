package ofbx


var tempMat = new THREE.Matrix4();
	var tempEuler = new THREE.Euler();
	var tempVec = new THREE.Vector3();
	var translation = new THREE.Vector3();
	var rotation = new THREE.Matrix4();
	// generate transformation from FBX transform data
	// ref: https://help.autodesk.com/view/FBX/2017/ENU/?guid=__files_GUID_10CDD63C_79C1_4F2D_BB28_AD2BE65A02ED_htm
	// transformData = {
	//	 eulerOrder: int,
	//	 translation: [],
	//   rotationOffset: [],
	//	 preRotation
	//	 rotation
	//	 postRotation
	//   scale
	// }
	// all entries are optional
	function generateTransform( transformData ) {
		var transform = new THREE.Matrix4();
		translation.set( 0, 0, 0 );
		rotation.identity();
		var order = ( transformData.eulerOrder ) ? getEulerOrder( transformData.eulerOrder ) : getEulerOrder( 0 );
		if ( transformData.translation ) translation.fromArray( transformData.translation );
		if ( transformData.rotationOffset ) translation.add( tempVec.fromArray( transformData.rotationOffset ) );
		if ( transformData.rotation ) {
			var array = transformData.rotation.map( THREE.Math.degToRad );
			array.push( order );
			rotation.makeRotationFromEuler( tempEuler.fromArray( array ) );
		}
		if ( transformData.preRotation ) {
			var array = transformData.preRotation.map( THREE.Math.degToRad );
			array.push( order );
			tempMat.makeRotationFromEuler( tempEuler.fromArray( array ) );
			rotation.premultiply( tempMat );
		}
		if ( transformData.postRotation ) {
			var array = transformData.postRotation.map( THREE.Math.degToRad );
			array.push( order );
			tempMat.makeRotationFromEuler( tempEuler.fromArray( array ) );
			tempMat.getInverse( tempMat );
			rotation.multiply( tempMat );
		}
		if ( transformData.scale ) transform.scale( tempVec.fromArray( transformData.scale ) );
		transform.setPosition( translation );
		transform.multiply( rotation );
		return transform;
	}
	var dataArray = [];
	// extracts the data from the correct position in the FBX array based on indexing type
	function getData( polygonVertexIndex, polygonIndex, vertexIndex, infoObject ) {
		var index;
		switch ( infoObject.mappingType ) {
			case 'ByPolygonVertex' :
				index = polygonVertexIndex;
				break;
			case 'ByPolygon' :
				index = polygonIndex;
				break;
			case 'ByVertice' :
				index = vertexIndex;
				break;
			case 'AllSame' :
				index = infoObject.indices[ 0 ];
				break;
			default :
				console.warn( 'THREE.FBXLoader: unknown attribute mapping type ' + infoObject.mappingType );
		}
		if ( infoObject.referenceType === 'IndexToDirect' ) index = infoObject.indices[ index ];
		var from = index * infoObject.dataSize;
		var to = from + infoObject.dataSize;
		return slice( dataArray, infoObject.buffer, from, to );
	}
	// Returns the three.js intrinsic Euler order corresponding to FBX extrinsic Euler order
	// ref: http://help.autodesk.com/view/FBX/2017/ENU/?guid=__cpp_ref_class_fbx_euler_html
	function getEulerOrder( order ) {
		var enums = [
			'ZYX', // -> XYZ extrinsic
			'YZX', // -> XZY extrinsic
			'XZY', // -> YZX extrinsic
			'ZXY', // -> YXZ extrinsic
			'YXZ', // -> ZXY extrinsic
			'XYZ', // -> ZYX extrinsic
		//'SphericXYZ', // not possible to support
		];
		if ( order === 6 ) {
			console.warn( 'THREE.FBXLoader: unsupported Euler Order: Spherical XYZ. Animations and rotations may be incorrect.' );
			return enums[ 0 ];
		}
		return enums[ order ];
	}

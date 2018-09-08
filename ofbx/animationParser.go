package ofbx

// parse animation data from FBXTree
function AnimationParser() {}
AnimationParser.prototype = {
	constructor: AnimationParser,
	// take raw animation clips and turn them into three.js animation clips
	parse: function () {
		var animationClips = [];
		var rawClips = this.parseClips();
		if ( rawClips === undefined ) return;
		for ( var key in rawClips ) {
			var rawClip = rawClips[ key ];
			var clip = this.addClip( rawClip );
			animationClips.push( clip );
		}
		return animationClips;
	},
	parseClips: function () {
		// since the actual transformation data is stored in FBXTree.Objects.AnimationCurve,
		// if this is undefined we can safely assume there are no animations
		if ( fbxTree.Objects.AnimationCurve === undefined ) return undefined;
		var curveNodesMap = this.parseAnimationCurveNodes();
		this.parseAnimationCurves( curveNodesMap );
		var layersMap = this.parseAnimationLayers( curveNodesMap );
		var rawClips = this.parseAnimStacks( layersMap );
		return rawClips;
	},
	// parse nodes in FBXTree.Objects.AnimationCurveNode
	// each AnimationCurveNode holds data for an animation transform for a model (e.g. left arm rotation )
	// and is referenced by an AnimationLayer
	parseAnimationCurveNodes: function () {
		var rawCurveNodes = fbxTree.Objects.AnimationCurveNode;
		var curveNodesMap = new Map();
		for ( var nodeID in rawCurveNodes ) {
			var rawCurveNode = rawCurveNodes[ nodeID ];
			if ( rawCurveNode.attrName.match( /S|R|T|DeformPercent/ ) !== null ) {
				var curveNode = {
					id: rawCurveNode.id,
					attr: rawCurveNode.attrName,
					curves: {},
				};
				curveNodesMap.set( curveNode.id, curveNode );
			}
		}
		return curveNodesMap;
	},
	// parse nodes in FBXTree.Objects.AnimationCurve and connect them up to
	// previously parsed AnimationCurveNodes. Each AnimationCurve holds data for a single animated
	// axis ( e.g. times and values of x rotation)
	parseAnimationCurves: function ( curveNodesMap ) {
		var rawCurves = fbxTree.Objects.AnimationCurve;
		// TODO: Many values are identical up to roundoff error, but won't be optimised
		// e.g. position times: [0, 0.4, 0. 8]
		// position values: [7.23538335023477e-7, 93.67518615722656, -0.9982695579528809, 7.23538335023477e-7, 93.67518615722656, -0.9982695579528809, 7.235384487103147e-7, 93.67520904541016, -0.9982695579528809]
		// clearly, this should be optimised to
		// times: [0], positions [7.23538335023477e-7, 93.67518615722656, -0.9982695579528809]
		// this shows up in nearly every FBX file, and generally time array is length > 100
		for ( var nodeID in rawCurves ) {
			var animationCurve = {
				id: rawCurves[ nodeID ].id,
				times: rawCurves[ nodeID ].KeyTime.a.map( convertFBXTimeToSeconds ),
				values: rawCurves[ nodeID ].KeyValueFloat.a,
			};
			var relationships = connections.get( animationCurve.id );
			if ( relationships !== undefined ) {
				var animationCurveID = relationships.parents[ 0 ].ID;
				var animationCurveRelationship = relationships.parents[ 0 ].relationship;
				if ( animationCurveRelationship.match( /X/ ) ) {
					curveNodesMap.get( animationCurveID ).curves[ 'x' ] = animationCurve;
				} else if ( animationCurveRelationship.match( /Y/ ) ) {
					curveNodesMap.get( animationCurveID ).curves[ 'y' ] = animationCurve;
				} else if ( animationCurveRelationship.match( /Z/ ) ) {
					curveNodesMap.get( animationCurveID ).curves[ 'z' ] = animationCurve;
				} else if ( animationCurveRelationship.match( /d|DeformPercent/ ) && curveNodesMap.has( animationCurveID ) ) {
					curveNodesMap.get( animationCurveID ).curves[ 'morph' ] = animationCurve;
				}
			}
		}
	},
	// parse nodes in FBXTree.Objects.AnimationLayer. Each layers holds references
	// to various AnimationCurveNodes and is referenced by an AnimationStack node
	// note: theoretically a stack can have multiple layers, however in practice there always seems to be one per stack
	parseAnimationLayers: function ( curveNodesMap ) {
		var rawLayers = fbxTree.Objects.AnimationLayer;
		var layersMap = new Map();
		for ( var nodeID in rawLayers ) {
			var layerCurveNodes = [];
			var connection = connections.get( parseInt( nodeID ) );
			if ( connection !== undefined ) {
				// all the animationCurveNodes used in the layer
				var children = connection.children;
				var self = this;
				children.forEach( function ( child, i ) {
					if ( curveNodesMap.has( child.ID ) ) {
						var curveNode = curveNodesMap.get( child.ID );
						// check that the curves are defined for at least one axis, otherwise ignore the curveNode
						if ( curveNode.curves.x !== undefined || curveNode.curves.y !== undefined || curveNode.curves.z !== undefined ) {
							if ( layerCurveNodes[ i ] === undefined ) {
								var modelID;
								connections.get( child.ID ).parents.forEach( function ( parent ) {
									if ( parent.relationship !== undefined ) modelID = parent.ID;
								} );
								var rawModel = fbxTree.Objects.Model[ modelID.toString() ];
								var node = {
									modelName: THREE.PropertyBinding.sanitizeNodeName( rawModel.attrName ),
									initialPosition: [ 0, 0, 0 ],
									initialRotation: [ 0, 0, 0 ],
									initialScale: [ 1, 1, 1 ],
									transform: self.getModelAnimTransform( rawModel ),
								};
								// if the animated model is pre rotated, we'll have to apply the pre rotations to every
								// animation value as well
								if ( 'PreRotation' in rawModel ) node.preRotations = rawModel.PreRotation.value;
								if ( 'PostRotation' in rawModel ) node.postRotations = rawModel.PostRotation.value;
								layerCurveNodes[ i ] = node;
							}
							layerCurveNodes[ i ][ curveNode.attr ] = curveNode;
						} else if ( curveNode.curves.morph !== undefined ) {
							if ( layerCurveNodes[ i ] === undefined ) {
								var deformerID;
								connections.get( child.ID ).parents.forEach( function ( parent ) {
									if ( parent.relationship !== undefined ) deformerID = parent.ID;
								} );
								var morpherID = connections.get( deformerID ).parents[ 0 ].ID;
								var geoID = connections.get( morpherID ).parents[ 0 ].ID;
								// assuming geometry is not used in more than one model
								var modelID = connections.get( geoID ).parents[ 0 ].ID;
								var rawModel = fbxTree.Objects.Model[ modelID ];
								var node = {
									modelName: THREE.PropertyBinding.sanitizeNodeName( rawModel.attrName ),
									morphName: fbxTree.Objects.Deformer[ deformerID ].attrName,
								};
								layerCurveNodes[ i ] = node;
							}
							layerCurveNodes[ i ][ curveNode.attr ] = curveNode;
						}
					}
				} );
				layersMap.set( parseInt( nodeID ), layerCurveNodes );
			}
		}
		return layersMap;
	},
	getModelAnimTransform: function ( modelNode ) {
		var transformData = {};
		if ( 'RotationOrder' in modelNode ) transformData.eulerOrder = parseInt( modelNode.RotationOrder.value );
		if ( 'Lcl_Translation' in modelNode ) transformData.translation = modelNode.Lcl_Translation.value;
		if ( 'RotationOffset' in modelNode ) transformData.rotationOffset = modelNode.RotationOffset.value;
		if ( 'Lcl_Rotation' in modelNode ) transformData.rotation = modelNode.Lcl_Rotation.value;
		if ( 'PreRotation' in modelNode ) transformData.preRotation = modelNode.PreRotation.value;
		if ( 'PostRotation' in modelNode ) transformData.postRotation = modelNode.PostRotation.value;
		if ( 'Lcl_Scaling' in modelNode ) transformData.scale = modelNode.Lcl_Scaling.value;
		return generateTransform( transformData );
	},
	// parse nodes in FBXTree.Objects.AnimationStack. These are the top level node in the animation
	// hierarchy. Each Stack node will be used to create a THREE.AnimationClip
	parseAnimStacks: function ( layersMap ) {
		var rawStacks = fbxTree.Objects.AnimationStack;
		// connect the stacks (clips) up to the layers
		var rawClips = {};
		for ( var nodeID in rawStacks ) {
			var children = connections.get( parseInt( nodeID ) ).children;
			if ( children.length > 1 ) {
				// it seems like stacks will always be associated with a single layer. But just in case there are files
				// where there are multiple layers per stack, we'll display a warning
				console.warn( 'THREE.FBXLoader: Encountered an animation stack with multiple layers, this is currently not supported. Ignoring subsequent layers.' );
			}
			var layer = layersMap.get( children[ 0 ].ID );
			rawClips[ nodeID ] = {
				name: rawStacks[ nodeID ].attrName,
				layer: layer,
			};
		}
		return rawClips;
	},
	addClip: function ( rawClip ) {
		var tracks = [];
		var self = this;
		rawClip.layer.forEach( function ( rawTracks ) {
			tracks = tracks.concat( self.generateTracks( rawTracks ) );
		} );
		return new THREE.AnimationClip( rawClip.name, - 1, tracks );
	},
	generateTracks: function ( rawTracks ) {
		var tracks = [];
		var initialPosition = new THREE.Vector3();
		var initialRotation = new THREE.Quaternion();
		var initialScale = new THREE.Vector3();
		if ( rawTracks.transform ) rawTracks.transform.decompose( initialPosition, initialRotation, initialScale );
		initialPosition = initialPosition.toArray();
		initialRotation = new THREE.Euler().setFromQuaternion( initialRotation ).toArray(); // todo: euler order
		initialScale = initialScale.toArray();
		if ( rawTracks.T !== undefined && Object.keys( rawTracks.T.curves ).length > 0 ) {
			var positionTrack = this.generateVectorTrack( rawTracks.modelName, rawTracks.T.curves, initialPosition, 'position' );
			if ( positionTrack !== undefined ) tracks.push( positionTrack );
		}
		if ( rawTracks.R !== undefined && Object.keys( rawTracks.R.curves ).length > 0 ) {
			var rotationTrack = this.generateRotationTrack( rawTracks.modelName, rawTracks.R.curves, initialRotation, rawTracks.preRotations, rawTracks.postRotations );
			if ( rotationTrack !== undefined ) tracks.push( rotationTrack );
		}
		if ( rawTracks.S !== undefined && Object.keys( rawTracks.S.curves ).length > 0 ) {
			var scaleTrack = this.generateVectorTrack( rawTracks.modelName, rawTracks.S.curves, initialScale, 'scale' );
			if ( scaleTrack !== undefined ) tracks.push( scaleTrack );
		}
		if ( rawTracks.DeformPercent !== undefined ) {
			var morphTrack = this.generateMorphTrack( rawTracks );
			if ( morphTrack !== undefined ) tracks.push( morphTrack );
		}
		return tracks;
	},
	generateVectorTrack: function ( modelName, curves, initialValue, type ) {
		var times = this.getTimesForAllAxes( curves );
		var values = this.getKeyframeTrackValues( times, curves, initialValue );
		return new THREE.VectorKeyframeTrack( modelName + '.' + type, times, values );
	},
	generateRotationTrack: function ( modelName, curves, initialValue, preRotations, postRotations ) {
		if ( curves.x !== undefined ) {
			this.interpolateRotations( curves.x );
			curves.x.values = curves.x.values.map( THREE.Math.degToRad );
		}
		if ( curves.y !== undefined ) {
			this.interpolateRotations( curves.y );
			curves.y.values = curves.y.values.map( THREE.Math.degToRad );
		}
		if ( curves.z !== undefined ) {
			this.interpolateRotations( curves.z );
			curves.z.values = curves.z.values.map( THREE.Math.degToRad );
		}
		var times = this.getTimesForAllAxes( curves );
		var values = this.getKeyframeTrackValues( times, curves, initialValue );
		if ( preRotations !== undefined ) {
			preRotations = preRotations.map( THREE.Math.degToRad );
			preRotations.push( 'ZYX' );
			preRotations = new THREE.Euler().fromArray( preRotations );
			preRotations = new THREE.Quaternion().setFromEuler( preRotations );
		}
		if ( postRotations !== undefined ) {
			postRotations = postRotations.map( THREE.Math.degToRad );
			postRotations.push( 'ZYX' );
			postRotations = new THREE.Euler().fromArray( postRotations );
			postRotations = new THREE.Quaternion().setFromEuler( postRotations ).inverse();
		}
		var quaternion = new THREE.Quaternion();
		var euler = new THREE.Euler();
		var quaternionValues = [];
		for ( var i = 0; i < values.length; i += 3 ) {
			euler.set( values[ i ], values[ i + 1 ], values[ i + 2 ], 'ZYX' );
			quaternion.setFromEuler( euler );
			if ( preRotations !== undefined ) quaternion.premultiply( preRotations );
			if ( postRotations !== undefined ) quaternion.multiply( postRotations );
			quaternion.toArray( quaternionValues, ( i / 3 ) * 4 );
		}
		return new THREE.QuaternionKeyframeTrack( modelName + '.quaternion', times, quaternionValues );
	},
	generateMorphTrack: function ( rawTracks ) {
		var curves = rawTracks.DeformPercent.curves.morph;
		var values = curves.values.map( function ( val ) {
			return val / 100;
		} );
		var morphNum = sceneGraph.getObjectByName( rawTracks.modelName ).morphTargetDictionary[ rawTracks.morphName ];
		return new THREE.NumberKeyframeTrack( rawTracks.modelName + '.morphTargetInfluences[' + morphNum + ']', curves.times, values );
	},
	// For all animated objects, times are defined separately for each axis
	// Here we'll combine the times into one sorted array without duplicates
	getTimesForAllAxes: function ( curves ) {
		var times = [];
		// first join together the times for each axis, if defined
		if ( curves.x !== undefined ) times = times.concat( curves.x.times );
		if ( curves.y !== undefined ) times = times.concat( curves.y.times );
		if ( curves.z !== undefined ) times = times.concat( curves.z.times );
		// then sort them and remove duplicates
		times = times.sort( function ( a, b ) {
			return a - b;
		} ).filter( function ( elem, index, array ) {
			return array.indexOf( elem ) == index;
		} );
		return times;
	},
	getKeyframeTrackValues: function ( times, curves, initialValue ) {
		var prevValue = initialValue;
		var values = [];
		var xIndex = - 1;
		var yIndex = - 1;
		var zIndex = - 1;
		times.forEach( function ( time ) {
			if ( curves.x ) xIndex = curves.x.times.indexOf( time );
			if ( curves.y ) yIndex = curves.y.times.indexOf( time );
			if ( curves.z ) zIndex = curves.z.times.indexOf( time );
			// if there is an x value defined for this frame, use that
			if ( xIndex !== - 1 ) {
				var xValue = curves.x.values[ xIndex ];
				values.push( xValue );
				prevValue[ 0 ] = xValue;
			} else {
				// otherwise use the x value from the previous frame
				values.push( prevValue[ 0 ] );
			}
			if ( yIndex !== - 1 ) {
				var yValue = curves.y.values[ yIndex ];
				values.push( yValue );
				prevValue[ 1 ] = yValue;
			} else {
				values.push( prevValue[ 1 ] );
			}
			if ( zIndex !== - 1 ) {
				var zValue = curves.z.values[ zIndex ];
				values.push( zValue );
				prevValue[ 2 ] = zValue;
			} else {
				values.push( prevValue[ 2 ] );
			}
		} );
		return values;
	},
	// Rotations are defined as Euler angles which can have values  of any size
	// These will be converted to quaternions which don't support values greater than
	// PI, so we'll interpolate large rotations
	interpolateRotations: function ( curve ) {
		for ( var i = 1; i < curve.values.length; i ++ ) {
			var initialValue = curve.values[ i - 1 ];
			var valuesSpan = curve.values[ i ] - initialValue;
			var absoluteSpan = Math.abs( valuesSpan );
			if ( absoluteSpan >= 180 ) {
				var numSubIntervals = absoluteSpan / 180;
				var step = valuesSpan / numSubIntervals;
				var nextValue = initialValue + step;
				var initialTime = curve.times[ i - 1 ];
				var timeSpan = curve.times[ i ] - initialTime;
				var interval = timeSpan / numSubIntervals;
				var nextTime = initialTime + interval;
				var interpolatedTimes = [];
				var interpolatedValues = [];
				while ( nextTime < curve.times[ i ] ) {
					interpolatedTimes.push( nextTime );
					nextTime += interval;
					interpolatedValues.push( nextValue );
					nextValue += step;
				}
				curve.times = inject( curve.times, i, interpolatedTimes );
				curve.values = inject( curve.values, i, interpolatedValues );
			}
		}
	},
};	
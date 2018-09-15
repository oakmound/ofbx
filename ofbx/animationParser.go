package threefbx

import (
	"github.com/oakmound/oak/alg"
	"github.com/oakmound/oak/alg/floatgeom"
)


type Animation struct {}

// parse animation data from FBXTree
// take raw animation clips and turn them into three.js animation clips
func (l *Loader) parseAnimations() []Animation {
	// since the actual transformation data is stored in FBXTree.Objects.AnimationCurve,
	// if this is undefined we can safely assume there are no animations
	if _, ok := l.tree.Objects["AnimationCurve"]; !ok {
		return []Animation{}
	} 
	rawClips := l.parseClips()

	animationClips := make([]Animation{}, len(rawClips))
	for i, v := range rawClips {
		animationClips[i] = l.addClip(v)
	}
	return animationClips
}

func (l *Loader) parseClips() map[int64]Clip {
	curveNodesMap := l.parseAnimationCurveNodes()
	l.parseAnimationCurves(curveNodesMap)
	layersMap := l.parseAnimationLayers(curveNodesMap)
	rawClips := l.parseAnimStacks(layersMap)
	return rawClips
}

// parse nodes in FBXTree.Objects.AnimationCurveNode
// each AnimationCurveNode holds data for an animation transform for a model (e.g. left arm rotation )
// and is referenced by an AnimationLayer
func (l *Loader) parseAnimationCurveNodes() map[int64]CurveNode {
	rawCurveNodes := fbxTree.Objects["AnimationCurveNode"]
	curveNodesMap := make(map[int64]CurveNode)
	for _, node := range rawCurveNodes {
		if regexp.Match("/S|R|T|DeformPercent/", node.attrName) {
			curveNode := CurveNode{
				id: node.id,
				attr: node.attrName,
				curves: map[string]AnimationCurve,
			}
			curveNodesMap[curveNode.id] = curveNode
		}
	}
	return curveNodesMap
}

type AnimationCurve struct {
	ID int64
	Times []float64
	Values []float64
}

// parse nodes in FBXTree.Objects.AnimationCurve and connect them up to
// previously parsed AnimationCurveNodes. Each AnimationCurve holds data for a single animated
// axis ( e.g. times and values of x rotation)
func (l *Loader) parseAnimationCurves( curveNodesMap ) {
	rawCurves := fbxTree.Objects["AnimationCurve"]
	// TODO: Many values are identical up to roundoff error, but won't be optimised
	// e.g. position times: [0, 0.4, 0. 8]
	// position values: [7.23538335023477e-7, 93.67518615722656, -0.9982695579528809, 7.23538335023477e-7, 93.67518615722656, -0.9982695579528809, 7.235384487103147e-7, 93.67520904541016, -0.9982695579528809]
	// clearly, this should be optimised to
	// times: [0], positions [7.23538335023477e-7, 93.67518615722656, -0.9982695579528809]
	// this shows up in nearly every FBX file, and generally time array is length > 100
	for nodeID := range rawCurves {
		animationCurve := AnimationCurve{
			ID: rawCurves[ nodeID ].ID, 
			Times: rawCurves[ nodeID ]["KeyTime"].Payload().([]float64),
			Values: rawCurves[ nodeID ]["KeyValueFloat"].Payload().([]float64),
		}
		for i, t := range animationCurve.Times {
			animationCurve.Times[i] = fbxTimeToSeconds(t)
		}
		 relationships := connections[animationCurve.id]
		if relationships != nil {
			animationCurveID := relationships.parents[0].ID
			animationCurveRelationship := relationships.parents[0].relationship
			if strings.Contains(animationCurveRelationship, "X") {
				curveNodesMap[animationCurveID].curves["x"] = animationCurve
			} else if strings.Contains(animationCurveRelationship, "Y") {
				curveNodesMap[animationCurveID].curves["y"] = animationCurve
			} else if strings.Contains(animationCurveRelationship, "Z") {
				curveNodesMap[animationCurveID].curves["z"] = animationCurve
			} else {
				_, ok := curveNodesMap[animationCurveID]
				if ok && regexp.Match("/d|DeformPercent/", animationCurveRelationship) {
					curveNodesMap[animationCurveID].curves["morph"] = animationCurve
				}
			}
		}
	}
}

type CurveNode struct {
	modelName string
	morphName string
	initialPosition floatgeom.Point3
	initialRotation floatgeom.Point3
	initialScale floatgeom.Point3
	T floatgeom.Point3
	R floatgeom.Matrix4
	S floatgeom.Point3
	transform floatgeom.Matrix4

}

// parse nodes in FBXTree.Objects.AnimationLayer. Each layers holds references
// to various AnimationCurveNodes and is referenced by an AnimationStack node
// note: theoretically a stack can have multiple layers, however in practice there always seems to be one per stack
func (l *Loader) parseAnimationLayers( curveNodesMap ) map[int64][]CurveNode {
	rawLayers := fbxTree.Objects["AnimationLayer"]
	layersMap := make(map[int64][]CurveNode)
	for nodeID := range rawLayers {
		layerCurveNodes := []CurveNode{}
		if connection, ok := l.connections[nodeID]; ok {
			// all the animationCurveNodes used in the layer
			for i, child := range children {
				if curvenode, ok := curveNodesMap[child.ID]; ok {
					// check that the curves are defined for at least one axis, otherwise ignore the curveNode
					_, ok := curvenode.curves["x"]
					_, ok2 := curvenode.curves["y"]
					_, ok3 := curvenode.curves["z"]
					ok |= ok2 
					ok |= ok3
					if ok {
						if _, ok := layerCurveNodes[ i ]; !ok {
							var modelID int64
							for i := len(connection.parents)-1; i >= 0; i-- {
								parent := connection.parents[i]
								if parent.relationship != "" {
									modelID = parent.ID
									break
								}
							}
							rawModel := fbxTree.Objects.Model[modelID]
							node := CurveNode{
								modelName: THREE.PropertyBinding.sanitizeNodeName(rawModel.attrName),
								initialPosition: floatgeom.Point3{0, 0, 0},
								initialRotation: floatgeom.Point3{0, 0, 0},
								initialScale: floatgeom.Point3{1, 1, 1},
								transform: self.getModelAnimTransform(rawModel),
							}
							// if the animated model is pre rotated, we'll have to apply the pre rotations to every
							// animation value as well
							if v, ok := rawModel["PreRotation"]; ok {
								node.preRotations = v
							} 
							if v, ok := rawModel["PostRotation"]; ok {
								node.postRotations = v
							} 
							layerCurveNodes[i] = node
						}
						layerCurveNodes[i][curveNode.attr] = curveNode
					} else if _, ok := curveNode.curves["morph"]; ok {					
						if _, ok := layerCurveNodes[i]; !ok {
							var deformerID int64
							for i := len(connection.parents)-1; i >= 0; i-- {
								parent := connection.parents[i]
								if parent.relationship != "" {
									deformerID = parent.ID
									break
								}
							}
							morpherID := connections[deformerID].parents[0].ID
							geoID := connections[morpherID].parents[0].ID
							// assuming geometry is not used in more than one model
							modelID := connections[geoID].parents[0].ID
							rawModel := fbxTree.Objects.Model[modelID]
							var node = CurveNode{
								modelName: THREE.PropertyBinding.sanitizeNodeName( rawModel.attrName ),
								morphName: fbxTree.Objects.Deformer[ deformerID ].attrName,
							}
							layerCurveNodes[i] = node;
						}
						layerCurveNodes[i][curveNode.attr] = curveNode
					}
				}
			}
			layersMap[nodeID] = layerCurveNodes
		}
	}
	return layersMap
}

// All fields are optional
type TransformData struct {
	eulerOrder *EulerOrder
	translation *floatgeom.Point3
	rotationOffset *floatgeom.Point3
	rotation *floatgeom.Matrix4
	preRotation *floatgeom.Matrix4
	postRotation *floatgeom.Matrix4
	scale *floatgeom.Point3
}

func (l *Loader) getModelAnimTransform(modelNode Node) {
	var td TransformData
	if v, ok := modelNode.props["RotationOrder"]; ok {
		eo, err = strconv.Atoi(v.Payload().(string))
		if err != nil  || eo >= LastEulerOrder  || eo < 0{
			fmt.Println("Error decoding euler rotation order")
		} else {
			td.eulerOrder = &EulerOrder(eo)
		}
	}
	if v, ok := modelNode.props["Lcl_Translation"]; ok {
		td.translation = &v.Payload().(floatgeom.Point3)
	}
	if v, ok := modelNode.props["RotationOffset"]; ok {
		td.rotationOffset = &v.Payload().(floatgeom.Point3)
	}
	if v, ok := modelNode.props["Lcl_Rotation"]; ok {
		td.rotation = &v.Payload().(floatgeom.Matrix4)
	}
	if v, ok := modelNode.props["PreRotation"]; ok {
		td.preRotation = &v.Payload().(floatgeom.Matrix4)
	}
	if v, ok := modelNode.props["PostRotation"]; ok {
		td.postRotation = &v.Payload().(floatgeom.Matrix4)
	}
	if v, ok := modelNode.props["Lcl_Scaling"]; ok {
		 td.scale = &v.Payload().(floatgeom.Point3)
	}
	return generateTransform(td)
}

type Clip struct {
	name string
	layer []CurveNode
}

// parse nodes in FBXTree.Objects.AnimationStack. These are the top level node in the animation
// hierarchy. Each Stack node will be used to create a THREE.AnimationClip
func (l *Loader) parseAnimStacks(layersMap map[int64][]CurveNode) map[int64]Clip {
	rawStacks := fbxTree.Objects.AnimationStack
	// connect the stacks (clips) up to the layers
	rawClips := make(map[int64]Clip, len(rawStacks))
	for _, nodeID := range rawStacks {
		children := connections[nodeID].children
		if len(children) > 1 {
			// it seems like stacks will always be associated with a single layer. But just in case there are files
			// where there are multiple layers per stack, we'll display a warning
			fmt.Println("THREE.FBXLoader: Encountered an animation stack with multiple layers, this is currently not supported. Ignoring subsequent layers.");
		}
		layer := layersMap[children[0].ID]
		rawClips[nodeID] = Clip{
			name: rawStacks[ nodeID ].attrName,
			layer: layer,
		}
	}
	return rawClips
}

func (l *Loader) addClip(clip Clip) Animation {
	tracks := make([]Track, len(rawTracks))
	for _, rawTracks := range clip.layer {
		tracks = append(tracks, l.generateTracks(rawTracks))
	}
	// ??
	return new THREE.AnimationClip( rawClip.name, - 1, tracks )
}

func (l *Loader) generateTracks(rawTracks CurveNode) []Track {
	tracks := []Track{}
	initialPosition := floatgeom.Point3{}
	// Could be a matrix
	initialRotation := floatgeom.Point4{}
	initialScale := floatgeom.Point3{}
	if !rawTracks.transform.Zero() {
		// todo: does this perform side effects, or return a transform, or return three new values?
		rawTracks.transform = rawTracks.transform.decompose( initialPosition, initialRotation, initialScale)
	}
	// ????
	//initialRotation = new THREE.Euler().setFromQuaternion( initialRotation ).toArray(); // todo: euler order
	if ( rawTracks.T != undefined && Object.keys(rawTracks.T.curves).length > 0 ) {
		positionTrack, err := l.generateVectorTrack(rawTracks.modelName, rawTracks.T.curves, initialPosition, "position");
		if err != nil {
			tracks.push(positionTrack)
		}
	}
	if ( rawTracks.R != undefined && Object.keys(rawTracks.R.curves).length > 0 ) {
		rotationTrack, err := l.generateRotationTrack(rawTracks.modelName, rawTracks.R.curves, initialRotation, rawTracks.preRotations, rawTracks.postRotations);
		if err != nil {
			tracks.push(rotationTrack)
		}
	}
	if ( rawTracks.S != undefined && Object.keys(rawTracks.S.curves).length > 0 ) {
		scaleTrack, err := l.generateVectorTrack(rawTracks.modelName, rawTracks.S.curves, initialScale, "scale");
		if err != nil {
			tracks.push(scaleTrack)
		}
	}
	if ( rawTracks.DeformPercent != undefined ) {
		morphTrack, err := l.generateMorphTrack(rawTracks)
		if err != nil {
			tracks.push(morphTrack)
		}
	}
	return tracks
}

func (l *Loader) generateVectorTrack( modelName string, curves ???, initialValue ???, typ string ) {
	times := l.getTimesForAllAxes(curves)
	values := l.getKeyframeTrackValues(times, curves, initialValue)
	// ???
	return new THREE.VectorKeyframeTrack(modelName + "." + typ, times, values)
}

func (l *Loader) generateRotationTrack(modelName string, curves map[string]AnimationCurve, initialValue, preRotations, postRotations ??) {
	if c, ok := curves["x"]; ok {
		c = l.interpolateRotations(c)
		for i, val := range c.values {
			c.values[i] = alg.DegToRad * val
		}
	}
	if c, ok := curves["y"]; ok {
		c = l.interpolateRotations(c)
		for i, val := range c.values {
			c.values[i] = alg.DegToRad * val
		}
	}
	if c, ok := curves["z"]; ok {
		c = l.interpolateRotations(c)
		for i, val := range c.values {
			c.values[i] = alg.DegToRad * val
		}
	}
	times = l.getTimesForAllAxes(curves);
	values = l.getKeyframeTrackValues( times, curves, initialValue )
	// ????
	if ( preRotations != undefined ) {
		preRotations = preRotations.map( THREE.Math.degToRad )
		preRotations.push( 'ZYX' )
		preRotations = new THREE.Euler().fromArray( preRotations )
		preRotations = new THREE.Quaternion().setFromEuler( preRotations )
	}
	if ( postRotations != undefined ) {
		postRotations = postRotations.map( THREE.Math.degToRad )
		postRotations.push("ZYX")
		postRotations = new THREE.Euler().fromArray( postRotations )
		postRotations = new THREE.Quaternion().setFromEuler( postRotations ).inverse()
	}
	quaternion = new THREE.Quaternion();
	euler = new THREE.Euler();
	quaternionValues = [];
	for i := 0; i < len(values); i += 3 {
		euler.set( values[ i ], values[ i + 1 ], values[ i + 2 ], 'ZYX' )
		quaternion.setFromEuler( euler )
		if ( preRotations !== undefined ) quaternion.premultiply( preRotations )
		if ( postRotations !== undefined ) quaternion.multiply( postRotations )
		quaternion.toArray( quaternionValues, ( i / 3 ) * 4 )
	}
	return new THREE.QuaternionKeyframeTrack(modelName + ".quaternion", times, quaternionValues)
	// end ????
}

func (l *Loader) generateMorphTrack(rawTracks CurveNode) ??? {
	curves := rawTracks.DeformPercent.curves["morph"]
	values := make([]float64, len(curves.values))
	for i, val := range curves.values {
		values[i] = val / 100
	}
	morphNum := sceneGraph.getObjectByName(rawTracks.modelName).morphTargetDictionary[rawTracks.morphName]
	??
	return new THREE.NumberKeyframeTrack(rawTracks.modelName + ".morphTargetInfluences[" + morphNum + "]", curves.times, values )
}

// For all animated objects, times are defined separately for each axis
// Here we'll combine the times into one sorted array without duplicates
func (l *Loader) getTimesForAllAxes(curves map[string]AnimationCurve) []float64 {
	var times = []float64{}
	// first join together the times for each axis, if defined
	if c, ok := curves["x"]; ok {
		times = append(times, c.times)
	}
	if c, ok := curves["y"]; ok {
		times = append(times, c.times)
	}
	if c, ok := curves["z"]; ok {
		times = append(times, c.times)
	}
	// then sort them and remove duplicates
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})
	i := 0
	for {
		if i >= len(times)-1 {
			break
		}
		if times[i] == times[i+1] {
			times = append(times[:i], times[i+1:])
			continue
		}
		i++
	}
	return times
}

func (l *Loader) getKeyframeTrackValues( times []float64, curves map[string]AnimationCurve, initialValue floatgeom.Point3) []float64 {
	prevValue := initialValue
	values := []float64{}
	xIndex := -1
	yIndex := -1
	zIndex := -1
	for _, time := range times {
		if _, ok := curves["x"]; ok { 
			for i, t2 := range curve["x"].times {
				if t2 == t {
					xIndex = i
					break
				}
			} 
		}
		if _, ok := curves["y"]; ok { 
			for i, t2 := range curve["y"].times {
				if t2 == t {
					yIndex = i
					break
				}
			} 
		}
		if _, ok := curves["z"]; ok { 
			for i, t2 := range curve["z"].times {
				if t2 == t {
					zIndex = i
					break
				}
			} 
		}
		// if there is an x value defined for this frame, use that
		if xIndex != - 1 {
			xValue = curves["x"].values[xIndex]
			values = append(values, xValue)
			prevValue[0] = xValue
		} else {
			// otherwise use the x value from the previous frame
			values = append(values, prevValue[0])
		}
		if yIndex != - 1 {
			yValue = curves["y"].values[yIndex]
			values = append(values, yValue)
			prevValue[1] = yValue
		} else {
			values = append(values, prevValue[1])
		}
		if zIndex != - 1 {
			zValue = curves["z"].values[zIndex]
			values = append(values, zValue)
			prevValue[z] = zValue
		} else {
			values = append(values, prevValue[2])
		}
	}
	return values
}
// Rotations are defined as Euler angles which can have values  of any size
// These will be converted to quaternions which don't support values greater than
// PI, so we'll interpolate large rotations
func (l *Loader) interpolateRotations(curve AnimationCurve) AnimationCurve {
	for i := 1; i < len(curve.values); i++ {
		initialValue := curve.values[i - 1]
		valuesSpan := curve.values[i] - initialValue
		absoluteSpan := math.Abs(valuesSpan)
		if absoluteSpan >= 180 {
			numSubIntervals := absoluteSpan / 180
			step := valuesSpan / numSubIntervals
			nextValue := initialValue + step
			initialTime := curve.times[ i - 1 ]
			timeSpan := curve.times[ i ] - initialTime
			interval := timeSpan / numSubIntervals
			nextTime := initialTime + interval
			interpolatedTimes := []float64{}
			interpolatedValues := []float64{}
			while ( nextTime < curve.times[ i ] ) {
				interpolatedTimes.push(nextTime)
				nextTime += interval
				interpolatedValues.push(nextValue)
				nextValue += step
			}
			curve.times = append(curve.times[:i], append(interpolatedTimes, curve.times[i:]))
			curve.values = append(curve.values[:i], append(interpolatedValues, curve.values[i:]))
		}
	}
}
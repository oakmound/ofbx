package threefbx

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/oakmound/oak/alg"
	"github.com/oakmound/oak/alg/floatgeom"
)

//Works like three.js animationclip.parse

// parse animation data from FBXTree
// take raw animation clips and turn them into three.js animation clips
func (l *Loader) parseAnimations() map[IDType]Animation {
	// since the actual transformation data is stored in FBXTree.Objects.AnimationCurve,
	// if this is undefined we can safely assume there are no animations
	if _, ok := l.tree.Objects["AnimationCurve"]; !ok {
		return map[IDType]Animation{}
	}
	rawClips := l.parseClips()

	animationClips := make(map[IDType]Animation, len(rawClips))
	for i, v := range rawClips {
		animationClips[i] = l.addClip(v)
	}
	return animationClips
}

func (l *Loader) parseClips() map[IDType]Clip {
	curveNodesMap := l.parseAnimationCurveNodes()
	fmt.Println("CurveNodesMap:", curveNodesMap)
	l.parseAnimationCurves(curveNodesMap)
	layersMap := l.parseAnimationLayers(curveNodesMap)
	fmt.Println("Layers Map", layersMap)
	rawClips := l.parseAnimStacks(layersMap)
	return rawClips
}

// parse nodes in FBXTree.Objects.AnimationCurveNode
// each AnimationCurveNode holds data for an animation transform for a model (e.g. left arm rotation )
// and is referenced by an AnimationLayer
func (l *Loader) parseAnimationCurveNodes() map[IDType]CurveNode {
	rawCurveNodes := l.tree.Objects["AnimationCurveNode"]
	curveNodesMap := make(map[IDType]CurveNode)
	for _, node := range rawCurveNodes {
		fmt.Println(" lengthy of ", len(node.attrName))
		fmt.Println("AnimCurve Name:", string(node.attrName[0:1]), " len of ", len(node.attrName))
		//if match, _ := regexp.Match("[S|R|T|DeformPercent]", []byte(node.attrName)); match {
		switch string(node.attrName[:1]) {
		case "S", "T", "R", "DeformPercent":

			curveNode := CurveNode{
				ID:       node.ID,
				AttrName: node.attrName,
				curves:   map[string]AnimationCurve{},
			}
			curveNodesMap[curveNode.ID] = curveNode
		default:
			fmt.Println("AnimCurve was found with name", node.attrName)
		}
	}
	return curveNodesMap
}

// parse nodes in FBXTree.Objects.AnimationCurve and connect them up to
// previously parsed AnimationCurveNodes. Each AnimationCurve holds data for a single animated
// axis ( e.g. times and values of x rotation)
func (l *Loader) parseAnimationCurves(curveNodesMap map[IDType]CurveNode) {
	rawCurves := l.tree.Objects["AnimationCurve"]
	// TODO: Many values are identical up to roundoff error, but won't be optimised
	// e.g. position times: [0, 0.4, 0. 8]
	// position values: [7.23538335023477e-7, 93.67518615722656, -0.9982695579528809, 7.23538335023477e-7, 93.67518615722656, -0.9982695579528809, 7.235384487103147e-7, 93.67520904541016, -0.9982695579528809]
	// clearly, this should be optimised to
	// times: [0], positions [7.23538335023477e-7, 93.67518615722656, -0.9982695579528809]
	// this shows up in nearly every FBX file, and generally time array is length > 100
	for nodeID := range rawCurves {
		fbxTimes := rawCurves[nodeID].props["KeyTime"].Payload.([]int64)
		animationCurve := AnimationCurve{
			ID:     rawCurves[nodeID].ID,
			Times:  make([]float64, len(fbxTimes)),
			Values: rawCurves[nodeID].props["KeyValueFloat"].Payload.([]float64),
		}
		for i, t := range fbxTimes {
			animationCurve.Times[i] = fbxTimeToSeconds(t)
		}
		relationships, ok := l.connections[animationCurve.ID]
		if ok {
			animationCurveID := relationships.parents[0].ID
			animationCurveRelationship := relationships.parents[0].Property
			if strings.Contains(animationCurveRelationship, "X") {
				curveNodesMap[animationCurveID].curves["x"] = animationCurve
			} else if strings.Contains(animationCurveRelationship, "Y") {
				curveNodesMap[animationCurveID].curves["y"] = animationCurve
			} else if strings.Contains(animationCurveRelationship, "Z") {
				curveNodesMap[animationCurveID].curves["z"] = animationCurve
			} else {
				_, ok := curveNodesMap[animationCurveID]
				if match, _ := regexp.Match("d|DeformPercent", []byte(animationCurveRelationship)); ok && match {
					curveNodesMap[animationCurveID].curves["morph"] = animationCurve
				}
			}
		}
	}
}

type CurveNode struct {
	ID       IDType
	AttrName string
	props    map[string]*Property

	modelName       string
	morphName       string
	initialPosition floatgeom.Point3
	initialRotation floatgeom.Point3
	initialScale    floatgeom.Point3
	T               *CurveNode
	R               *CurveNode
	S               *CurveNode
	transform       mgl64.Mat4
	preRotations    *[3]float64
	postRotations   *[3]float64
	curves          map[string]AnimationCurve

	DeformPercent float64
}

var (
	sanitizeRe    = regexp.MustCompile("/s")
	sanitizeEnclo = regexp.MustCompile("[(.*)]")
)

// sanitizeNodeName was a method on propertybinding that would: Replaces spaces with underscores and removes unsupported characters from
//  * node names, to ensure compatibility with parseTrackName().
func sanitizeNodeName(nodeName string) string {
	nodeName = sanitizeRe.ReplaceAllString(nodeName, "_")
	return sanitizeEnclo.ReplaceAllString(nodeName, "$1")
}

// parse nodes in FBXTree.Objects.AnimationLayer. Each layers holds references
// to various AnimationCurveNodes and is referenced by an AnimationStack node
// note: theoretically a stack can have multiple layers, however in practice there always seems to be one per stack
func (l *Loader) parseAnimationLayers(curveNodesMap map[IDType]CurveNode) map[IDType][]CurveNode {
	rawLayers := l.tree.Objects["AnimationLayer"]
	layersMap := make(map[IDType][]CurveNode)
	for _, node := range rawLayers {
		if connSet, ok := l.connections[node.ID]; ok {
			layerCurveNodes := make([]CurveNode, len(connSet.children))
			// all the animationCurveNodes used in the layer
			for i, child := range connSet.children {
				if curveNode, ok := curveNodesMap[child.ID]; ok {
					// check that the curves are defined for at least one axis, otherwise ignore the curveNode
					_, ok := curveNode.curves["x"]
					_, ok2 := curveNode.curves["y"]
					_, ok3 := curveNode.curves["z"]
					if ok || ok2 || ok3 {
						var modelID IDType
						for i := len(connSet.parents) - 1; i >= 0; i-- {
							parent := connSet.parents[i]
							if parent.Property != "" {
								modelID = parent.ID
								break
							}
						}
						if modelID == "" {
							break
						}
						rawModel := l.tree.Objects["Model"][modelID]
						node := CurveNode{
							modelName:       sanitizeNodeName(rawModel.attrName),
							initialPosition: floatgeom.Point3{0, 0, 0},
							initialRotation: floatgeom.Point3{0, 0, 0},
							initialScale:    floatgeom.Point3{1, 1, 1},
							transform:       l.getModelAnimTransform(rawModel),
						}
						// if the animated model is pre rotated, we'll have to apply the pre rotations to every
						// animation value as well
						if v, ok := rawModel.props["PreRotation"]; ok {
							v2 := v.Payload.([3]float64)
							node.preRotations = &v2
						}
						if v, ok := rawModel.props["PostRotation"]; ok {
							v2 := v.Payload.([3]float64)
							node.postRotations = &v2
						}
						// This is questionable!
						layerCurveNodes[i] = node
						layerCurveNodes[i].props[curveNode.AttrName] = &Property{Payload: curveNode}
					} else if _, ok := curveNode.curves["morph"]; ok {
						var deformerID IDType
						for i := len(connSet.parents) - 1; i >= 0; i-- {
							parent := connSet.parents[i]
							if parent.Property != "" {
								deformerID = parent.ID
								break
							}
						}
						morpherID := l.connections[deformerID].parents[0].ID
						geoID := l.connections[morpherID].parents[0].ID
						// assuming geometry is not used in more than one model
						modelID := l.connections[geoID].parents[0].ID
						rawModel := l.tree.Objects["Model"][modelID]
						var node = CurveNode{
							modelName: sanitizeNodeName(rawModel.attrName),
							morphName: l.tree.Objects["Deformer"][deformerID].attrName,
						}
						layerCurveNodes[i] = node
						layerCurveNodes[i].props[curveNode.AttrName] = &Property{Payload: curveNode}
					}
				}
			}
			layersMap[node.ID] = layerCurveNodes
		}
	}
	return layersMap
}

// All fields are optional
type TransformData struct {
	eulerOrder     *EulerOrder
	translation    *floatgeom.Point3
	rotationOffset *floatgeom.Point3
	rotation       *floatgeom.Point3
	preRotation    *floatgeom.Point3
	postRotation   *floatgeom.Point3
	scale          *floatgeom.Point3
}

func (l *Loader) getModelAnimTransform(modelNode *Node) mgl64.Mat4 {
	var td TransformData
	if v, ok := modelNode.props["RotationOrder"]; ok {
		eo, err := strconv.Atoi(v.Payload.(string))
		if err != nil || eo >= int(LastEulerOrder) || eo < 0 {
			fmt.Println("Error decoding euler rotation order")
		} else {
			eo2 := EulerOrder(eo)
			td.eulerOrder = &eo2
		}
	}
	if v, ok := modelNode.props["Lcl_Translation"]; ok {
		v2 := v.Payload.(floatgeom.Point3)
		td.translation = &v2
	}
	if v, ok := modelNode.props["RotationOffset"]; ok {
		v2 := v.Payload.(floatgeom.Point3)
		td.rotationOffset = &v2
	}
	if v, ok := modelNode.props["Lcl_Rotation"]; ok {
		v2 := v.Payload.(floatgeom.Point3)
		td.rotation = &v2
	}
	if v, ok := modelNode.props["PreRotation"]; ok {
		v2 := v.Payload.(floatgeom.Point3)
		td.preRotation = &v2
	}
	if v, ok := modelNode.props["PostRotation"]; ok {
		v2 := v.Payload.(floatgeom.Point3)
		td.postRotation = &v2
	}
	if v, ok := modelNode.props["Lcl_Scaling"]; ok {
		v2 := v.Payload.(floatgeom.Point3)
		td.scale = &v2
	}
	return generateTransform(td)
}

type Clip struct {
	name  string
	layer []CurveNode
}

// parse nodes in FBXTree.Objects.AnimationStack. These are the top level node in the animation
// hierarchy. Each Stack node will be used to create a THREE.AnimationClip
func (l *Loader) parseAnimStacks(layersMap map[IDType][]CurveNode) map[IDType]Clip {
	rawStacks := l.tree.Objects["AnimationStack"]
	// connect the stacks (clips) up to the layers
	rawClips := make(map[IDType]Clip, len(rawStacks))
	for _, node := range rawStacks {
		children := l.connections[node.ID].children
		if len(children) > 1 {
			// it seems like stacks will always be associated with a single layer. But just in case there are files
			// where there are multiple layers per stack, we'll display a warning
			fmt.Println("THREE.FBXLoader: Encountered an animation stack with multiple layers, this is currently not supported. Ignoring subsequent layers.")
		}
		layer := layersMap[children[0].ID]
		rawClips[node.ID] = Clip{
			name:  rawStacks[node.ID].attrName,
			layer: layer,
		}
	}
	return rawClips
}

func (l *Loader) addClip(clip Clip) Animation {

	fmt.Println("Len of tracks", len(clip.layer))

	tracks := make([]KeyframeTrack, 0, len(clip.layer))
	for _, rawTracks := range clip.layer {
		tracks = append(tracks, l.generateTracks(rawTracks)...)
	}
	fmt.Println("Len of completed tracks", len(tracks))

	// ??
	return NewAnimationClip(clip.name, -1, tracks)
}

func (l *Loader) generateTracks(rawTracks CurveNode) []KeyframeTrack {

	fmt.Println("Raw tracks", rawTracks)

	tracks := []KeyframeTrack{}
	initialPosition := floatgeom.Point3{}
	initialRotation := Euler{}
	initialScale := floatgeom.Point3{}
	if !IsZeroMat(rawTracks.transform) {
		initialPosition, initialRotation, initialScale = decomposeMat(rawTracks.transform)
	}
	if rawTracks.T != nil && len(rawTracks.T.curves) > 0 {
		positionTrack := l.generateVectorTrack(rawTracks.modelName, rawTracks.T.curves, initialPosition, "position")
		tracks = append(tracks, positionTrack)
	}
	if rawTracks.R != nil && len(rawTracks.R.curves) > 0 {
		rotationTrack := l.generateRotationTrack(rawTracks.modelName, rawTracks.R.curves, initialRotation, rawTracks.preRotations, rawTracks.postRotations)
		tracks = append(tracks, rotationTrack)
	}
	if rawTracks.S != nil && len(rawTracks.S.curves) > 0 {
		scaleTrack := l.generateVectorTrack(rawTracks.modelName, rawTracks.S.curves, initialScale, "scale")
		tracks = append(tracks, scaleTrack)
	}
	if rawTracks.DeformPercent != 0.0 {
		fmt.Println("Morph tracks not supported")
		//morphTrack := l.generateMorphTrack(rawTracks)
		//tracks = append(tracks, morphTrack)
	}
	return tracks
}

func (l *Loader) generateVectorTrack(modelName string, curves map[string]AnimationCurve, initialValue floatgeom.Point3, typ string) KeyframeTrack {
	times := l.getTimesForAllAxes(curves)
	values := l.getKeyframeTrackValues(times, curves, initialValue)
	return VectorKeyframeTrack(modelName+"."+typ, times, values)
}

func (l *Loader) generateRotationTrack(modelName string, curves map[string]AnimationCurve, initialValue Euler, preRotations, postRotations *[3]float64) KeyframeTrack {
	if c, ok := curves["x"]; ok {
		c = l.interpolateRotations(c)
		for i, val := range c.Values {
			c.Values[i] = alg.DegToRad * val
		}
	}
	if c, ok := curves["y"]; ok {
		c = l.interpolateRotations(c)
		for i, val := range c.Values {
			c.Values[i] = alg.DegToRad * val
		}
	}
	if c, ok := curves["z"]; ok {
		c = l.interpolateRotations(c)
		for i, val := range c.Values {
			c.Values[i] = alg.DegToRad * val
		}
	}
	times := l.getTimesForAllAxes(curves)
	values := l.getKeyframeTrackValues(times, curves, initialValue.Point3)
	var preRot *floatgeom.Point4
	if preRotations != nil {
		preRotations[0] *= alg.DegToRad
		preRotations[1] *= alg.DegToRad
		preRotations[2] *= alg.DegToRad
		eul := Euler{
			floatgeom.Point3(*preRotations),
			ZYXOrder,
		}
		q := eul.ToQuaternion()
		preRot = &q
	}
	var postRot *floatgeom.Point4
	if postRotations != nil {
		postRotations[0] *= alg.DegToRad
		postRotations[1] *= alg.DegToRad
		postRotations[2] *= alg.DegToRad
		eul := Euler{
			floatgeom.Point3(*postRotations),
			ZYXOrder,
		}
		q := eul.ToQuaternion().Inverse()
		postRot = &q
	}

	keyFrameVals := []float64{}
	for i := 0; i < len(values); i += 3 {
		euler := Euler{
			floatgeom.Point3{values[i], values[i+1], values[i+2]},
			ZYXOrder,
		}
		quaternion := euler.ToQuaternion()
		if preRot != nil {
			quaternion = preRot.MulQuat(quaternion)
		}
		if postRot != nil {
			quaternion = quaternion.MulQuat(*postRot)
		}
		keyFrameVals = append(keyFrameVals, quaternion[0], quaternion[1], quaternion[2], quaternion[3])
	}
	return QuaternionKeyframeTrack(modelName+".quaternion", times, keyFrameVals)
}

/*
func (l *Loader) generateMorphTrack(rawTracks CurveNode) KeyframeTrack {
	curves := rawTracks.DeformPercent.curves["morph"]
	values := make([]float64, len(curves.values))
	for i, val := range curves.values {
		values[i] = val / 100
	}
	found, err := SearchModelsByName(l.sceneGraph, rawTracks.modelName)
	if err != nil {
		fmt.Println("Unable to find model with name", rawTracks.modelName)
		return KeyframeTrack{}
	}
	morphNum := found.morphTargetDictionary[rawTracks.morphName]
	return NumberKeyframeTrack(rawTracks.modelName+".morphTargetInfluences["+morphNum+"]", curves.Times, values)
}
*/

// For all animated objects, times are defined separately for each axis
// Here we'll combine the times into one sorted array without duplicates
func (l *Loader) getTimesForAllAxes(curves map[string]AnimationCurve) []float64 {
	var times = []float64{}
	// first join together the times for each axis, if defined
	if c, ok := curves["x"]; ok {
		times = append(times, c.Times...)
	}
	if c, ok := curves["y"]; ok {
		times = append(times, c.Times...)
	}
	if c, ok := curves["z"]; ok {
		times = append(times, c.Times...)
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
			times = append(times[:i], times[i+1:]...)
			continue
		}
		i++
	}
	return times
}

func (l *Loader) getKeyframeTrackValues(times []float64, curves map[string]AnimationCurve, initialValue floatgeom.Point3) []float64 {
	prevValue := initialValue
	values := []float64{}
	xIndex := -1
	yIndex := -1
	zIndex := -1
	for _, time := range times {
		if _, ok := curves["x"]; ok {
			for i, t2 := range curves["x"].Times {
				if t2 == time {
					xIndex = i
					break
				}
			}
		}
		if _, ok := curves["y"]; ok {
			for i, t2 := range curves["y"].Times {
				if t2 == time {
					yIndex = i
					break
				}
			}
		}
		if _, ok := curves["z"]; ok {
			for i, t2 := range curves["z"].Times {
				if t2 == time {
					zIndex = i
					break
				}
			}
		}
		// if there is an x value defined for this frame, use that
		if xIndex != -1 {
			xValue := curves["x"].Values[xIndex]
			values = append(values, xValue)
			prevValue[0] = xValue
		} else {
			// otherwise use the x value from the previous frame
			values = append(values, prevValue[0])
		}
		if yIndex != -1 {
			yValue := curves["y"].Values[yIndex]
			values = append(values, yValue)
			prevValue[1] = yValue
		} else {
			values = append(values, prevValue[1])
		}
		if zIndex != -1 {
			zValue := curves["z"].Values[zIndex]
			values = append(values, zValue)
			prevValue[2] = zValue
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
	for i := 1; i < len(curve.Values); i++ {
		initialValue := curve.Values[i-1]
		valuesSpan := curve.Values[i] - initialValue
		absoluteSpan := math.Abs(valuesSpan)
		if absoluteSpan >= 180 {
			numSubIntervals := absoluteSpan / 180
			step := valuesSpan / numSubIntervals
			nextValue := initialValue + step
			initialTime := curve.Times[i-1]
			timeSpan := curve.Times[i] - initialTime
			interval := timeSpan / numSubIntervals
			nextTime := initialTime + interval
			interpolatedTimes := []float64{}
			interpolatedValues := []float64{}
			for nextTime < curve.Times[i] {
				interpolatedTimes = append(interpolatedTimes, nextTime)
				nextTime += interval
				interpolatedValues = append(interpolatedValues, nextValue)
				nextValue += step
			}
			curve.Times = append(curve.Times[:i], append(interpolatedTimes, curve.Times[i:]...)...)
			curve.Values = append(curve.Values[:i], append(interpolatedValues, curve.Values[i:]...)...)
		}
	}
	return curve
}

type Animation struct {
	Name     string
	Duration float64
	Tracks   []KeyframeTrack
}

type AnimationCurve struct {
	ID     IDType
	Times  []float64
	Values []float64
}

func NewAnimationClip(name string, duration float64, tracks []KeyframeTrack) Animation {
	if duration < 0 {
		//resetDuration function
		for _, t := range tracks {
			duration = math.Max(duration, t.Times[len(t.Times)-1])
		}
	}
	return Animation{
		Name:     name,
		Duration: duration,
		Tracks:   tracks,
	}
}

// Copy is a NOP right now because Animation doesn't have any fields yet
func (a *Animation) Copy() *Animation {
	return &Animation{}
}

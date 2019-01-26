package threefbx

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/oakmound/oak/alg/floatgeom"
	"github.com/oakmound/oak/dlog"
)

//Notes:
// 		* geometries with mutliple models that have different transforms may break this.
//		* if Vertex has more than 4 skinning weights we throw out the extra

// TODO: check assumptions
//Assumptions: parsing a geo probably returns a geometry.

// Geometry tries to replace the need for THREE.BufferGeometry
type Geometry struct {
	Name     string
	Position []floatgeom.Point3
	Color    []Color

	SkinIndex    [][4]uint16
	SkinWeight   [][4]float64
	FBX_Deformer *Skeleton

	Normal []floatgeom.Point3
	Uvs    [][]floatgeom.Point2

	Groups []Group

	MorphAttributes MorphAttributes
}

type MorphAttributes struct {
	position []floatgeom.Point3
	normal   []floatgeom.Point3
}

func NewGeometry() Geometry {
	g := Geometry{}
	g.Groups = make([]Group, 0)
	g.Uvs = make([][]floatgeom.Point2, 0)
	g.Position = make([]floatgeom.Point3, 0)
	g.Color = make([]Color, 0)

	return g
}

type UVRaw [2]float32

type Group struct {
	start, count int
	value        int32
}

type WeightEntry struct { //TODO: see if we can find a better way that doesnt use this (we could do array or split weightable itself out to two props.)
	ID     uint16
	Weight float64
}

// floatsToVertex3s is a helper function to convert flat float arrays into vertex3s
func floatsToVertex3s(arr []float64) ([]floatgeom.Point3, error) {
	if len(arr)%3 != 0 {
		return nil, errors.New("Something went wrong parsing an array of floats to vertices as it was not divisible by 3")
	}
	output := make([]floatgeom.Point3, len(arr)/3)
	for i := 0; i < len(arr)/3; i++ {
		output[i] = floatgeom.Point3{float64(arr[i*3]), float64(arr[i*3+1]), float64(arr[i*3+2])}
	}
	return output, nil
}
func floatsToVertex4s(arr []float32) ([]floatgeom.Point4, error) {
	if len(arr)%4 != 0 {
		return nil, errors.New("Something went wrong parsing an array of floats to vertices as it was not divisible by 3")
	}
	output := make([]floatgeom.Point4, len(arr)/3)
	for i := 0; i < len(arr)/4; i++ {
		output[i] = floatgeom.Point4{float64(arr[i*4]), float64(arr[i*4+1]), float64(arr[i*4+2]), float64(arr[i*4+3])}
	}
	return output, nil
}

// AddGroup was a THREE.js thing start of conversion is here it seems to store a range for which a value is the same
func (g *Geometry) AddGroup(rangeStart, count int, groupValue int32) {
	g.Groups = append(g.Groups, Group{rangeStart, count, groupValue})
}

// parseGeometry converted from parse in the geometry section of the original code
// parse Geometry data from FBXTree and return map of BufferGeometries
// Parse nodes in FBXTree.Objects.Geometry
func (l *Loader) parseGeometry(skeletons map[IDType]Skeleton, morphTargets map[IDType]MorphTarget) (map[IDType]Geometry, error) {
	geometryMap := make(map[IDType]Geometry)
	if geoNodes, ok := l.tree.Objects["Geometry"]; ok {
		for _, node := range geoNodes {
			nodeID := node.ID
			relationships := l.connections[nodeID]
			geo, err := l.parseGeometrySingle(relationships, geoNodes[nodeID], skeletons, morphTargets)
			if err != nil {
				return nil, err
			}
			fmt.Println("Geo normality:", len(geo.Normal))
			geometryMap[nodeID] = geo
		}
	}
	return geometryMap, nil
}

// parseGeometrySingle parses a single node in FBXTree.Objects.Geometry //updated name due to collisions
func (l *Loader) parseGeometrySingle(relationships ConnectionSet, geoNode *Node, skeletons map[IDType]Skeleton, morphTargets map[IDType]MorphTarget) (Geometry, error) {
	switch geoNode.attrType {
	case "Mesh":
		return l.parseMeshGeometry(relationships, *geoNode, skeletons, morphTargets)
	case "NurbsCurve":
		return l.parseNurbsGeometry(*geoNode)
	}

	return Geometry{}, errors.New("Unknown geometry type when parsing" + geoNode.attrType)
}

// parseMeshGeometry parses a single node mesh geometry in FBXTree.Objects.Geometry
func (l *Loader) parseMeshGeometry(relationships ConnectionSet, geoNode Node, skeletons map[IDType]Skeleton, morphTargets map[IDType]MorphTarget) (Geometry, error) {

	modelNodes := make([]*Node, len(relationships.parents))
	for i, parent := range relationships.parents {
		modelNodes[i] = l.tree.Objects["Model"][parent.ID]
	}

	// dont create if geometry has no models
	if len(modelNodes) == 0 {
		return Geometry{}, nil
	}

	var skeleton *Skeleton //TODO: whats this type
	for i := len(relationships.children) - 1; i >= 0; i-- {
		chID := relationships.children[i].ID
		if skel, ok := skeletons[chID]; ok {
			skeleton = &skel
			break
		}
	}
	var morphTarget *MorphTarget //TODO: whats this type
	for i := len(relationships.children) - 1; i >= 0; i-- {
		chID := relationships.children[i].ID
		if morp, ok := morphTargets[chID]; ok {
			morphTarget = &morp
			break
		}
	}
	// TODO: if there is more than one model associated with the geometry, AND the models have
	// different geometric transforms, then this will cause problems
	// if ( modelNodes.length > 1 ) { }
	// For now just assume one model and get the preRotations from that
	modelNode := modelNodes[0]
	transformData := TransformData{}
	if val, ok := modelNode.props["RotationOrder"]; ok {
		vp := (val.Payload.(EulerOrder))
		transformData.eulerOrder = &vp

	}
	if val, ok := modelNode.props["GeometricTranslation"]; ok {
		vp := (val.Payload.(floatgeom.Point3))
		transformData.translation = &vp
	}
	if val, ok := modelNode.props["GeometricRotation"]; ok {
		vp := (val.Payload.(floatgeom.Point3))
		transformData.rotation = &vp
	}
	if val, ok := modelNode.props["GeometricScaling"]; ok {
		vp := (val.Payload.(floatgeom.Point3))
		transformData.scale = &vp
	}
	transform := generateTransform(transformData) //TODO: see above about how this ordering might change
	return l.genGeometry(geoNode, skeleton, morphTarget, transform)

}

// genGeometry generates a THREE.BufferGeometry(ish) from a node in FBXTree.Objects.Geometry
func (l *Loader) genGeometry(geoNode Node, skeleton *Skeleton, morphTarget *MorphTarget, preTransform mgl64.Mat4) (Geometry, error) {
	geo := NewGeometry() //https://threejs.org/docs/#api/en/core/BufferGeometry
	geo.Name = geoNode.attrName

	geoInfo, err := l.parseGeoNode(geoNode, skeleton)
	if err != nil {
		return Geometry{}, err
	}

	//TODO: unroll buffers into its consituent slices and do away with the buffer construct
	buffers := l.genBuffers(geoInfo)

	positionAttribute, err := floatsToVertex3s(buffers.vertex) //https://threejs.org/docs/#api/en/core/BufferAttribute

	positionAttribute = applyBufferAttribute(preTransform, positionAttribute)

	geo.Position = positionAttribute
	if len(buffers.colors) > 0 {
		colors, err := floatsToVertex3s(buffers.colors)
		if err != nil {
			return Geometry{}, err
		}
		geo.Color = make([]Color, len(colors))
		for i, c := range colors {
			geo.Color[i] = Color{float32(c.X()), float32(c.Y()), float32(c.Z())}
		}
	}

	if skeleton != nil {
		geo.SkinIndex = make([][4]uint16, len(buffers.weightsIndices)/4)
		geo.SkinWeight = make([][4]float64, len(buffers.vertexWeights)/4)

		for i := 0; i < len(buffers.weightsIndices); i += 4 {
			geo.SkinIndex[i/4] = [4]uint16{buffers.weightsIndices[i], buffers.weightsIndices[i+1], buffers.weightsIndices[i+2], buffers.weightsIndices[i+3]}
		}
		for i := 0; i < len(buffers.vertexWeights); i += 4 {
			geo.SkinWeight[i/4] = [4]float64{buffers.vertexWeights[i], buffers.vertexWeights[i+1], buffers.vertexWeights[i+2], buffers.vertexWeights[i+3]}
		}
		geo.FBX_Deformer = skeleton
	}

	if len(buffers.normal) > 0 {
		normalAttribute, err := floatsToVertex3s(buffers.normal)
		if err != nil {
			return Geometry{}, err
		}
		normalMatrix := mgl64.Mat4Normal(preTransform)

		normalAttribute = applyBufferAttributeMat3(normalMatrix, normalAttribute)
		geo.Normal = normalAttribute
	}

	geo.Uvs = buffers.uvs //NOTE: pulled back from variadic array of uvs where they progress down uv -> uv1 -> uv2 and so on

	if geoInfo.material != nil {
		mat := *geoInfo.material
		if mat.mappingType != "AllSame" {
			// Convert the material indices of each vertex into rendering groups on the geometry.
			prevMaterialIndex := buffers.materialIndex[0]
			startIndex := 0
			for i, matIndex := range buffers.materialIndex {
				if matIndex != prevMaterialIndex {
					geo.AddGroup(startIndex, i-startIndex, prevMaterialIndex)
					prevMaterialIndex = matIndex
					startIndex = i
				}
			}
			if len(geo.Groups) > 0 { //add last group
				lastGroup := geo.Groups[len(geo.Groups)-1]
				lastIndex := lastGroup.start + lastGroup.count
				if lastIndex != len(buffers.materialIndex) {
					geo.AddGroup(lastIndex, len(buffers.materialIndex)-lastIndex, prevMaterialIndex)
				}
			}
			if len(geo.Groups) == 0 {
				geo.AddGroup(0, len(buffers.materialIndex), buffers.materialIndex[0])
			}
		}
	}

	err = l.addMorphTargets(&geo, geoNode, morphTarget, preTransform)

	return geo, nil
}

type GeoInfo struct {
	vertexPositions []float64 //Todo: parse this immediately as floatgeom.Point3
	vertexIndices   []int32
	color           *floatBuffer
	material        *intBuffer
	normal          *floatBuffer
	skeleton        *Skeleton
	weightTable     map[int32][]WeightEntry
	uv              []floatBuffer
}

func (l *Loader) parseGeoNode(geoNode Node, skeleton *Skeleton) (*GeoInfo, error) {
	geoInfo := &GeoInfo{}

	if v, ok := geoNode.props["Vertices"]; ok {
		geoInfo.vertexPositions = v.Payload.([]float64)
		fmt.Println("geoinfo.vertexPositions", geoInfo.vertexPositions)
	}

	if v, ok := geoNode.props["PolygonVertexIndex"]; ok {
		geoInfo.vertexIndices = v.Payload.([]int32)
	}

	if v, ok := geoNode.props["LayerElementColor"]; ok {
		v2 := parseVertexColors(v.Payload.(*Node))
		geoInfo.color = &v2
	}
	if v, ok := geoNode.props["LayerElementMaterial"]; ok {
		v2 := l.parseMaterialIndices(v.Payload.(*Node))
		geoInfo.material = &v2
	}
	if v, ok := geoNode.props["LayerElementNormal"]; ok {
		v2 := parseNormals(v.Payload.(*Node))
		geoInfo.normal = &v2
	}

	if uvList, ok := geoNode.props["LayerElementUV"]; ok {
		uvBuffs := make([]floatBuffer, 1)
		uvBuffs[0] = l.parseUVs(uvList.Payload.(*Node))
		geoInfo.uv = uvBuffs
	}

	if skeleton != nil {
		geoInfo.skeleton = skeleton
		wt := map[int32][]WeightEntry{}
		if len(skeleton.rawBones) == 0 {
			dlog.Warn("No bones found for skeleton", skeleton)
		}
		// loop over the bone's vertex indices and weights
		for i, rawb := range skeleton.rawBones {
			if len(rawb.Indices) == 0 {
				dlog.Warn("Bone without indicies found", rawb)
			}
			for j, rIndex := range rawb.Indices {
				if _, ok := wt[rIndex]; !ok {
					wt[rIndex] = []WeightEntry{}
				}
				wt[rIndex] = append(wt[rIndex], WeightEntry{uint16(i), rawb.Weights[j]})
			}
		}

		geoInfo.weightTable = wt
	} else {
		dlog.Warn("No skeleton found")
	}

	return geoInfo, nil
}

func (l *Loader) genBuffers(geoInfo *GeoInfo) gBuffers {
	buffers := gBuffers{}
	buffers.uvs = make([][]floatgeom.Point2, len(geoInfo.uv))
	polygonIndex := int32(0)
	faceLength := 0
	displayedWeightsWarning := false

	// tracking faces
	facePositionIndexes := []int32{}
	faceNormals := []float64{}
	faceColors := []float64{}

	faceWeights := []float64{}
	faceWeightIndices := []uint16{}
	var materialIndex int32 = -1

	faceUVs := make([][]float64, len(geoInfo.uv))
	for polygonVertexIndex, vertexIndex := range geoInfo.vertexIndices {
		// Face index and vertex index arrays are combined in a single array
		// A cube with quad faces looks like this:
		// PolygonVertexIndex: *24 {
		//  a: 0, 1, 3, -3, 2, 3, 5, -5, 4, 5, 7, -7, 6, 7, 1, -1, 1, 7, 5, -4, 6, 0, 2, -5
		//  }
		// Negative numbers mark the end of a face - first face here is 0, 1, 3, -3
		// to find index of last vertex bit shift the index: ^ - 1
		endOfFace := false
		if vertexIndex < 0 {
			vertexIndex = vertexIndex ^ -1
			endOfFace = true
		}
		weightIndices := []uint16{}
		weights := []float64{}
		facePositionIndexes = append(facePositionIndexes, vertexIndex*3, vertexIndex*3+1, vertexIndex*3+2)
		if geoInfo.color != nil {
			data := geoInfo.color.getData(polygonVertexIndex, polygonIndex, vertexIndex)
			faceColors = append(faceColors, data[0], data[1], data[2])
		}
		if geoInfo.skeleton != nil {
			if ws, ok := geoInfo.weightTable[vertexIndex]; ok {
				for _, w := range ws {
					weights = append(weights, w.Weight)
					weightIndices = append(weightIndices, w.ID)
				}
			}
			if len(weights) > 4 {
				if !displayedWeightsWarning {
					fmt.Println("Vertex has more than 4 skinning weights assigned to vertex. Deleting additional weights.")
					displayedWeightsWarning = true
				}

				//--------------------------------------------------------------------------------------------------------------
				//--------------------------------------------------------------------------------------------------------------
				// TODO: we should talk about this because I am confused about the for each on the Weight as we instantiate it as a [4]int
				// Given that we know js can foreach on an object and interact with its props that makes sense but not incontext of the weight being [4]int
				wIndex := []uint16{0, 0, 0, 0}
				Weight := []float64{0, 0, 0, 0}
				for currentIndex, currentWeight := range weights {
					for comparedWeightIndex, comparedWeight := range Weight {
						if currentWeight > comparedWeight {
							Weight[comparedWeightIndex] = currentWeight
							currentWeight = comparedWeight
							tmp := wIndex[comparedWeightIndex]
							wIndex[comparedWeightIndex] = uint16(currentIndex)
							currentIndex = int(tmp)
						}
					}
				}
				weightIndices = wIndex
				weights = Weight
				//--------------------------------------------------------------------------------------------------------------
				//--------------------------------------------------------------------------------------------------------------
			}
			// if the weight array is shorter than 4 pad with 0s
			for len(weights) < 4 {
				weights = append(weights, 0)
				weightIndices = append(weightIndices, 0)
			}

			for i := 0; i < 4; i++ {
				faceWeights = append(faceWeights, weights[i])
				faceWeightIndices = append(faceWeightIndices, weightIndices[i])
			}
		}

		if geoInfo.normal != nil {
			data := geoInfo.normal.getData(polygonVertexIndex, polygonIndex, vertexIndex)
			faceNormals = append(faceNormals, data[0], data[1], data[2])
		}
		if geoInfo.material != nil && geoInfo.material.mappingType != "AllSame" {
			materialIndex = geoInfo.material.getData(polygonVertexIndex, polygonIndex, vertexIndex)[0]
		}
		if geoInfo.uv != nil {

			for i, uv := range geoInfo.uv {
				data := uv.getData(polygonVertexIndex, polygonIndex, vertexIndex)
				faceUVs[i] = append(faceUVs[i], data[0], data[1])
			}
		}
		faceLength++
		if endOfFace {
			l.genFace(&buffers, geoInfo, facePositionIndexes, materialIndex, faceNormals, faceColors, faceUVs, faceWeights, faceWeightIndices, faceLength)
			polygonIndex++
			faceLength = 0
			// reset arrays for the next face
			facePositionIndexes = []int32{}
			faceNormals = []float64{}
			faceColors = []float64{}
			faceWeights = []float64{}
			faceWeightIndices = []uint16{}
			faceUVs = make([][]float64, len(geoInfo.uv))
		}
	}
	fmt.Println("Returning Buffers with a len ", len(buffers.normal), "of normals")
	return buffers
}

type gBuffers struct {
	vertex         []float64
	normal         []float64
	colors         []float64
	uvs            [][]floatgeom.Point2
	materialIndex  []int32
	vertexWeights  []float64
	weightsIndices []uint16
}

// genFace generates data for a single face in a geometry. If the face is a quad then split it into 2 tris
func (l *Loader) genFace(buffers *gBuffers, geoInfo *GeoInfo, facePositionIndexes []int32, materialIndex int32,
	faceNormals []float64, faceColors []float64, faceUVs [][]float64, faceWeights []float64, faceWeightIndices []uint16, faceLength int) {

	// Current understanding of this code:
	// It effectively triangulates the faces
	for i := 2; i < faceLength; i++ {
		buffers.vertex = append(buffers.vertex, genFaceVertex(geoInfo.vertexPositions, facePositionIndexes, i)...)
		if geoInfo.skeleton != nil {
			buffers.vertexWeights = append(buffers.vertexWeights, genFloatFaceArray(4, faceWeights, i)...)
			buffers.weightsIndices = append(buffers.weightsIndices, genUint16FaceArray(4, faceWeightIndices, i)...)
		}
		if geoInfo.color != nil {
			buffers.colors = append(buffers.colors, genFloatFaceArray(3, faceColors, i)...)
		}
		if geoInfo.material != nil && geoInfo.material.mappingType != "AllSame" {
			buffers.materialIndex = append(buffers.materialIndex, materialIndex) //TODO: This seems wrong as it checks that materials are NOT all the same...
			buffers.materialIndex = append(buffers.materialIndex, materialIndex)
			buffers.materialIndex = append(buffers.materialIndex, materialIndex)
		}
		if geoInfo.normal != nil {
			buffers.normal = append(buffers.normal, genFloatFaceArray(3, faceNormals, i)...)
		}

		if geoInfo.uv != nil {

			for i, u := range faceUVs {
				fmt.Println("i, len:", i, len(u))
			}
			for j := range geoInfo.uv {
				fmt.Println("geoinfo UV", len(faceUVs[j]), "face", i)
				fmt.Println("norms versus uvs ", len(faceNormals), " and, ", len(faceUVs[j]), " with i of ")
				arr := genFloatFaceArray(2, faceUVs[j], i)
				for k := 0; k < len(arr); k += 2 {
					buffers.uvs[j] = append(buffers.uvs[j], floatgeom.Point2{arr[k], arr[k+1]})
				}
			}
		}
	}
}

// genFaceVertex simplifies a portion of genface to be less wordy in creation of the vertices for the face
func genFaceVertex(vPos []float64, posIdxs []int32, idx int) []float64 {
	return []float64{
		vPos[posIdxs[0]], vPos[posIdxs[1]], vPos[posIdxs[2]],
		vPos[posIdxs[(idx-1)*3+0]], vPos[posIdxs[(idx-1)*3+1]], vPos[posIdxs[(idx-1)*3+2]],
		vPos[posIdxs[idx*3+0]], vPos[posIdxs[idx*3+1]], vPos[posIdxs[idx*3+2]],
	}
}

func genUint16FaceArray(size int, sourceArr []uint16, idx int) []uint16 {
	out := make([]uint16, size*3)
	for j := 0; j < size; j++ {
		out[j] = sourceArr[j]
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < size; j++ {
			out[(i+1)*size+j] = sourceArr[j+((idx-1+i)*size)]
		}
	}
	return out
}

func genIntFaceArray(size int, sourceArr []int, idx int) []int {
	out := make([]int, size*3)
	for j := 0; j < size; j++ {
		out[j] = sourceArr[j]
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < size; j++ {
			out[(i+1)*size+j] = sourceArr[j+((idx-1+i)*size)]
		}
	}
	return out
}

func genFloatFaceArray(size int, sourceArr []float64, idx int) []float64 {
	out := make([]float64, size*3)
	for j := 0; j < size; j++ {
		out[j] = sourceArr[j]
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < size; j++ {
			out[(i+1)*size+j] = sourceArr[j+((idx-1+i)*size)]
		}
	}
	return out
}

type FaceBuffer struct {
	vertex         floatgeom.Point3
	color          Color
	weightsIndices []int
	vertexWeights  []float64
	normal         floatgeom.Point3
	materialIndex  int
}

func (l *Loader) addMorphTargets(parentGeo *Geometry, parentGeoNode Node, morphTarget *MorphTarget, preTransform mgl64.Mat4) error {
	if morphTarget == nil {
		return nil
	}
	parentGeo.MorphAttributes.position = []floatgeom.Point3{}
	parentGeo.MorphAttributes.normal = []floatgeom.Point3{}
	for _, t := range morphTarget.RawTargets {
		morphGeoNode := l.tree.Objects["Geometry"][t.geoID]
		if morphGeoNode != nil {
			err := l.genMorphGeometry(parentGeo, parentGeoNode, *morphGeoNode, preTransform)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// a morph geometry node is similar to a standard  node, and the node is also contained
// in FBXTree.Objects.Geometry, however it can only have attributes for position, normal
// and a special attribute Index defining which vertices of the original geometry are affected
// Normal and position attributes only have data for the vertices that are affected by the morph
func (l *Loader) genMorphGeometry(parentGeo *Geometry, parentGeoNode, morphGeoNode Node, preTransform mgl64.Mat4) error {
	//morphGeo := THREE.BufferGeometry() //TODO: figure out type
	//if morphGeoNode.attrName != "" {
	//	morphGeo.name = morphGeoNode.attrName
	//}

	// make a copy of the parent's vertex positions
	verts := parentGeoNode.props["Vertices"].Payload.([]float64)
	vertexPositions := make([]float64, len(verts))
	copy(vertexPositions, verts)
	morphPositions := morphGeoNode.props["Vertices"].Payload.([]float64)
	indices := morphGeoNode.props["Indexes"].Payload.([]int)
	for i := 0; i < len(indices); i++ {
		morphIndex := indices[i] * 3
		// FBX format uses blend shapes rather than morph targets. This can be converted
		// by additively combining the blend shape positions with the original geometry's positions
		vertexPositions[morphIndex] += morphPositions[i*3]
		vertexPositions[morphIndex+1] += morphPositions[i*3+1]
		vertexPositions[morphIndex+2] += morphPositions[i*3+2]
	}

	morphGeoInfo := &GeoInfo{
		vertexIndices:   parentGeoNode.props["PolygonVertexIndex"].Payload.([]int32),
		vertexPositions: vertexPositions,
	}
	morphBuffers := l.genBuffers(morphGeoInfo)

	positionAttribute, err := floatsToVertex3s(morphBuffers.vertex)
	if err != nil {
		return err
	}
	//positionAttribute.name = morphGeoNode.attrName
	positionAttribute = applyBufferAttribute(preTransform, positionAttribute)

	parentGeo.MorphAttributes.position = append(parentGeo.MorphAttributes.position, positionAttribute...)
	return nil
}

// Parse normal from FBXTree.Objects.Geometry.LayerElementNormal if it exists
func parseNormals(n *Node) floatBuffer {
	mappingType := n.props["MappingInformationType"].Payload.(string)
	referenceType := n.props["ReferenceInformationType"].Payload.(string)
	indexBuffer := []int32{}
	if referenceType == "IndexToDirect" {
		if idx, ok := n.props["NormalIndex"]; ok {
			indexBuffer = idx.Payload.([]int32)
		} else if idx, ok2 := n.props["NormalsIndex"]; ok2 {
			indexBuffer = idx.Payload.([]int32)
		}
	}
	return floatBuffer{
		bufferDefinition: bufferDefinition{
			dataSize:      3,
			indices:       indexBuffer,
			mappingType:   mappingType,
			referenceType: referenceType,
		},
		buffer: n.props["Normals"].Payload.([]float64),
	}
}

// Parse UVs from FBXTree.Objects.Geometry.LayerElementUV if it exists
func (l *Loader) parseUVs(n *Node) floatBuffer {
	mappingType := n.props["MappingInformationType"].Payload.(string)
	referenceType := n.props["ReferenceInformationType"].Payload.(string)
	indexBuffer := []int32{}
	if referenceType == "IndexToDirect" {
		indexBuffer = n.props["UVIndex"].Payload.([]int32)
	}
	return floatBuffer{
		bufferDefinition: bufferDefinition{
			dataSize:      2,
			indices:       indexBuffer,
			mappingType:   mappingType,
			referenceType: referenceType,
		},
		buffer: n.props["UV"].Payload.([]float64),
	}
}

// Parse Vertex Colors from FBXTree.Objects.Geometry.LayerElementColor if it exists
func parseVertexColors(n *Node) floatBuffer {
	mappingType := n.props["MappingInformationType"].Payload.(string)
	referenceType := n.props["ReferenceInformationType"].Payload.(string)
	indexBuffer := []int32{}
	if referenceType == "IndexToDirect" {
		ci := n.props["ColorIndex"].Payload.(map[string]Property)
		n2 := ci[""].Payload.(*Node)
		prop := n2.propertyList[0]
		indexBuffer = prop.Payload.([]int32)
	}
	return floatBuffer{
		bufferDefinition: bufferDefinition{
			dataSize:      4,
			indices:       indexBuffer,
			mappingType:   mappingType,
			referenceType: referenceType,
		},
		buffer: n.props["Colors"].Payload.([]float64),
	}
}

// Parse mapping and material data in FBXTree.Objects.Geometry.LayerElementMaterial if it exists
func (l *Loader) parseMaterialIndices(node *Node) intBuffer {
	n := *node
	mappingType := n.props["MappingInformationType"].Payload.(string)
	referenceType := n.props["ReferenceInformationType"].Payload.(string)

	if mappingType == MappingNone {
		return intBuffer{
			bufferDefinition: bufferDefinition{
				dataSize:      1,
				indices:       []int32{0},
				mappingType:   MappingAllSame,
				referenceType: referenceType,
			},
			buffer: []int32{0},
		}
	}

	materialIndexBuffer := n.props["Materials"].Payload.([]int32)
	// Since materials are stored as indices, there's a bit of a mismatch between FBX and what
	// we expect.So we create an intermediate buffer that points to the index in the buffer,
	// for conforming with the other functions we've written for other data.
	materialIndices := make([]int32, len(materialIndexBuffer))
	for i := 0; i < len(materialIndexBuffer); i++ {
		materialIndices[i] = int32(i)
	}
	return intBuffer{
		bufferDefinition: bufferDefinition{
			dataSize:      1,
			indices:       materialIndices,
			mappingType:   mappingType,
			referenceType: referenceType,
		},
		buffer: materialIndexBuffer,
	}
}

func (l *Loader) parseNurbsGeometry(geoNode Node) (Geometry, error) {
	orderStr, ok := geoNode.props["order"]
	if !ok {
		return NewGeometry(), errors.New("No order prop on geometry") // FIgure out how to do this fail state as it used to return new THREE.BufferGeometry();
	}
	order, err := strconv.Atoi(orderStr.Payload.(string))
	if err != nil {
		return NewGeometry(), err // FIgure out how to do this fail state as it used to return new THREE.BufferGeometry();

	}

	degree := order - 1
	knots := geoNode.props["KnotVector"].Payload.([]float64)

	pointsValues := geoNode.props["Points"]
	controlPoints, err := floatsToVertex4s(pointsValues.Payload.([]float32))
	if err != nil {
		return NewGeometry(), err
	}

	var startKnot, endKnot int
	form, ok := geoNode.props["Form"].Payload.(string)
	if ok {
		if form == "Closed" {
			controlPoints = append(controlPoints, controlPoints[0])
		} else if form == "Periodic" {
			startKnot = degree
			endKnot = len(knots) - 1 - startKnot
			for i := 1; i <= degree; i++ {
				controlPoints = append(controlPoints, controlPoints[i])
			}
		}
	}

	curve := NewNurbsCurve(degree, knots, controlPoints, startKnot, endKnot)
	vertLen := float64(len(controlPoints) * 7)
	vertValues := make([]floatgeom.Point3, len(controlPoints)*7)
	for i := 0.0; i < vertLen; i++ {
		vertValues[int(i)] = curve.getPoint(i / vertLen)
	}
	g := NewGeometry()
	g.Position = vertValues
	return g, err
}

func floatSliceToPoint3(fs []float64) ([]floatgeom.Point3, error) {
	if len(fs)%3 != 0 {
		return nil, errors.New("input floats not divisible by 3")
	}
	out := make([]floatgeom.Point3, len(fs)/3)
	for i := 0; i < len(fs); i += 3 {
		out[i/3] = floatgeom.Point3{
			fs[i], fs[i+1], fs[i+2],
		}
	}
	return out, nil
}

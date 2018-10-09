package threefbx

import (
	"fmt"
	"errors"
	"github.com/oakmound/oak/alg/floatgeom"
	"github.com/go-gl/mathgl/mgl64"
)

//Notes:
// 		* geometries with mutliple models that have different transforms may break this.
//		* if Vertex has more than 4 skinning weights we throw out the extra

// TODO: check assumptions
//Assumptions: parsing a geo probably returns a geometry.

// Geometry tries to replace the need for THREE.BufferGeometry
type Geometry struct {
	name string
	position []floatgeom.Point3 
	color []Color

	skinIndex [][4]uint16
	skinWeight [][4]float64
	FBX_Deformer *Skeleton

	normal []floatgeom.Point3
	uvs [][]float64

	groups []Group
}

type UV struct {} // ???

func NewGeometry() Geometry {
	g := Geometry{}
	g.groups = make([]Group, 0)
	g.uvs = make([][]float64, 0)
	g.position = make([]floatgeom.Point3, 0)
	g.color = make([]Color, 0)

	return g
}

type UVRaw [2]float32

type Group [3]int

type WeightEntry struct { //TODO: see if we can find a better way that doesnt use this (we could do array or split weightable itself out to two props.)
	ID uint16
	Weight float64
}

// floatsToVertex3s is a helper function to convert flat float arrays into vertex3s 
func floatsToVertex3s(arr []float64) ([]floatgeom.Point3, error){
	if len(arr) %3 != 0{
		return nil, errors.New("Something went wrong parsing an array of floats to vertices as it was not divisible by 3")
	}
	output := make([]floatgeom.Point3, len(arr)/3)
	for i := 0; i < len(arr)/3; i++{
		output[i] = floatgeom.Point3{float64(arr[i*3]),float64(arr[i*3+1]),float64(arr[i*3+2])}
	}
	return output, nil
}
func floatsToVertex4s(arr []float32) ([]floatgeom.Point4, error){
	if len(arr) %4 != 0{
		return nil, errors.New("Something went wrong parsing an array of floats to vertices as it was not divisible by 3")
	}
	output := make([]floatgeom.Point4, len(arr)/3)
	for i := 0; i < len(arr)/4; i++{
		output[i] = floatgeom.Point4{float64(arr[i*4]),float64(arr[i*4+1]),float64(arr[i*4+2]),float64(arr[i*4+3])}
	}
	return output, nil
}
// AddGroup was a THREE.js thing start of conversion is here it seems to store a range for which a value is the same
func (g *Geometry) AddGroup(rangeStart, count, groupValue int){
	g.groups = append(g.groups, Group{rangeStart, count, groupValue})
}




// parseGeometry converted from parse in the geometry section of the original code
// parse Geometry data from FBXTree and return map of BufferGeometries
// Parse nodes in FBXTree.Objects.Geometry
func (l *Loader) parseGeometry(skeletons map[int64]Skeleton,morphTargets map[int64]MorphTarget ) (map[int64]Geometry, error) {
	geometryMap := make(map[int64]Geometry)
	if geoNodes, ok := l.tree.Objects["Geometry"]; ok{
		for _, node := range geoNodes{
			nodeID := node.ID
			relationships := l.connections[nodeID]
			geo, err := l.parseGeometrySingle(relationships, geoNodes[nodeID], skeletons, morphTargets)
			if err != nil{
				return nil, err
			}
			geometryMap[nodeID] = geo
		}
	}
	return geometryMap, nil
}
// parseGeometrySingle parses a single node in FBXTree.Objects.Geometry //updated name due to collisions
func (l *Loader) parseGeometrySingle(relationships ConnectionSet, geoNode *Node, skeletons map[int64]Skeleton,morphTargets map[int64]MorphTarget ) (Geometry, error){
	switch geoNode.attrType{
	case "Mesh":
		return l.parseMeshGeometry(relationships, *geoNode, skeletons, morphTargets)
	case "NurbsCurve":
		return l.parseNurbsGeometry(*geoNode),nil
	}
	
	return Geometry{}, errors.New("Unknown geometry type when parsing" + geoNode.attrType)
}




// parseMeshGeometry parses a single node mesh geometry in FBXTree.Objects.Geometry
func (l *Loader) parseMeshGeometry( relationships ConnectionSet, geoNode Node,  skeletons map[int64]Skeleton,morphTargets map[int64]MorphTarget) (Geometry, error) {

	modelNodes := make([]*Node,len(relationships.parents))
	for i, parent := range relationships.parents{
		modelNodes[i] = l.tree.Objects["Model"][parent.ID]
	}

	// dont create if geometry has no models
	if len(modelNodes) ==0{
		return Geometry{}, nil
	}

	var skeleton *Skeleton	//TODO: whats this type
	for i := len(relationships.children)-1; i >= 0; i--{
		chID := relationships.children[i].ID
		if skel, ok := skeletons[chID] ; ok{
			skeleton = &skel
			break
		}
	}
	var morphTarget *MorphTarget //TODO: whats this type
	for i := len(relationships.children)-1; i >= 0; i--{
		chID := relationships.children[i].ID
		if morp, ok := morphTargets[chID] ; ok{
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
	if val, ok := modelNode.props["RotationOrder"]; ok{
		vp := (val.Payload().(EulerOrder))
		transformData.eulerOrder = &vp
		
	}
	if val, ok := modelNode.props["GeometricTranslation"]; ok{
		vp := (val.Payload().(floatgeom.Point3))
		transformData.translation = &vp
	}
	if val, ok := modelNode.props["GeometricRotation"]; ok{
		vp := (val.Payload().(floatgeom.Point3))
		transformData.rotation =&vp
	}
	if val, ok := modelNode.props["GeometricScaling"]; ok{
		vp := (val.Payload().(floatgeom.Point3))
		transformData.scale = &vp
	}
	transform := generateTransform(transformData) //TODO: see above about how this ordering might change
	return l.genGeometry(geoNode, skeleton, morphTarget, transform )

}

// genGeometry generates a THREE.BufferGeometry(ish) from a node in FBXTree.Objects.Geometry
func (l *Loader) genGeometry (geoNode Node, skeleton *Skeleton, morphTarget *MorphTarget, preTransform mgl64.Mat4) (Geometry, error) {
	geo := NewGeometry() //https://threejs.org/docs/#api/en/core/BufferGeometry
	geo.name = geoNode.attrName

	geoInfo, err := l.parseGeoNode(geoNode, skeleton)
	if err != nil {
		return Geometry{}, err
	}

	//TODO: unroll buffers into its consituent slices and do away with the buffer construct
	buffers := l.genBuffers(geoInfo)

	positionAttribute, err :=floatsToVertex3s(buffers.vertex) //https://threejs.org/docs/#api/en/core/BufferAttribute

	positionAttribute = applyBufferAttribute(preTransform,positionAttribute)

	geo.position =  positionAttribute
	if len(buffers.colors)> 0 {
		colors, err := floatsToVertex3s(buffers.colors)
		geo.color = make([]Color, len(colors))
		for i, c := range colors{
			geo.color[i] = Color{float32(c.X()), float32(c.Y()), float32(c.Z())}
		}
	}

	if skeleton != nil {
		geo.skinIndex = make([][4]uint16, len(buffers.weightsIndices)/4)
		geo.skinWeight = make([][4]float64, len(buffers.vertexWeights)/4)

		for i:= 0 ; i < len(buffers.weightsIndices); i+=4{
			geo.skinIndex[i/4] = [4]uint16{buffers.weightsIndices[i], buffers.weightsIndices[i+1], buffers.weightsIndices[i+2], buffers.weightsIndices[i+3]}	
		}
		for i:= 0 ; i < len(buffers.vertexWeights); i+=4{
			geo.skinWeight[i/4] = [4]float64{buffers.vertexWeights[i], buffers.vertexWeights[i+1], buffers.vertexWeights[i+2], buffers.vertexWeights[i+3]}	
		}
		geo.FBX_Deformer = skeleton;
	}

	if len(buffers.normal) > 0{
		normalAttribute, err := floatsToVertex3s(buffers.normal)
		normalMatrix := mgl64.Mat4Normal(preTransform) 

		normalAttribute = applyBufferAttributeMat3(normalMatrix,normalAttribute)
		geo.normal = normalAttribute
	}

	geo.uvs = buffers.uvs //NOTE: pulled back from variadic array of uvs where they progress down uv -> uv1 -> uv2 and so on

	if geoInfo.material != nil {
		mat := *geoInfo.material
		if mat.mappingType != "AllSame"{
		// Convert the material indices of each vertex into rendering groups on the geometry.
		prevMaterialIndex := buffers.materialIndex[0]
		startIndex := 0
		for i, matIndex := range buffers.materialIndex{
			if matIndex != prevMaterialIndex{
				geo.AddGroup(startIndex, i-startIndex, prevMaterialIndex) 
				prevMaterialIndex = matIndex
				startIndex = i
			}
		}
		if len( geo.groups) > 0{ //add last group
			lastGroup := geo.groups[ len(geo.groups) - 1 ]
			lastIndex := lastGroup[0] + lastGroup[1]
			if lastIndex != len(buffers.materialIndex) {
				geo.AddGroup( lastIndex, len(buffers.materialIndex) - lastIndex, prevMaterialIndex )
			}
		}
		if len(geo.groups) == 0  {
			geo.AddGroup( 0, len(buffers.materialIndex), buffers.materialIndex[ 0 ] )
		}
	}
		}

	addMorphTargets( l, &geo , geoNode, morphTarget, preTransform )
	return geo, nil
}

type GeoInfo struct {
	vertexPositions []float64 //Todo: parse this immediately as floatgeom.Point3
	vertexIndices []int
	color *floatBuffer
	material *intBuffer
	normal *floatBuffer
	skeleton *Skeleton
	weightTable map[int][]WeightEntry
	uv []floatBuffer
}

func (l *Loader) parseGeoNode(geoNode Node, skeleton *Skeleton) (*GeoInfo, error) {
	geoInfo := &GeoInfo{}
 
	if v, ok := geoNode.props["Vertices"]; ok{
		// Todo: we need to parse out float64s?
		geoInfo.vertexPositions = v.Payload().([]float64)
	}
	
	if v, ok := geoNode.props["PolygonVertexIndex"]; ok{
		// Todo: parse out ints?
		geoInfo.vertexIndices = v.Payload().([]int) 
	}

	if v, ok :=  geoNode.props["LayerElementColor"]; ok{
		v2 := parseVertexColors(v.Payload().([]Node)[0])
		geoInfo.color = &v2
	}
	if v, ok :=  geoNode.props["LayerElementMaterial"]; ok{
		v2 := l.parseMaterialIndices(v.Payload().([]MaterialNode)[0])
		geoInfo.material = &v2
	}
	if v, ok :=  geoNode.props["LayerElementNormal"]; ok{
		v2 := parseNormals(v.Payload().([]Node)[0])
		geoInfo.normal = &v2
	}

	if uvList, ok :=  geoNode.props["LayerElementUV"]; ok{
		uvBuffs := make([]floatBuffer, len(uvList.Payload().([]Node)))
		for i, v := range uvList.Payload().([]Node) { 
			uvBuffs[i] = l.parseUVs(v)
		}
		geoInfo.uv = uvBuffs
	}
	
	if skeleton != nil {
		geoInfo.skeleton = skeleton
		wt := map[int][]WeightEntry{}
		// loop over the bone's vertex indices and weights
		for i, rawb := range skeleton.rawBones{
			for j,rIndex := range rawb.Indices{
				if _, ok := wt[rIndex]; !ok {
					wt[rIndex] = []WeightEntry{}
				}
				wt[rIndex] = append(wt[rIndex],WeightEntry{uint16(i),rawb.Weights[j]})
			}
		}
		geoInfo.weightTable = wt
	}

	return geoInfo, nil
}

func (l *Loader) genBuffers(geoInfo *GeoInfo) gBuffers{
	buffers := gBuffers{}
	polygonIndex := 0
	faceLength := 0
	displayedWeightsWarning := false	

	// tracking faces
	facePositionIndexes := []int{}
	faceNormals := []float64{}
	faceColors := []float64{}
	faceUVs := [][]float64{}
	faceWeights := []float64{}
	faceWeightIndices := []uint16{}
	materialIndex := -1

	for polygonVertexIndex, vertexIndex := range geoInfo.vertexIndices {
		// Face index and vertex index arrays are combined in a single array
		// A cube with quad faces looks like this:
		// PolygonVertexIndex: *24 {
		//  a: 0, 1, 3, -3, 2, 3, 5, -5, 4, 5, 7, -7, 6, 7, 1, -1, 1, 7, 5, -4, 6, 0, 2, -5
		//  }
		// Negative numbers mark the end of a face - first face here is 0, 1, 3, -3
		// to find index of last vertex bit shift the index: ^ - 1
		endOfFace := false
		if vertexIndex < 0{
			vertexIndex = vertexIndex ^ -1
			endOfFace = true
		}
		weightIndices := []uint16{}
		weights := []float64{}
		facePositionIndexes = append(facePositionIndexes,  vertexIndex * 3, vertexIndex * 3 + 1, vertexIndex * 3 + 2 )
		if geoInfo.color!= nil{ 
			data := geoInfo.color.getData(polygonVertexIndex, polygonIndex, vertexIndex)
			faceColors = append(faceColors, data[ 0 ], data[ 1 ], data[ 2 ] )  
		}
		if geoInfo.skeleton!= nil{
			if ws, ok := geoInfo.weightTable[vertexIndex]; ok{
				for _, w := range ws{
					weights = append(weights, w.Weight)
					weightIndices = append(weightIndices, w.ID)
				}
			}
			if len(weights) > 4{
				if !displayedWeightsWarning{
					fmt.Println("Vertex has more than 4 skinning weights assigned to vertex. Deleting additional weights.")
					displayedWeightsWarning = true
				}
				
				//--------------------------------------------------------------------------------------------------------------
				//--------------------------------------------------------------------------------------------------------------
				// TODO: we should talk about this because I am confused about the for each on the Weight as we instantiate it as a [4]int
				// Given that we know js can foreach on an object and interact with its props that makes sense but not incontext of the weight being [4]int
				wIndex := []uint16{0,0,0,0}
				Weight := []float64{0,0,0,0}
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
			for len(weights) <4  {
				weights = append(weights, 0)
				weightIndices = append(weightIndices, 0)
			}

			for i := 1 ; i <= 4; i++ {
				faceWeights = append(faceWeights, weights[i])
				faceWeightIndices = append(faceWeightIndices, weightIndices[i])
			}
		}

		if geoInfo.normal != nil{
			data := geoInfo.normal.getData( polygonVertexIndex, polygonIndex, vertexIndex )
			faceNormals = append(faceNormals,  data[ 0 ], data[ 1 ], data[ 2 ] )
		}
		if  geoInfo.material != nil && geoInfo.material.mappingType != "AllSame"{
			materialIndex = geoInfo.material.getData( polygonVertexIndex, polygonIndex, vertexIndex, )[ 0 ]
		}
		if geoInfo.uv !=nil{
			for i, uv := range geoInfo.uv{
				data := uv.getData( polygonVertexIndex, polygonIndex, vertexIndex,  )
				uvs := faceUVs[i]
				faceUVs[i] = append(faceUVs[i], data[0], data[1])
			}
		}
		faceLength++
		if endOfFace{
			l.genFace( buffers, geoInfo, facePositionIndexes, materialIndex, faceNormals, faceColors, faceUVs, faceWeights, faceWeightIndices, faceLength )
			polygonIndex++ 
			faceLength = 0;
			// reset arrays for the next face
			facePositionIndexes = []int{}
			faceNormals = []float64{}
			faceColors = []float64{}
			faceUVs = [][]float64{}
			faceWeights = []float64{}
			faceWeightIndices = []uint16{}
		}
	}
	return buffers
}

type gBuffers struct{
	vertex []float64
	normal []float64
	colors []float64 
	uvs [][]float64
	materialIndex []int
	vertexWeights []float64
	weightsIndices []uint16
}

// genFace generates data for a single face in a geometry. If the face is a quad then split it into 2 tris
func (l *Loader) genFace(buffers gBuffers, geoInfo *GeoInfo, facePositionIndexes []int, materialIndex int, 
		faceNormals []float64, faceColors []float64, faceUVs [][]float64, faceWeights []float64, faceWeightIndices []uint16, faceLength int) {
	for i := 2 ; i < faceLength; i++{
		buffers.vertex = append(buffers.vertex, genFaceVertex(geoInfo.vertexPositions, facePositionIndexes, i)...)
		if geoInfo.skeleton != nil {
			buffers.vertexWeights = append(buffers.vertexWeights, genFloatFaceArray(4,faceWeights,i)...)
			buffers.weightsIndices = append(buffers.weightsIndices, genUint16FaceArray(4,faceWeightIndices, i)...)
		}
		if geoInfo.color != nil{
			buffers.colors = append(buffers.colors, genFloatFaceArray(3,faceColors,i)...)
		}
		if geoInfo.material != nil && geoInfo.material.mappingType != "AllSame" {
			buffers.materialIndex = append(buffers.materialIndex, materialIndex)//TODO: This seems wrong as it checks that materials are NOT all the same...
			buffers.materialIndex = append(buffers.materialIndex, materialIndex)
			buffers.materialIndex = append(buffers.materialIndex, materialIndex)
		}
		if geoInfo.normal != nil {
			buffers.normal  = append(buffers.normal, genFloatFaceArray(3,faceNormals, i)...)
		}

		if geoInfo.uv != nil {
			buffers.uvs = make([][]float64, len(geoInfo.uv))
			for j, u := range geoInfo.uv{
				buffers.uvs[j] = append(buffers.uvs[j], genFloatFaceArray(2,faceUVs[ j ], i)...)
			}
		}
	}
}

// genFaceVertex simplifies a portion of genface to be less wordy in creation of the vertices for the face
func genFaceVertex(vPos []float64, posIdxs []int, idx int) []float64 {
	return []float64{
		vPos[ posIdxs[0]],vPos[posIdxs[1]],vPos[ posIdxs[2]],
		vPos[ posIdxs[( idx - 1 ) * 3 + 0]],vPos[posIdxs[(idx-1)*3+1]],vPos[ posIdxs[(idx-1)*3+2]],
		vPos[ posIdxs[ idx * 3 + 0]],vPos[posIdxs[ idx * 3 +1]],vPos[ posIdxs[idx * 3 +2]],
	}
}

func genUint16FaceArray(size int, sourceArr []uint16, idx int) []uint16{
	out := make([]uint16, size * 3)
	for j := 0 ; j < size; j++{
		out[ j] = sourceArr[j ]
	}
	for i := 0 ; i < 2; i++{
		for j := 0 ; j < size; j++{
			out[(i+1)*size + j] = sourceArr[j + ((idx-1+i) *size)]
		}
	}
	return out
}

func genIntFaceArray(size int, sourceArr []int, idx int) []int{
	out := make([]int, size * 3)
	for j := 0 ; j < size; j++{
		out[ j] = sourceArr[j ]
	}
	for i := 0 ; i < 2; i++{
		for j := 0 ; j < size; j++{
			out[(i+1)*size + j] = sourceArr[j + ((idx-1+i) *size)]
		}
	}
	return out
}

func genFloatFaceArray(size int, sourceArr []float64, idx int) []float64{
	out := make([]float64, size * 3)
	for j := 0 ; j < size; j++{
		out[ j] = sourceArr[j ]
	}
	for i := 0 ; i < 2; i++{
		for j := 0 ; j < size; j++{
			out[(i+1)*size + j] = sourceArr[j + ((idx-1+i) *size)]
		}
	}
	return out
}

type FaceBuffer struct{
    vertex floatgeom.Point3
    color Color
    weightsIndices []int
    vertexWeights []float64
    normal floatgeom.Point3
    materialIndex int
}

func addMorphTargets( l *Loader, parentGeo *Geometry, parentGeoNode Node, morphTarget *MorphTarget, preTransform mgl64.Mat4) {
	if morphTarget == nil{
		return
	}
	parentGeo.morphAttributes.position = []floatgeom.Point3{}
	parentGeo.morphAttributes.normal = []Color{}
	for  idx, t := range	morphTarget.rawTargets{
		morphGeoNode := l.Tree.Objects.Geometry[t.geoID]
		if morphGeoNode != nil{
			g.genMorphGeometry( parentGeo, parentGeoNode, morphGeoNode, preTransform )
		}
	}
}

type MorphGeoInfo struct {
	vertexIndices []int
	vertexPositions []float64
}

// a morph geometry node is similar to a standard  node, and the node is also contained
// in FBXTree.Objects.Geometry, however it can only have attributes for position, normal
// and a special attribute Index defining which vertices of the original geometry are affected
// Normal and position attributes only have data for the vertices that are affected by the morph
func (g *Geometry) genMorphGeometry(parentGeo *Geometry, parentGeoNode, morphGeoNode Node, preTransform mgl64.Mat4) {
	morphGeo = THREE.BufferGeometry() //TODO: figure out type
	if morphGeoNode.attrName{
		morphGeo.name = morphGeoNode.attrName
	}
	vertexIndices := parentGeoNode.PolygonVertexIndex
	
	// make a copy of the parent's vertex positions
	vertexPositions := []int{}
	copy(vertexPositions,  parentGeoNode.Vertices)
	morphPositions := morphGeoNode.Vertices
	indices := morphGeoNode.Indexes
	for i := 0 ; i < len(indices); i++{
		morphIndex = indices[i] * 3
		// FBX format uses blend shapes rather than morph targets. This can be converted
		// by additively combining the blend shape positions with the original geometry's positions
		vertexPositions[ morphIndex ] += morphPositions[ i * 3 ]
		vertexPositions[ morphIndex + 1 ] += morphPositions[ i * 3 + 1 ]
		vertexPositions[ morphIndex + 2 ] += morphPositions[ i * 3 + 2 ]
	}

	morphGeoInfo := MorphGeoInfo{
		vertexIndices: vertexIndices,
		vertexPositions: vertexPositions,
	}
	morphBuffers = g.genBuffers()

	positionAttribute  := floatsToVertex3s(morphBuffers.vertex)
	positionAttribute.name = morphGeoNode.attrName
	preTransform.applyToBufferAttribute( positionAttribute )

	parentGeo.morphAttirbutes.position = append(parentGeo.morphAttirbutes.position, positionAttribute)
}

// Parse normal from FBXTree.Objects.Geometry.LayerElementNormal if it exists
func parseNormals(NormalNode Node) floatBuffer{
	indexBuffer := []int
	if  referenceType == "IndexToDirect"  {
		if _, ok := NormalNode.props["NormalIndex"] ; ok{
			indexBuffer = NormalNode.NormalIndex.a;
		} else if _, ok2 := NormalNode.props["NormalsIndex"] ; ok2{
			indexBuffer = NormalNode.NormalsIndex.a;
		}
	}
	return floatBuffer{
		dataSize:3,
		buffer: NormalNode.props["Normals"],
		indices: indexBuffer,
		mappingType: NormalNode.MappingInformationType,
		referenceType: NormalNode.ReferenceInformationType,
	}
}
// Parse UVs from FBXTree.Objects.Geometry.LayerElementUV if it exists
func (l *Loader) parseUVs(UVNode Node) floatBuffer{
	mappingType := UVNode.MappingInformationType
	referenceType := UVNode.ReferenceInformationType
	buffer := UVNode.UV
	indexBuffer := []int{}
	if referenceType =="IndexToDirect"  {
		indexBuffer = UVNode.UVIndex
	}
	return floatBuffer{
		dataSize: 2,
		buffer: buffer,
		indices: indexBuffer,
		mappingType: mappingType,
		referenceType: referenceType,
	}
}
// Parse Vertex Colors from FBXTree.Objects.Geometry.LayerElementColor if it exists
func parseVertexColors(ColorNode Node) floatBuffer{
	 indexBuffer := []int
	if referenceType == "IndexToDirect" {
		indexBuffer = ColorNode.ColorIndex
	}
	return floatBuffer{
		dataSize: 4,
		buffer: ColorNode.Colors,
		indices: indexBuffer,
		mappingType: ColorNode.MappingInformationType,
		referenceType: ColorNode.ReferenceInformationType,
	}
}
// Parse mapping and material data in FBXTree.Objects.Geometry.LayerElementMaterial if it exists
func (l *Loader) parseMaterialIndices(MaterialNode) intBuffer{
	mappingType := MaterialNode.MappingInformationType
	referenceType := MaterialNode.ReferenceInformationType

	if mappingType == "NoMappingInformation"{
		return intBuffer{
			dataSize: 1,
			buffer: []int{0},
			indices: []int{0},
			mappingType: "AllSame",
			referenceType: referenceType,
		}
	}

	materialIndexBuffer = MaterialNode.Materials
	for i:=0; i < len(materialIndexBuffer) ; i++{
		materialIndices = append(materialIndices, i)
	}
	return intBuffer{
		dataSize: 1,
		buffer: materialIndexBuffer,
		indices: materialIndices,
		mappingType: mappingType,
		referenceType: referenceType,
	}
}

func (l *Loader) parseNurbsGeometry(geoNode Node) Geometry{
	orderStr, ok :=  geoNode.props["order"]
	if !ok{
		return NewGeometry() // FIgure out how to do this fail state as it used to return new THREE.BufferGeometry();
	}
	order, err := strconv.Atoi(orderStr)
	if err!=nil{
		return NewGeometry() // FIgure out how to do this fail state as it used to return new THREE.BufferGeometry();
		
	}

	
	 degree := order - 1
	 knots := geoNode.props["KnotVector"]
	
	 pointsValues := geoNode.props["Points"]
	 controlPoints := floatsToVertex4s(pointsValues)

	 var startKnot, endKnot int
	 if geoNode.props["Form"] == "Closed"{
		 controlPoints = append(controlPoints, controlPoints[0])
	 }else if geoNode.props["Form"] == "Periodic"{
		startKnot = degree
		endKnot = knots.length - 1 - startKnot
		for i:=1; i <= degree; i++{
			controlPoints = append(controlPoints, controlPoints[i])
		}
	 }

	curve := THREE.NURBSCurve( degree, knots, controlPoints, startKnot, endKnot );
	 vertices := curve.getPoints( controlPoints.length * 7 )
	 positions := floatsToVertex3s(vertices)
	g := NewGeometry()
	g.position =  THREE.BufferAttribute( positions, 3 )
	return g
}
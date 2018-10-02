package threefbx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/oakmound/oak/alg"
	"github.com/oakmound/oak/alg/floatgeom"
)

type Loader struct {
	tree        Tree
	connections ParsedConnections
	sceneGraph  Group
}

func NewLoader() *Loader {
	return &Loader{}
}

type Scene struct{} // ???
type Track struct{} // ???
type KeyframeTrack struct {
	name          string
	times         []float64
	values        []float64 // as long as the times
	interpolation string
} // ???
type Camera struct{}
type Light struct{}
type Model struct{}
type Material struct{}
type Curve struct{}

func (l *Loader) Load(r io.Reader, textureDir string) (*Scene, error) {
	var err error
	if isBinary(r) {
		l.tree, err = l.parseBinary(r)
	} else {
		l.tree, err = l.parseASCII(r)
	}
	if err != nil {
		return nil, err
	}
	if l.tree.Objects.LayeredTexture != nil {
		fmt.Println("layered textures are not supported. Discarding all but first layer.")
	}
	return l.parseTree(textureDir)
}

func (l *Loader) parseTree(textureDir string) (*Scene, error) {
	l.connections = l.parseConnections()
	images, err := l.parseImages()
	if err != nil {
		return nil, err
	}
	textures, err := l.parseTextures(images)
	if err != nil {
		return nil, err
	}
	materials, err := l.parseMaterials(textures)
	if err != nil {
		return nil, err
	}
	skeletons, morphTargets, err := l.parseDeformers()
	if err != nil {
		return nil, err
	}
	geometry, err := l.parseGeometry(skeletons, morphTargets)
	return l.parseScene(skeletons, morphTargets, geometry, materials)
}

type ParsedConnections map[int64]ConnectionSet

type ConnectionSet struct {
	parents  []Connection
	children []Connection
}

func NewParsedConnections() ParsedConnections {
	return ParsedConnections(make(map[int64]ConnectionSet))
}

type Connection struct {
	ID           int64
	To, From     int64
	Relationship string
}

func (l *Loader) parseConnections() ParsedConnections {
	cns := NewParsedConnections()
	for _, cn := range l.tree.connections {
		cns[cn.From].parents = append(cns[cn.From].parents, cn)
		cns[cn.To].children = append(cns[cn.To].children, cn)
	}
	return cns
}

type VideoNode struct {
	ContentType string
	Filename    string
	Content     io.Reader
}

func (l *Loader) parseImages() (map[int64]io.Reader, error) {
	fnms := make(map[int64]string)
	inBlobs := make(map[string]io.Reader)
	outBlobs := make(map[int64]io.Reader)
	var err error
	for id, v := range l.tree.Videos {
		fnms[id] = v.Filename
		if v.Content != nil {
			inBlobs[v.Filename], err = l.parseImage(l.tree.Videos[nodeID])
			if err != nil {
				return nil, err
			}
		}
	}
	for id, fn := range fnms {
		if data, ok := inBlobs[fn]; ok {
			outBlobs[id] = data
		}
	}
	return outBlobs, nil
}

func (l *Loader) parseImage(vn VideoNode) (io.Reader, error) {
	// Todo: handle file types?
	return vn.Content, nil
}

func (l *Loader) parseTextures() (map[int64]Texture, error) {
	txm := make(map[int64]Texture)
	for id, txn := range l.tree.Objects.Texture {
		t, err := l.parseTexture(txn, images)
		if err != nil {
			return nil, err
		}
		txm[id] = t
	}
	return txm, nil
}

type Wrapping int

const (
	RepeatWrapping      Wrapping = iota
	ClampToEdgeWrapping Wrapping = iota
)

type Texture struct {
	ID      int64
	name    string
	wrapS   Wrapping
	wrapT   Wrapping
	repeat  floatgeom.Point2
	content io.Reader
}

func (l *Loader) parseTexture(tx Node, images map[int64]io.Reader) (Texture, error) {
	r, err := l.loadTexture(tx, images)
	return Texture{
		ID:      tx.ID,
		name:    tx.attrName,
		wrapS:   tx.WrapModeU,
		wrapT:   tx.WrapModeV,
		repeat:  tx.Scaling,
		content: r,
	}, err
}

func (l *Loader) parseMaterials(txs map[int64]Texture) map[int64]Material {
	mm := make(map[int64]Material)
	for id, mn := range l.tree.Objects.Material {
		mat := l.parseMaterial(mn, txs)
		if mat != nil {
			mm[id] = mat
		}
	}
	return mm
}

func (l *Loader) parseMaterial(mn MaterialNode, txs map[int64]Texture) Material {
	params := l.parseParameters(mn, txs, mn.ID)
	mat := mn.ShadingModel.Material()
	mat.setValues(params)
	mat.name = mn.attrName
	return mat
}

func (l *Loader) loadTexture(tx Node, images map[int64]io.Reader) (io.Reader, error) {
	// Todo: filetype parsing
	cns := l.connections[tx.ID].children
	if len(cns) < 1 {
		return nil, errors.New("Expected texture node child")
	}
	return images[cns[0].ID], nil
}

type Color struct {
	R, G, B float32
}

type MaterialNode struct {
	BumpFactor         *float64
	Diffuse            *[3]float32
	DiffuseColor       *[3]float32
	DisplacementFactor *float64
	Emissive           *[3]float32
	EmissiveColor      *[3]float32
	EmissiveFactor     *float64
	Opacity            *float64
	ReflectionFactor   *float64
	Shininess          *float64
	Specular           *[3]float32
	SpecularColor      *[3]float32
}

type MaterialParameters struct {
	BumpFactor         float64
	Diffuse            Color
	DisplacementFactor float64
	Emissive           Color
	EmissiveFactor     float64
	Opacity            float64
	ReflectionFactor   float64
	Shininess          float64
	Specular           Color
	Transparent        bool

	bumpMap         Texture
	normalMap       Texture
	specularMap     Texture
	emissiveMap     Texture
	diffuseMap      Texture
	alphaMap        Texture
	displacementMap Texture
	envMap          Texture
}

func (l *Loader) parseParameters(mn MaterialNode, txs map[int64]Texture, id int64) MaterialParameters {
	parameters := MaterialParameters{}

	if materialNode.BumpFactor != nil {
		parameters.BumpFactor = *materialNode.BumpFactor
	}
	if materialNode.Diffuse != nil {
		parameters.Diffuse.R = (*MaterialNode.Diffuse)[0]
		parameters.Diffuse.G = (*MaterialNode.Diffuse)[1]
		parameters.Diffuse.B = (*MaterialNode.Diffuse)[2]
	} else if materialNode.DiffuseColor != nil {
		// The blender exporter exports diffuse here instead of in materialNode.Diffuse
		parameters.Diffuse.R = (*MaterialNode.DiffuseColor)[0]
		parameters.Diffuse.G = (*MaterialNode.DiffuseColor)[1]
		parameters.Diffuse.B = (*MaterialNode.DiffuseColor)[2]
	}
	if materialNode.DisplacementFactor != nil {
		parameters.displacementScale = *materialNode.DisplacementFactor
	}
	if materialNode.Emissive != nil {
		parameters.Emissive.R = (*MaterialNode.Emissive)[0]
		parameters.Emissive.G = (*MaterialNode.Emissive)[1]
		parameters.Emissive.B = (*MaterialNode.Emissive)[2]
	} else if materialNode.EmissiveColor != nil {
		// The blender exporter exports emissive color here instead of in materialNode.Emissive
		parameters.Emissive.R = (*MaterialNode.EmissiveColor)[0]
		parameters.Emissive.G = (*MaterialNode.EmissiveColor)[1]
		parameters.Emissive.B = (*MaterialNode.EmissiveColor)[2]
	}
	if materialNode.EmissiveFactor != nil {
		parameters.EmissiveFactor = *materialNode.EmissiveFactor
	}
	if materialNode.Opacity != nil {
		parameters.Opacity = *materialNode.Opacity
	}
	if parameters.opacity < 1.0 != nil {
		parameters.Transparent = true
	}
	if materialNode.ReflectionFactor != nil {
		parameters.ReflectionFactor = *materialNode.ReflectionFactor
	}
	if materialNode.Shininess != nil {
		parameters.Shininess = *materialNode.Shininess
	}
	if materialNode.Specular != nil {
		parameters.Specular.R = (*MaterialNode.Specular)[0]
		parameters.Specular.G = (*MaterialNode.Specular)[1]
		parameters.Specular.B = (*MaterialNode.Specular)[2]
	} else if materialNode.SpecularColor != nil {
		// The blender exporter exports specular color here instead of in materialNode.Specular
		parameters.Specular.R = (*MaterialNode.SpecularColor)[0]
		parameters.Specular.G = (*MaterialNode.SpecularColor)[1]
		parameters.Specular.B = (*MaterialNode.SpecularColor)[2]
	}

	for _, child := range l.connections[id].children {
		// TODO: Remember to throw away layered things and use the first layer's
		// ID for layered textures
		txt := txs[child.id]
		switch child.Relationship {
		case "Bump":
			parameters.bumpMap = txt
		case "DiffuseColor":
			parameters.diffuseMap = txt
		case "DisplacementColor":
			parameters.displacementMap = txt
		case "EmissiveColor":
			parameters.emissiveMap = txt
		case "NormalMap":
			parameters.normalMap = txt
		case "ReflectionColor":
			parameters.envMap = txt
			parameters.envMap.mapping = EquirectangularReflectionMapping
		case "SpecularColor":
			parameters.specularMap = txt
		case "TransparentColor":
			parameters.alphaMap = txt
			parameters.transparent = true
		//case "AmbientColor":
		//case "ShininessExponent": // AKA glossiness map
		//case "SpecularFactor": // AKA specularLevel
		//case "VectorDisplacementColor": // NOTE: Seems to be a copy of DisplacementColor
		default:
			fmt.Printf("%s map is not supported in three.js, skipping texture.\n", child.Relationship)
		}
	}
	return parameters
}

type Skeleton struct {
	ID int64
	// Todo: instead of rawBones and Bones,
	// if rawBones isn't used after it is 'refined'
	// into bones, have a 'processed' bool?
	rawBones []Bone
	bones    []Bone
}

func (l *Loader) parseDeformers() (map[int64]Skeleton, map[int64]MorphTarget) {
	skeletons := make(map[int64]Skeleton)
	morphTargets := make(map[int64]MorphTarget)
	for id, dn := range l.tree.Objects.Deformer {
		relationships := l.connections[id]
		if dn.attrType == "Skin" {
			skel := l.parseSkeleton(relationships, dn)
			skel.ID = id
			if len(relationships.parents) > 1 {
				fmt.Println("skeleton attached to more than one geometry is not supported.")
			}
			skel.geometryID = relationships.parents[0].ID
			skeletons[id] = skel
		} else if dn.attrType == "BlendShape" {
			mt := MorphTarget{}
			mt.ID = id
			mt.rawTargets = l.parseMorphTargets(relationships, l.tree.Objects.Deformer)
			if len(relationships.parents) > 1 {
				fmt.Println("morph target attached to more than one geometry is not supported.")
			}
			morphTargets[id] = mt
		}
	}
	return skeletons, morphTargets
}

type Bone struct {
	ID            int64
	Indices       []int
	Weights       []float64
	Transform     mgl64.Mat4
	TransformLink mgl64.Mat4
	LinkMode      interface{}
}

// Parse single nodes in tree.Objects.Deformer
// The top level skeleton node has type 'Skin' and sub nodes have type 'Cluster'
// Each skin node represents a skeleton and each cluster node represents a bone
func (l *Loader) parseSkeleton(relationships ConnectionSet, deformerNodes map[int64]Node) Skeleton {
	rawBones := make([]Bone)
	for _, child := range relationships.children {
		boneNode := deformerNodes[child.ID]
		if boneNode.attrType != "Cluster" {
			return
		}
		rawBone := Bone{
			ID:      child.ID,
			Indices: []int{},
			Weights: []float64{},
			// Todo: matrices
			Transform:     mgl64.Mat4FromSlice(boneNode.Transform.a),
			TransformLink: mgl64.Mat4FromSlice(boneNode.TransformLink.a),
			LinkMode:      boneNode.props["Mode"],
		}
		// Todo types, what has 'a' as a field?
		if idxs, ok := boneNode.props["Indexes"]; ok {
			rawBone.Indices = idxs.Payload().([]int)
			rawBone.Weights = boneNode.props["Weights"].Payload().([]float64)
		}
		rawBones = append(rawBones, rawBone)
	}
	return Skeleton{
		rawBones: rawBones,
		bones:    []Bone{},
	}
}

type MorphTarget struct {
	ID            int64
	Name          string
	InitialWeight float64
	FullWeights   []float64
}

// The top level morph deformer node has type "BlendShape" and sub nodes have type "BlendShapeChannel"
func (l *Loader) parseMorphTargets(relationships ConnectionSet, deformerNodes map[int64]Node) []MorphTarget {
	rawMorphTargets := make([]MorphTarget, 0)
	for i := 0; i < relationships.children.length; i++ {
		if i == 8 {
			fmt.Println("FBXLoader: maximum of 8 morph targets supported. Ignoring additional targets.")
			break
		}
		child := relationships.children[i]
		morphTargetNode := deformerNodes[child.ID]
		if morphTargetNode.attrType != "BlendShapeChannel" {
			return
		}
		target := MorphTarget{
			name:          morphTargetNode.attrName,
			initialWeight: morphTargetNode.props["DeformPercent"],
			id:            morphTargetNode.ID,
			fullWeights:   morphTargetNode.FullWeights.Payload().([]float64),
		}
		for _, child := range connections.get(child.ID).children {
			if child.relationship == "" {
				rawMorphTarget.geoID = child.ID
			}
		}
		rawMorphTargets = append(rawMorphTargets, target)
	}
	return rawMorphTargets
}

// create the main THREE.Group() to be returned by the loader
func (l *Loader) parseScene(
	skeletons map[int64]Skeleton,
	morphTargets map[int64]MorphTarget,
	geometryMap Geometry,
	materialMap map[int64]Material) {

	sceneGraph := THREE.Group()
	modelMap := l.parseModels(deformers.skeletons, geometryMap, materialMap)
	modelNodes := tree.Objects.Model
	for _, model := range modelMap {
		modelNode := modelNodes[model.ID]
		l.setLookAtProperties(model, modelNode)
		parentConnections := l.connections.get(model.ID).parents
		for _, connection := range parentConnections {
			var parent = modelMap.get(connection.ID)
			if parent != nil {
				parent.add(model)
			}
		}
		if model.parent == nil {
			sceneGraph.add(model)
		}
	}
	l.bindSkeleton(deformers.skeletons, geometryMap, modelMap)
	l.createAmbientLight()
	l.setupMorphMaterials()
	animations := l.parseAnimations()
	// if all the models where already combined in a single group, just return that
	if sceneGraph.children.length == 1 && sceneGraph.children[0].isGroup {
		sceneGraph.children[0].animations = animations
		sceneGraph = sceneGraph.children[0]
	}
	sceneGraph.animations = animations
}

// parse nodes in FBXTree.Objects.Model
func (l *Loader) parseModels(skeletons map[int64]Skeleton, geometryMap map[int64]Geometry, materialMap map[int64]Material) map[int64]Model {
	modelMap := map[int64]Model{}
	modelNodes := tree.Objects.Model
	for id, node := range modelNodes {
		relationships := connections.get(id)
		model := l.buildSkeleton(relationships, skeletons, id, node.attrName)
		if model == nil {
			switch node.attrType {
			case "Camera":
				model = l.createCamera(relationships)
				break
			case "Light":
				model = l.createLight(relationships)
				break
			case "Mesh":
				model = l.createMesh(relationships, geometryMap, materialMap)
				break
			case "NurbsCurve":
				model = l.createCurve(relationships, geometryMap)
				break
			case "LimbNode": // usually associated with a Bone, however if a Bone was not created we"ll make a Group instead
			case "Null":
			default:
				model = THREE.Group()
				break
			}
			model.name = sanitizeNodeName(node.attrName)
			model.ID = id
		}
		l.setModelTransforms(model, node)
		modelMap.set(id, model)
	}
	return modelMap
}

func (l *Loader) buildSkeleton(relationships ConnectionSet, skeletons map[int64]Skeleton, id int64, name string) Model {
	var bone Model
	for _, parent := range relationships.parents {
		for id, skeleton := range skeletons {
			for i, rawBone := range skeleton.rawBones {
				if rawBone.ID == parent.ID {
					subBone := bone
					bone = THREE.Bone()
					bone.matrixWorld.copy(rawBone.transformLink)
					// set name and id here - otherwise in cases where "subBone" is created it will not have a name / id
					bone.name = sanitizeNodeName(name)
					bone.ID = id
					skeleton.bones[i] = bone
					// In cases where a bone is shared between multiple meshes
					// duplicate the bone here and and it as a child of the first bone
					if subBone != nil {
						bone.add(subBone)
					}
				}
			}
		}
	}
	return bone
}

// create a THREE.PerspectiveCamera or THREE.OrthographicCamera

func createCamera(relationships ConnectionSet) Camera {
	var model Camera
	var cameraAttribute map[string]interface{}
	for i := len(relationships.children) - 1; i > 0; i-- {
		child := relationships.children[i]
		attr := tree.Objects.NodeAttribute[child.ID]
		if attr != nil {
			cameraAttribute = attr.(map[string]interface{})
			break
		}
	}
	if cameraAttribute == "" {
		model = THREE.Object3D()
	} else {
		typ := 0
		v, ok := cameraAttribute["CameraProjectionType"]
		if ok && v.(int64) == 1 {
			typ := 1
		}
		nearClippingPlane := 1
		if v, ok := cameraAttribute["NearPlane"]; ok {
			nearClippingPlane = v / 1000
		}
		farClippingPlane := 1000
		if v, ok := cameraAttribute["FarPlane"]; ok {
			farClippingPlane = v / 1000
		}
		width := 240
		height := 240
		if v, ok := cameraAttribute["AspectWidth"]; ok {
			width = v
		}
		if v, ok := cameraAttribute["AspectHeight"]; ok {
			height = v
		}
		aspect := width / height
		fov := 45
		if v, ok := cameraAttribute["FieldOfView"]; ok {
			fov = v
		}
		focalLength := 0
		if v, ok := cameraAttribute["FocalLength"]; ok {
			focalLength = v
		}
		switch typ {
		case 0: // Perspective
			model = THREE.PerspectiveCamera(fov, aspect, nearClippingPlane, farClippingPlane)
			if focalLength != 0 {
				model.setFocalLength(focalLength)
			}
		case 1: // Orthographic
			model = THREE.OrthographicCamera(-width/2, width/2, height/2, -height/2, nearClippingPlane, farClippingPlane)
		default:
			fmt.Println("Unknown camera type", typ)
			model = THREE.Object3D()
		}
	}
	return model
}

type LightType int

const (
	LightPoint       LightType = iota
	LightDirectional LightType = iota
	LightSpot        LightType = iota
)

// Todo: less pointers, more reasonable zero values
type LightAttribute struct {
	LightType         *LightType
	Color             Color
	Intensity         *float64
	FarAttenuationEnd *float64
	InnerAngle        *float64
	OuterAngle        *float64

	CastLightOnObject    bool
	EnableFarAttenuation bool
	CastShadows          bool
}

// Create a THREE.DirectionalLight, THREE.PointLight or THREE.SpotLight
func (l *Loader) createLight(relationships ConnectionSet) Light {
	var model Light
	var la LightAttribute

	for i := len(relationships.children) - 1; i >= 0; i-- {
		var attr = tree.Objects.NodeAttribute[relationships.children[i].ID]
		if attr != "" {
			la = attr.(LightAttribute)
			break
		}
	}
	if la == "" {
		model = THREE.Object3D()
	} else {
		var typ int
		// LightType can be undefined for Point lights
		if la.LightType != nil {
			typ = *la.LightType
		}
		var color = color.RBGA{255, 255, 255, 255}
		if la.Color != nil {
			color = *la.Color
		}
		intensity := 1.0
		if *la.Intensity != nil {
			intensity = *la.Intensity / 100
		}
		// light disabled
		if !la.CastLightOnObject {
			intensity = 0
		}
		distance := 0.0
		if la.FarAttenuationEnd != nil {
			if la.EnableFarAttenuation {
				distance = *la.FarAttenuationEnd
			}
		}
		// TODO: could this be calculated linearly from FarAttenuationStart to FarAttenuationEnd?
		var decay = 1
		switch typ {
		case 0: // Point
			model = THREE.PointLight(color, intensity, distance, decay)
		case 1: // Directional
			model = THREE.DirectionalLight(color, intensity)
		case 2: // Spot
			angle := math.Pi / 3
			if la.InnerAngle != nil {
				angle = alg.DegToRad * *la.InnerAngle.value
			}
			penumbra := 0.0
			if la.OuterAngle != nil {
				// TODO: this is not correct - FBX calculates outer and inner angle in degrees
				// with OuterAngle > InnerAngle && OuterAngle <= Math.PI
				// while three.js uses a penumbra between (0, 1) to attenuate the inner angle
				penumbra = alg.DegToRad * *la.OuterAngle
				penumbra = math.Max(penumbra, 1)
			}
			model = THREE.SpotLight(color, intensity, distance, angle, penumbra, decay)
		default:
			fmt.Println("THREE.FBXLoader: Unknown light type " + la.LightType.value + ", defaulting to a THREE.PointLight.")
			model = THREE.PointLight(color, intensity)
		}
		if la.CastShadows {
			model.castShadow = true
		}
	}
	return model
}

func (l *Loader) createMesh(relationships ConnectionSet, geometryMap map[int64]Geometry, materialMap map[int64]Material) {
	var model Model
	var geometry *Geometry
	var materials = []*Material{}
	// get geometry and materials(s) from connections
	for _, child := range relationships.children {
		if g, ok := geometryMap[child.ID]; ok {
			geometry = g
		}
		if m, ok := materialMap[child.ID]; ok {
			materials = append(materials, m)
		}
	}
	if len(materials) == 0 {
		materials = []*Material{THREE.MeshPhongMaterial(map[string]Color{"color": 0xcccccc})}
	}
	if len(geometry.color) > 0 {
		for _, m := range materials {
			m.vertexColors = THREE.VertexColors
		}
	}
	if geometry.FBX_Deformer != nil {
		for _, m := range materials {
			m.skinning = true
		}
		model = THREE.SkinnedMesh(geometry, materials)
	} else {
		model = THREE.Mesh(geometry, materials)
	}
	return model
}

// parse the model node for transform details and apply them to the model
func (l *Loader) setModelTransforms(model Model, modelNode Node) {
	var td = TransformData{}
	if v, ok := modelNode.props["RotationOrder"]; ok {
		td.eulerOrder = &EulerOrder(v.Payload().(int))
	}
	if v, ok := modelNode.props["Lcl_Translation"]; ok {
		td.translation = &v.Payload().(floatgeom.Point3)
	}
	if v, ok := modelNode.props["RotationOffset"]; ok {
		td.rotationOffset = &v.Payload().(floatgeom.Point3)
	}
	if v, ok := modelNode.props["Lcl_Rotation"]; ok {
		td.rotation = &v.Payload().(mgl64.Mat4)
	}
	if v, ok := modelNode.props["PreRotation"]; ok {
		td.preRotation = &v.Payload().(mgl64.Mat4)
	}
	if v, ok := modelNode.props["PostRotation"]; ok {
		td.postRotation = &v.Payload().(mgl64.Mat4)
	}
	if v, ok := modelNode.props["Lcl_Scaling"]; ok {
		td.scale = &v.Payload().(floatgeom.Point3)
	}
	model.applyMatrix(generateTransform(td))
}

func (l *Loader) createCurve(relationships ConnectionSet, geometryMap map[int64]Geometry) Curve {
	var geometry Geometry
	for i := len(relationships.children) - 1; i >= 0; i-- {
		child := relationships.children[i]
		if geo, ok := geometryMap[child.ID]; ok {
			geometry = geo
			break
		}
	}
	// FBX does not list materials for Nurbs lines, so we'll just put our own in here.
	material := THREE.LineBasicMaterial(0x3300ff, 1)
	return THREE.Line(geometry, material)
}

func (l *Loader) setLookAtProperties(model Model, modelNode Node) {
	if _, ok := modelNode.props["LookAtProperty"]; ok {
		children := l.connections[model.ID].children
		for _, child := range children {
			if child.Relationship == "LookAtProperty" {
				lookAtTarget := tree.Objects.Model[child.ID]
				if _, ok := lookAtTarget["Lcl_Translation"]; ok {
					pos := lookAtTarget["Lcl_Translation"].Payload().([]float64)
					// DirectionalLight, SpotLight
					if model.target != nil {
						model.target.position.fromArray(pos)
						sceneGraph.add(model.target)
					} else { // Cameras and other Object3Ds
						model.lookAt(floatgeom.Point3{pos[0], pos[1], pos[2]})
					}
				}
			}
		}
	}
}

func bindSkeleton(skeletons map[int64]Skeleton, geometryMap map[int64]Geometry, modelMap map[int64]*Model) {
	bindMatrices := l.parsePoseNodes()
	for id, skeleton := range skeletons {
		for _, parent := range connections.get(skeleton.ID).parents {
			if _, ok := geometryMap[parent.ID]; ok {
				for _, geoConnParent := range connections.get(parent.ID).parents {
					if model, ok := modelMap[geoConnParent.ID]; ok {
						model.bind(THREE.Skeleton(skeleton.bones), bindMatrices[geoConnParent.ID])
					}
				}
			}
		}
	}
}

func parsePoseNodes() {
	var bindMatrices = map[Node]Matrix4{}
	if BindPoseNode, ok := tree.Objects["Pose"]; ok {
		for nodeID, v := range BindPoseNode {
			if v.attrType == "BindPose" {
				poseNodes := v.props["PoseNode"]
				if poseNodes.IsArray() {
					for _, n := range poseNodes {
						bindMatrices[n.Node] = n.Payload().(Matrix4)
					}
				} else {
					bindMatrices[poseNodes.Node] = poseNodes.Payload().(Matrix4)
				}
			}
		}
	}
	return bindMatrices
}

// Parse ambient color in tree.GlobalSettings - if it's not set to black (default), create an ambient light
func createAmbientLight() {
	_, ok := tree["GlobalSetttings"]
	ambientColor, ok2 := tree["AmbientColor"].(color.RGBA)
	if ok && ok2 {
		if ambientColor.R != 0 || ambientColor.G != 0 || ambientColor.B != 0 {
			sceneGraph.add(THREE.AmbientLight(ambientColor, 1))
		}
	}
}

func (l *Loader) setupMorphMaterials() {
	sceneGraph.traverse(func(c SceneNode) {
		if c.isMesh {
			_, ok := c.geometry.morphAttributes["position"]
			_, ok2 := c.geometry.morphAttributes["normal"]
			if ok || ok2 {
				// if a geometry has morph targets, it cannot share the material with other geometries
				sharedMat := false
				sceneGraph.traverse(func(c2 SceneNode) {
					if c2.isMesh {
						if c2.material.uuid == c.material.uuid && c2.uuid != c.uuid {
							sharedMat = true
							return
						}
					}
				})
				if sharedMat {
					c.material = child.material.clone()
				}
				c.material.morphTargets = true
			}
		}
	})
}

// FBXTree holds a representation of the FBX data, returned by the TextParser ( FBX ASCII format)
// and BinaryParser( FBX Binary format)
type Tree struct {
	Objects map[string][]*Node
}

func isBinary(r io.Reader) bool {
	magic := append([]byte("Kaydara FBX Binary  "), 0)
	header := make([]byte, len(magic))
	n, err := r.Read(header)
	if n != len(header) {
		return false
	}
	if err != nil {
		return false
	}
	return bytes.Equal(magic, header)
}

var fbxVersionMatch *regexp.Regexp

func init() {
	var err error
	fbxVersionMatch, err = regexp.Compile("FBXVersion: (\\d+)")
	if err != nil {
		fmt.Println("Unable to compile fbx version regex:", err)
	}
}

func getFBXVersion(text string) (int, error) {
	matches := fbxVersionMatch.FindStringSubmatch(text)
	if len(matches) > 0 {
		i, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, err
		}
		return i, nil
	}
	return 0, errors.New("THREE.FBXLoader: Cannot find the version number for the file given.")
}

func fbxTimeToSeconds(value int64) float64 {
	return float64(value) / float64(46186158000)
}

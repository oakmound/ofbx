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
	tree        *Tree
	connections ParsedConnections
	sceneGraph  Group
}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Load(r io.Reader, textureDir string) (Model, error) {
	var err error
	if isBinary(r) {
		l.tree, err = l.parseBinary(r)
	} else {
		l.tree, err = l.parseASCII(r)
	}
	if err != nil {
		return nil, err
	}
	if _, ok := l.tree.Objects["LayeredTexture"]; ok {
		fmt.Println("layered textures are not supported. Discarding all but first layer.")
	}
	return l.parseTree(textureDir)
}

func (l *Loader) parseTree(textureDir string) (Model, error) {
	var err error
	l.connections, err = l.parseConnections()
	if err != nil {
		return nil, err
	}
	images, err := l.parseImages()
	if err != nil {
		return nil, err
	}
	textures, err := l.parseTextures(images)
	if err != nil {
		return nil, err
	}
	materials := l.parseMaterials(textures)
	skeletons, morphTargets := l.parseDeformers()
	geometry, err := l.parseGeometry(skeletons, morphTargets)
	return l.parseScene(skeletons, morphTargets, geometry, materials), nil
}

type ParsedConnections map[int]ConnectionSet

type ConnectionSet struct {
	parents  []Connection
	children []Connection
}

func NewParsedConnections() ParsedConnections {
	return ParsedConnections(make(map[int]ConnectionSet))
}

type Connection struct {
	ID           int
	To, From     int
	Relationship string
}

func NewConnection(n *Node) (Connection, error) {
	cn := Connection{}
	var ok bool
	cn.ID, ok = n.props["ID"].Payload().(int)
	if !ok {
		return cn, errors.New("Node lacking Connection properties")
	}
	cn.To, ok = n.props["To"].Payload().(int)
	if !ok {
		return cn, errors.New("Node lacking Connection properties")
	}
	cn.From, ok = n.props["From"].Payload().(int)
	if !ok {
		return cn, errors.New("Node lacking Connection properties")
	}
	cn.Relationship, ok = n.props["Relationship"].Payload().(string)
	if !ok {
		return cn, errors.New("Node lacking Connection properties")
	}
	return cn, nil
}

func (l *Loader) parseConnections() (ParsedConnections, error) {
	cns := NewParsedConnections()
	for _, n := range l.tree.Objects["Connections"] {
		cn, err := NewConnection(n)
		if err != nil {
			return nil, err
		}
		cf := cns[cn.From]
		ct := cns[cn.To]
		cf.parents = append(cf.parents, cn)
		ct.children = append(ct.children, cn)
		cns[cn.From] = cf
		cns[cn.To] = ct
	}
	return cns, nil
}

type VideoNode struct {
	ContentType string
	Filename    string
	Content     io.Reader
}

func NewVideoNode(n *Node) (VideoNode, error) {
	vn := VideoNode{}
	var ok bool
	vn.Filename, ok = n.props["Filename"].Payload().(string)
	if !ok {
		return vn, errors.New("Node lacking VideoNode properties")
	}
	vn.ContentType, ok = n.props["ContentType"].Payload().(string)
	if !ok {
		return vn, errors.New("Node lacking VideoNode properties")
	}
	vn.Content, _ = n.props["Content"].Payload().(io.Reader)
	return vn, nil
}

func (l *Loader) parseImages() (map[int]io.Reader, error) {
	fnms := make(map[int]string)
	inBlobs := make(map[string]io.Reader)
	outBlobs := make(map[int]io.Reader)
	var err error
	vids := l.tree.Objects["Videos"]
	for id, v := range vids {
		vn, err := NewVideoNode(v)
		if err != nil {
			return nil, err
		}
		fnms[id] = vn.Filename
		if vn.Content != nil {
			inBlobs[fnms[id]], err = l.parseImage(vn)
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

func (l *Loader) parseTextures(images map[int]io.Reader) (map[int]Texture, error) {
	txm := make(map[int]Texture)
	for id, txn := range l.tree.Objects["Texture"] {
		t, err := l.parseTexture(*txn, images)
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
	ID      int
	name    string
	wrapS   Wrapping
	wrapT   Wrapping
	repeat  floatgeom.Point2
	mapping TextureMapping
	content io.Reader
}

func (l *Loader) parseTexture(tx Node, images map[int]io.Reader) (Texture, error) {
	r, err := l.loadTexture(tx, images)
	return Texture{
		ID:      tx.ID,
		name:    tx.attrName,
		wrapS:   tx.props["WrapModeU"].Payload().(Wrapping),
		wrapT:   tx.props["WrapModeV"].Payload().(Wrapping),
		repeat:  tx.props["Scaling"].Payload().(floatgeom.Point2),
		content: r,
	}, err
}

func (l *Loader) parseMaterials(txs map[int]Texture) map[int]Material {
	mm := make(map[int]Material)
	for id, mnt := range l.tree.Objects["Material"] {
		mn := NewMaterialNode(mnt)
		mm[id] = l.parseMaterial(mn, txs)
	}
	return mm
}

func (l *Loader) parseMaterial(mn MaterialNode, txs map[int]Texture) Material {
	mat := mn.ShadingModel.Material()
	mat.Name = mn.attrName
	mat = l.parseParameters(mn, txs, mn.ID, mat)
	return mat
}

func (l *Loader) loadTexture(tx Node, images map[int]io.Reader) (io.Reader, error) {
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
	*Node
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
	ShadingModel       *ShadingModel
}

// NewMaterialNode creates a material node from a Node!
func NewMaterialNode(n *Node) MaterialNode {
	mn := MaterialNode{Node: n}
	v, ok := mn.props["BumpFactor"].Payload().(float64)
	if ok {
		mn.BumpFactor = &v
	}
	v2, ok := mn.props["Diffuse"].Payload().([3]float32)
	if ok {
		mn.Diffuse = &v2
	}
	v3, ok := mn.props["DiffuseColor"].Payload().([3]float32)
	if ok {
		mn.DiffuseColor = &v3
	}
	v4, ok := mn.props["DisplacementFactor"].Payload().(float64)
	if ok {
		mn.DisplacementFactor = &v4
	}
	v5, ok := mn.props["Emissive"].Payload().([3]float32)
	if ok {
		mn.Emissive = &v5
	}
	v6, ok := mn.props["EmissiveColor"].Payload().([3]float32)
	if ok {
		mn.EmissiveColor = &v6
	}
	v7, ok := mn.props["EmissiveFactor"].Payload().(float64)
	if ok {
		mn.EmissiveFactor = &v7
	}
	v8, ok := mn.props["Opacity"].Payload().(float64)
	if ok {
		mn.Opacity = &v8
	}
	v9, ok := mn.props["ReflectionFactor"].Payload().(float64)
	if ok {
		mn.ReflectionFactor = &v9
	}
	v10, ok := mn.props["Shininess"].Payload().(float64)
	if ok {
		mn.Shininess = &v10
	}
	v11, ok := mn.props["Specular"].Payload().([3]float32)
	if ok {
		mn.Specular = &v11
	}
	v12, ok := mn.props["SpecularColor"].Payload().([3]float32)
	if ok {
		mn.SpecularColor = &v12
	}
	v13, ok := mn.props["ShadingModel"].Payload().(ShadingModel)
	if ok {
		mn.ShadingModel = &v13
	}

	return mn
}

func (l *Loader) parseParameters(materialNode MaterialNode, txs map[int]Texture, id int, mat Material) Material {

	if materialNode.BumpFactor != nil {
		mat.BumpFactor = *materialNode.BumpFactor
	}
	if materialNode.Diffuse != nil {
		mat.Diffuse.R = (*materialNode.Diffuse)[0]
		mat.Diffuse.G = (*materialNode.Diffuse)[1]
		mat.Diffuse.B = (*materialNode.Diffuse)[2]
	} else if materialNode.DiffuseColor != nil {
		// The blender exporter exports diffuse here instead of in materialNode.Diffuse
		mat.Diffuse.R = (*materialNode.DiffuseColor)[0]
		mat.Diffuse.G = (*materialNode.DiffuseColor)[1]
		mat.Diffuse.B = (*materialNode.DiffuseColor)[2]
	}
	if materialNode.DisplacementFactor != nil {
		mat.DisplacementFactor = *materialNode.DisplacementFactor
	}
	if materialNode.Emissive != nil {
		mat.Emissive.R = (*materialNode.Emissive)[0]
		mat.Emissive.G = (*materialNode.Emissive)[1]
		mat.Emissive.B = (*materialNode.Emissive)[2]
	} else if materialNode.EmissiveColor != nil {
		// The blender exporter exports emissive color here instead of in materialNode.Emissive
		mat.Emissive.R = (*materialNode.EmissiveColor)[0]
		mat.Emissive.G = (*materialNode.EmissiveColor)[1]
		mat.Emissive.B = (*materialNode.EmissiveColor)[2]
	}
	if materialNode.EmissiveFactor != nil {
		mat.EmissiveFactor = *materialNode.EmissiveFactor
	}
	if materialNode.Opacity != nil {
		mat.Opacity = *materialNode.Opacity
	}
	if mat.Opacity < 1.0 {
		mat.Transparent = true
	}
	if materialNode.ReflectionFactor != nil {
		mat.ReflectionFactor = *materialNode.ReflectionFactor
	}
	if materialNode.Shininess != nil {
		mat.Shininess = *materialNode.Shininess
	}
	if materialNode.Specular != nil {
		mat.Specular.R = (*materialNode.Specular)[0]
		mat.Specular.G = (*materialNode.Specular)[1]
		mat.Specular.B = (*materialNode.Specular)[2]
	} else if materialNode.SpecularColor != nil {
		// The blender exporter exports specular color here instead of in materialNode.Specular
		mat.Specular.R = (*materialNode.SpecularColor)[0]
		mat.Specular.G = (*materialNode.SpecularColor)[1]
		mat.Specular.B = (*materialNode.SpecularColor)[2]
	}

	for _, child := range l.connections[id].children {
		// TODO: Remember to throw away layered things and use the first layer's
		// ID for layered textures
		txt := txs[int(child.ID)]
		switch child.Relationship {
		case "Bump":
			mat.bumpMap = txt
		case "DiffuseColor":
			mat.diffuseMap = txt
		case "DisplacementColor":
			mat.displacementMap = txt
		case "EmissiveColor":
			mat.emissiveMap = txt
		case "NormalMap":
			mat.normalMap = txt
		case "ReflectionColor":
			mat.envMap = txt
			mat.envMap.mapping = EquirectangularReflectionMapping
		case "SpecularColor":
			mat.specularMap = txt
		case "TransparentColor":
			mat.alphaMap = txt
			mat.Transparent = true
		//case "AmbientColor":
		//case "ShininessExponent": // AKA glossiness map
		//case "SpecularFactor": // AKA specularLevel
		//case "VectorDisplacementColor": // NOTE: Seems to be a copy of DisplacementColor
		default:
			fmt.Printf("%s map is not supported in three.js, skipping texture.\n", child.Relationship)
		}
	}
	return mat
}

type Skeleton struct {
	ID         int
	geometryID int
	// Todo: instead of rawBones and Bones,
	// if rawBones isn't used after it is 'refined'
	// into bones, have a 'processed' bool?
	rawBones []Bone
	bones    []Model
}

func (l *Loader) parseDeformers() (map[int]Skeleton, map[int]MorphTarget) {
	skeletons := make(map[int]Skeleton)
	morphTargets := make(map[int]MorphTarget)
	deformer := l.tree.Objects["Deformer"]
	for id, dn := range deformer {
		relationships := l.connections[id]
		if dn.attrType == "Skin" {
			skel := l.parseSkeleton(relationships, deformer)
			skel.ID = id
			if len(relationships.parents) > 1 {
				fmt.Println("skeleton attached to more than one geometry is not supported.")
			}
			skel.geometryID = relationships.parents[0].ID
			skeletons[id] = skel
		} else if dn.attrType == "BlendShape" {
			mt := MorphTarget{}
			mt.ID = id
			mt.RawTargets = l.parseMorphTargets(relationships, l.tree.Objects["Deformer"])
			if len(relationships.parents) > 1 {
				fmt.Println("morph target attached to more than one geometry is not supported.")
			}
			morphTargets[id] = mt
		}
	}
	return skeletons, morphTargets
}

type Bone struct {
	ID            int
	Indices       []int
	Weights       []float64
	Transform     mgl64.Mat4
	TransformLink mgl64.Mat4
	LinkMode      interface{}
}

// Parse single nodes in tree.Objects.Deformer
// The top level skeleton node has type 'Skin' and sub nodes have type 'Cluster'
// Each skin node represents a skeleton and each cluster node represents a bone
func (l *Loader) parseSkeleton(relationships ConnectionSet, deformerNodes map[int]*Node) Skeleton {
	rawBones := make([]Bone, 0)
	for _, child := range relationships.children {
		boneNode := deformerNodes[child.ID]
		if boneNode.attrType != "Cluster" {
			continue
		}
		rawBone := Bone{
			ID:      child.ID,
			Indices: []int{},
			Weights: []float64{},
			// Todo: matrices
			Transform:     Mat4FromSlice(boneNode.props["Transform"].Payload().([]float64)),
			TransformLink: Mat4FromSlice(boneNode.props["TransformLink"].Payload().([]float64)),
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
		bones:    []Model{},
	}
}

type MorphTarget struct {
	ID         int
	RawTargets []RawTarget
}

type RawTarget struct {
	id            int
	geoID         int
	name          string
	initialWeight float64
	fullWeights   []float64
}

// The top level morph deformer node has type "BlendShape" and sub nodes have type "BlendShapeChannel"
func (l *Loader) parseMorphTargets(relationships ConnectionSet, deformerNodes map[int]*Node) []RawTarget {
	rawMorphTargets := make([]RawTarget, 0)
	for i := 0; i < len(relationships.children); i++ {
		child := relationships.children[i]
		morphTargetNode := deformerNodes[child.ID]
		if morphTargetNode.attrType != "BlendShapeChannel" {
			continue
		}
		target := RawTarget{
			name:          morphTargetNode.attrName,
			initialWeight: morphTargetNode.props["DeformPercent"].Payload().(float64),
			id:            morphTargetNode.ID,
			fullWeights:   morphTargetNode.props["FullWeights"].Payload().([]float64),
		}
		for _, child := range l.connections[child.ID].children {
			if child.Relationship == "" {
				target.geoID = child.ID
			}
		}
		rawMorphTargets = append(rawMorphTargets, target)
		if len(rawMorphTargets) == 8 {
			fmt.Println("FBXLoader: maximum of 8 morph targets supported. Ignoring additional targets.")
			break
		}
	}
	return rawMorphTargets
}

// create the main THREE.Group() to be returned by the loader
func (l *Loader) parseScene(
	skeletons map[int]Skeleton,
	morphTargets map[int]MorphTarget,
	geometryMap map[int]Geometry,
	materialMap map[int]Material) Model {

	var sceneGraph Model = &ModelGroup{}
	modelMap := l.parseModels(skeletons, geometryMap, materialMap)
	modelNodes := l.tree.Objects["Model"]
	for id, model := range modelMap {
		modelNode := modelNodes[id]
		l.setLookAtProperties(model, modelNode)
		parentConnections := l.connections[id].parents
		for _, connection := range parentConnections {
			var parent, ok = modelMap[connection.ID]
			if ok {
				parent.AddChild(model)
			}
		}
		if model.Parent() == nil {
			sceneGraph.AddChild(model)
		}
	}
	l.bindSkeleton(skeletons, geometryMap, modelMap)
	l.createAmbientLight()
	l.setupMorphMaterials()
	// if all the models where already combined in a single group, just return that
	if len(sceneGraph.Children()) == 1 && sceneGraph.Children()[0].IsGroup() {
		sceneGraph = sceneGraph.Children()[0]
	}
	sceneGraph.SetAnimations(l.parseAnimations())
	return sceneGraph
}

// parse nodes in FBXTree.Objects.Model
func (l *Loader) parseModels(skeletons map[int]Skeleton, geometryMap map[int]Geometry, materialMap map[int]Material) map[int]Model {
	modelMap := map[int]Model{}
	modelNodes := l.tree.Objects["Model"]
	for id, node := range modelNodes {
		relationships := l.connections[id]
		var model Model = l.buildSkeleton(relationships, skeletons, id, node.attrName)
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
				model = &ModelGroup{}
				break
			}
			model.SetName(sanitizeNodeName(node.attrName))
			model.SetID(id)
		}
		l.setModelTransforms(model, node)
		modelMap[id] = model
	}
	return modelMap
}

func (l *Loader) buildSkeleton(relationships ConnectionSet, skeletons map[int]Skeleton, id int, name string) *BoneModel {
	var bone *BoneModel
	for _, parent := range relationships.parents {
		for id, skeleton := range skeletons {
			for i, rawBone := range skeleton.rawBones {
				if rawBone.ID == parent.ID {
					subBone := bone
					bone = &BoneModel{}
					bone.matrixWorld = rawBone.TransformLink
					// set name and id here - otherwise in cases where "subBone" is created it will not have a name / id
					bone.SetName(sanitizeNodeName(name))
					bone.SetID(id)
					skeleton.bones[i] = bone
					// In cases where a bone is shared between multiple meshes
					// duplicate the bone here and and it as a child of the first bone
					if subBone != nil {
						bone.AddChild(subBone)
					}
				}
			}
		}
	}
	return bone
}

// create a THREE.PerspectiveCamera or THREE.OrthographicCamera

func (l *Loader) createCamera(relationships ConnectionSet) Model {
	var model Camera
	var cameraAttribute map[string]interface{}
	for i := len(relationships.children) - 1; i > 0; i-- {
		child := relationships.children[i]
		attr := l.tree.Objects["NodeAttribute"][child.ID]
		if attr != nil {
			for k, prop := range attr.props {
				cameraAttribute[k] = prop.Payload()
			}
			break
		}
	}
	if cameraAttribute == nil {
		return nil
	}
	typ := 0
	v, ok := cameraAttribute["CameraProjectionType"]
	if ok && v.(int64) == 1 {
		typ := 1
	}
	nearClippingPlane := 1
	if v, ok := cameraAttribute["NearPlane"]; ok {
		nearClippingPlane = v.(int) / 1000
	}
	farClippingPlane := 1000
	if v, ok := cameraAttribute["FarPlane"]; ok {
		farClippingPlane = v.(int) / 1000
	}
	width := 240
	height := 240
	if v, ok := cameraAttribute["AspectWidth"]; ok {
		width = v.(int)
	}
	if v, ok := cameraAttribute["AspectHeight"]; ok {
		height = v.(int)
	}
	aspect := width / height
	fov := 45
	if v, ok := cameraAttribute["FieldOfView"]; ok {
		fov = v.(int)
	}
	focalLength := 0
	if v, ok := cameraAttribute["FocalLength"]; ok {
		focalLength = v.(int)
	}
	switch typ {
	case 0: // Perspective
		model = NewPerspectiveCamera(fov, aspect, nearClippingPlane, farClippingPlane)
		if focalLength != 0 {
			model.SetFocalLength(focalLength)
		} 
	case 1: // Orthographic
		model = NewOrthographicCamera(-width/2, width/2, height/2, -height/2, nearClippingPlane, farClippingPlane)
	default:
		fmt.Println("Unknown camera type", typ)
		return nil
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
type LightNode struct {
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

func NewLightNode(n *Node) (*LightNode, error) {
	ln := &LightNode{}
	lt, ok := n.props["LightType"].Payload().(int)
	if ok {
		v := LightType(lt)
		ln.LightType = &v
	}
	ln.Color, ok = n.props["Color"].Payload().(Color)
	if !ok {
		return nil, errors.New("Invalid LightNode")
	}
	var f float64
	f, ok = n.props["Intensity"].Payload().(float64)
	if ok {
		ln.Intensity = &f
	}
	f, ok = n.props["FarAttenuationEnd"].Payload().(float64)
	if ok {
		ln.FarAttenuationEnd = &f
	}
	f, ok = n.props["InnerAngle"].Payload().(float64)
	if ok {
		ln.InnerAngle = &f
	}
	f, ok = n.props["OuterAngle"].Payload().(float64)
	if ok {
		ln.OuterAngle = &f
	}

	var b bool 
	ln.CastLightOnObject, ok = n.props["CastLightOnObject"].Payload().(bool)
	if !ok {
		return nil, errors.New("Invalid LightNode")
	}
	ln.EnableFarAttenuation, ok = n.props["EnableFarAttenuation"].Payload().(bool)
	if !ok {
		return nil, errors.New("Invalid LightNode")
	}
	ln.CastShadows, ok = n.props["CastShadows"].Payload().(bool)
	if !ok {
		return nil, errors.New("Invalid LightNode")
	}
	return ln, nil
}

// Create a THREE.DirectionalLight, THREE.PointLight or THREE.SpotLight
func (l *Loader) createLight(relationships ConnectionSet) Light {
	var model Light
	var err error
	var la *LightNode

	for i := len(relationships.children) - 1; i >= 0; i-- {
		node := l.tree.Objects["NodeAttribute"][relationships.children[i].ID]
		if node != nil {
			la, err = NewLightNode(node)
			if err != nil {
				fmt.Println("Error creating light node", err)
				continue
			}
			break
		}
	}
	if la == nil {
		return nil
	} 
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
		model = NewPointLight(color, intensity, distance, decay)
	case 1: // Directional
		model = NewDirectionalLight(color, intensity)
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
		model = NewSpotLight(color, intensity, distance, angle, penumbra, decay)
	default:
		fmt.Println("THREE.FBXLoader: Unknown light type " + la.LightType.value + ", defaulting to a THREE.PointLight.")
		model = NewPointLight(color, intensity)
	}
	if la.CastShadows {
		model.castShadow = true
	}
	return model
}

func (l *Loader) createMesh(relationships ConnectionSet, geometryMap map[int]Geometry, materialMap map[int]Material) Model {
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
func (l *Loader) setModelTransforms(model Model, modelNode *Node) {
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

func (l *Loader) createCurve(relationships ConnectionSet, geometryMap map[int]Geometry) Curve {
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

func (l *Loader) setLookAtProperties(model Model, modelNode *Node) {
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

func (l *Loader) bindSkeleton(skeletons map[int]Skeleton, geometryMap map[int]Geometry, modelMap map[int]Model) {
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
func (l *Loader) createAmbientLight() {
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
	Objects map[string]map[int]*Node
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

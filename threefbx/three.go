package threefbx

import (
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/oakmound/oak/alg"
	"github.com/oakmound/oak/alg/floatgeom"
	"github.com/oakmound/ofbx"
)

type Loader struct {
	tree           *Tree
	connections    ParsedConnections
	rawConnections []Connection
	sceneGraph     *Scene
}

func NewLoader() *Loader {
	return &Loader{}
}

func Load(r io.Reader, textureDir string) (*Scene, error) {
	l := NewLoader()
	return l.Load(r, textureDir)
}

func (l *Loader) Load(r io.Reader, textureDir string) (*Scene, error) {
	var err error
	if ofbx.IsBinary(r) {
		l.tree, err = l.ParseBinary(r)
	} else {
		tree, err := l.ParseASCII(r)
		if err != nil {
			return nil, err
		}
		l.tree = &tree
	}
	if err != nil {
		return nil, err
	}

	// Objects["Objects"] is a map which we want to pull up a level
	//We anticipate that there should only be one thing in objects but if not we have this loop
	for _, objects := range l.tree.Objects["Objects"] { //TODO: should we panic if it loops?
		for oType, propObj := range objects.props {
			//propObj is a property that contains all of the things of oType (like animationStacks)

			objs := propObj.Payload.(map[string]Property)
			l.tree.Objects[oType] = make(map[IDType]*Node)

			// fmt.Printf("type of object, %s is %T \n", oType, obj)
			for oID, oNode := range objs {
				obj := oNode.Payload.(*Node)
				l.tree.Objects[oType][oID] = obj
			}
		}
	}
	delete(l.tree.Objects, "Objects")

	if _, ok := l.tree.Objects["LayeredTexture"]; ok {
		fmt.Println("layered textures are not supported. Discarding all but first layer.")
	}
	fmt.Println("Tree:", l.tree.Objects)
	return l.parseTree(textureDir)
}

func (l *Loader) parseTree(textureDir string) (*Scene, error) {
	var err error
	l.connections, err = l.parseConnections()
	if err != nil {
		return nil, err
	}
	fmt.Println("Parsed connections:", l.connections)
	images, err := l.parseImages()
	if err != nil {
		return nil, err
	}
	fmt.Println("Parsed images:", images)
	textures, err := l.parseTextures(images)
	if err != nil {
		return nil, err
	}
	fmt.Println("Parsed textures:", textures)
	materials := l.parseMaterials(textures)
	fmt.Println("Parsed materials:", materials)
	skeletons, morphTargets := l.parseDeformers()
	fmt.Println("Parsed skeletons:", skeletons)
	fmt.Println("Parsed morphTargets:", morphTargets)
	geometry, err := l.parseGeometry(skeletons, morphTargets)
	if err != nil {
		return nil, err
	}

	l.sceneGraph = l.parseScene(skeletons, morphTargets, geometry, materials)
	return l.sceneGraph, nil
}

type VideoNode struct {
	ContentType string
	Filename    string
	Content     io.Reader
}

func NewVideoNode(n *Node) (VideoNode, error) {
	vn := VideoNode{}
	var ok bool
	vn.Filename, ok = n.props["Filename"].Payload.(string)
	if !ok {
		return vn, errors.New("Node lacking VideoNode properties")
	}
	vn.ContentType, ok = n.props["ContentType"].Payload.(string)
	if !ok {
		return vn, errors.New("Node lacking VideoNode properties")
	}
	vn.Content, _ = n.props["Content"].Payload.(io.Reader)
	return vn, nil
}

func (l *Loader) parseImages() (map[IDType]io.Reader, error) {
	fnms := make(map[IDType]string)
	inBlobs := make(map[string]io.Reader)
	outBlobs := make(map[IDType]io.Reader)
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

func (l *Loader) parseTextures(images map[IDType]io.Reader) (map[IDType]Texture, error) {
	txm := make(map[IDType]Texture)
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
	ID      IDType
	name    string
	wrapS   Wrapping
	wrapT   Wrapping
	repeat  floatgeom.Point2
	mapping TextureMapping
	content io.Reader
}

func (l *Loader) parseTexture(tx Node, images map[IDType]io.Reader) (Texture, error) {
	r, err := l.loadTexture(tx, images)
	return Texture{
		ID:      tx.ID,
		name:    tx.attrName,
		wrapS:   tx.props["WrapModeU"].Payload.(Wrapping),
		wrapT:   tx.props["WrapModeV"].Payload.(Wrapping),
		repeat:  tx.props["Scaling"].Payload.(floatgeom.Point2),
		content: r,
	}, err
}

func (l *Loader) parseMaterials(txs map[IDType]Texture) map[IDType]Material {
	mm := make(map[IDType]Material)
	for id, mnt := range l.tree.Objects["Material"] {
		mn := NewMaterialNode(mnt)
		mm[id] = l.parseMaterial(mn, txs)
	}
	return mm
}

func (l *Loader) parseMaterial(mn MaterialNode, txs map[IDType]Texture) Material {
	mat := mn.ShadingModel.Material()
	mat.Name = mn.attrName
	mat = l.parseParameters(mn, txs, mn.ID, mat)
	return mat
}

func (l *Loader) loadTexture(tx Node, images map[IDType]io.Reader) (io.Reader, error) {
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
	v, ok := mn.props["BumpFactor"].Payload.(float64)
	if ok {
		mn.BumpFactor = &v
	}
	v2, ok := mn.props["Diffuse"].Payload.([3]float32)
	if ok {
		mn.Diffuse = &v2
	}
	v3, ok := mn.props["DiffuseColor"].Payload.([3]float32)
	if ok {
		mn.DiffuseColor = &v3
	}
	v4, ok := mn.props["DisplacementFactor"].Payload.(float64)
	if ok {
		mn.DisplacementFactor = &v4
	}
	v5, ok := mn.props["Emissive"].Payload.([3]float32)
	if ok {
		mn.Emissive = &v5
	}
	v6, ok := mn.props["EmissiveColor"].Payload.([3]float32)
	if ok {
		mn.EmissiveColor = &v6
	}
	v7, ok := mn.props["EmissiveFactor"].Payload.(float64)
	if ok {
		mn.EmissiveFactor = &v7
	}
	v8, ok := mn.props["Opacity"].Payload.(float64)
	if ok {
		mn.Opacity = &v8
	}
	v9, ok := mn.props["ReflectionFactor"].Payload.(float64)
	if ok {
		mn.ReflectionFactor = &v9
	}
	v10, ok := mn.props["Shininess"].Payload.(float64)
	if ok {
		mn.Shininess = &v10
	}
	v11, ok := mn.props["Specular"].Payload.([3]float32)
	if ok {
		mn.Specular = &v11
	}
	v12, ok := mn.props["SpecularColor"].Payload.([3]float32)
	if ok {
		mn.SpecularColor = &v12
	}
	v13, ok := mn.props["ShadingModel"].Payload.(ShadingModel)
	if ok {
		mn.ShadingModel = &v13
	}

	return mn
}

func (l *Loader) parseParameters(materialNode MaterialNode, txs map[IDType]Texture, id IDType, mat Material) Material {

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
		txt := txs[child.ID]
		switch child.Property {
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
			fmt.Printf("%s map is not supported in three.js, skipping texture.\n", child.Property)
		}
	}
	return mat
}

func (l *Loader) parseDeformers() (map[IDType]Skeleton, map[IDType]MorphTarget) {
	skeletons := make(map[IDType]Skeleton)
	morphTargets := make(map[IDType]MorphTarget)
	deformer := l.tree.Objects["Deformer"]
	fmt.Println("deformer objects", deformer)
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
	ID            IDType
	Indices       []int32
	Weights       []float64
	Transform     mgl64.Mat4
	TransformLink mgl64.Mat4
	LinkMode      interface{}
}

// Parse single nodes in tree.Objects.Deformer
// The top level skeleton node has type 'Skin' and sub nodes have type 'Cluster'
// Each skin node represents a skeleton and each cluster node represents a bone
func (l *Loader) parseSkeleton(relationships ConnectionSet, deformerNodes map[IDType]*Node) Skeleton {
	rawBones := make([]Bone, 0)
	for _, child := range relationships.children {
		boneNode := deformerNodes[child.ID]
		if boneNode.attrType != "Cluster" {
			continue
		}
		rawBone := Bone{
			ID:      child.ID,
			Indices: []int32{},
			Weights: []float64{},
			// Todo: matrices
			Transform:     Mat4FromSlice(boneNode.props["Transform"].Payload.([]float64)),
			TransformLink: Mat4FromSlice(boneNode.props["TransformLink"].Payload.([]float64)),
			LinkMode:      boneNode.props["Mode"],
		}
		if idxs, ok := boneNode.props["Indexes"]; ok {
			rawBone.Indices = idxs.Payload.([]int32)
			rawBone.Weights = boneNode.props["Weights"].Payload.([]float64)
		}
		rawBones = append(rawBones, rawBone)
	}
	return Skeleton{
		rawBones: rawBones,
		bones:    []BoneModel{},
	}
}

type MorphTarget struct {
	ID         IDType
	RawTargets []RawTarget
}

type RawTarget struct {
	id            IDType
	geoID         IDType
	name          string
	initialWeight float64
	fullWeights   []float64
}

// The top level morph deformer node has type "BlendShape" and sub nodes have type "BlendShapeChannel"
func (l *Loader) parseMorphTargets(relationships ConnectionSet, deformerNodes map[IDType]*Node) []RawTarget {
	rawMorphTargets := make([]RawTarget, 0)
	for i := 0; i < len(relationships.children); i++ {
		child := relationships.children[i]
		morphTargetNode := deformerNodes[child.ID]
		if morphTargetNode.attrType != "BlendShapeChannel" {
			continue
		}
		target := RawTarget{
			name:          morphTargetNode.attrName,
			initialWeight: morphTargetNode.props["DeformPercent"].Payload.(float64),
			id:            morphTargetNode.ID,
			fullWeights:   morphTargetNode.props["FullWeights"].Payload.([]float64),
		}
		for _, child := range l.connections[child.ID].children {
			if child.Property == "" {
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
	skeletons map[IDType]Skeleton,
	morphTargets map[IDType]MorphTarget,
	geometryMap map[IDType]Geometry,
	materialMap map[IDType]Material) *Scene {

	fmt.Println("parse scene start", skeletons, morphTargets, materialMap)

	var sceneGraph Model = NewModelGroup()
	modelMap := l.parseModels(skeletons, geometryMap, materialMap)
	modelNodes := l.tree.Objects["Model"]

	fmt.Println(modelMap, modelNodes)

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
	children := sceneGraph.Children()
	// if all the models where already combined in a single group, just return that
	if len(children) == 1 && children[0].IsGroup() {
		sceneGraph = children[0]
	}
	return &Scene{
		Model:      sceneGraph,
		Animations: l.parseAnimations(),
	}
}

// parse nodes in FBXTree.Objects.Model
func (l *Loader) parseModels(skeletons map[IDType]Skeleton, geometryMap map[IDType]Geometry, materialMap map[IDType]Material) map[IDType]Model {
	modelMap := map[IDType]Model{}
	modelNodes := l.tree.Objects["Model"]
NodeLoop:
	for id, node := range modelNodes {
		fmt.Println("Parsing model:", id, node)
		relationships := l.connections[id]
		var model Model
		m := l.buildSkeleton(relationships, skeletons, id, node.attrName)
		if m != nil {
			model = m
		} else {
			switch node.attrType {
			case "Camera":
				m := l.createCamera(relationships)
				if m == nil {
					continue NodeLoop
				}
				model = m
			case "Light":
				m := l.createLight(relationships)
				if m == nil {
					continue NodeLoop
				}
				model = m
			case "Mesh":
				m := l.createMesh(relationships, geometryMap, materialMap)
				if m == nil {
					continue NodeLoop
				}
				model = m
			case "NurbsCurve":
				m := l.createCurve(relationships, geometryMap)
				if m == nil {
					continue NodeLoop
				}
				model = m
			case "LimbNode": // usually associated with a Bone, however if a Bone was not created we"ll make a Group instead
				fallthrough
			case "Null":
				fallthrough
			default:
				model = NewModelGroup()
			}
			model.SetName(sanitizeNodeName(node.attrName))
			model.SetID(id)
		}
		fmt.Println("model:", model, node.attrType)
		if model == nil {
			fmt.Println("For some reason, a", node.attrType, "model was nil")
			continue
		}
		l.setModelTransforms(model, node)
		modelMap[id] = model
	}
	return modelMap
}

func (l *Loader) buildSkeleton(relationships ConnectionSet, skeletons map[IDType]Skeleton, id IDType, name string) *BoneModel {
	var bone *BoneModel
	for _, parent := range relationships.parents {
		for id, skeleton := range skeletons {
			skeleton.bones = make([]BoneModel, len(skeleton.rawBones))
			for i, rawBone := range skeleton.rawBones {
				if rawBone.ID == parent.ID {
					subBone := bone
					bone = NewBoneModel()
					bone.matrixWorld = rawBone.TransformLink
					// set name and id here - otherwise in cases where "subBone" is created it will not have a name / id
					bone.SetName(sanitizeNodeName(name))
					bone.SetID(id)
					bone.Indices = make([]int, len(rawBone.Indices))
					for j, in := range rawBone.Indices {
						bone.Indices[j] = int(in)
					}
					bone.Weights = rawBone.Weights
					bone.Transform = rawBone.Transform
					skeleton.bones[i] = *bone
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
				cameraAttribute[k] = prop.Payload
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
		typ = 1
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
	lt, ok := n.props["LightType"].Payload.(int)
	if ok {
		v := LightType(lt)
		ln.LightType = &v
	}
	ln.Color, ok = n.props["Color"].Payload.(Color)
	if !ok {
		return nil, errors.New("Invalid LightNode")
	}
	var f float64
	f, ok = n.props["Intensity"].Payload.(float64)
	if ok {
		ln.Intensity = &f
	}
	f, ok = n.props["FarAttenuationEnd"].Payload.(float64)
	if ok {
		ln.FarAttenuationEnd = &f
	}
	f, ok = n.props["InnerAngle"].Payload.(float64)
	if ok {
		ln.InnerAngle = &f
	}
	f, ok = n.props["OuterAngle"].Payload.(float64)
	if ok {
		ln.OuterAngle = &f
	}

	ln.CastLightOnObject, ok = n.props["CastLightOnObject"].Payload.(bool)
	if !ok {
		return nil, errors.New("Invalid LightNode")
	}
	ln.EnableFarAttenuation, ok = n.props["EnableFarAttenuation"].Payload.(bool)
	if !ok {
		return nil, errors.New("Invalid LightNode")
	}
	ln.CastShadows, ok = n.props["CastShadows"].Payload.(bool)
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
	var typ LightType
	// LightType can be undefined for Point lights
	if la.LightType != nil {
		typ = *la.LightType
	}
	var color = la.Color

	intensity := 1.0
	if la.Intensity != nil {
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
	var decay = 1.0
	switch typ {
	case 0: // Point
		model = NewPointLight(color, intensity, distance, decay)
	case 1: // Directional
		model = NewDirectionalLight(color, intensity)
	case 2: // Spot
		angle := math.Pi / 3
		if la.InnerAngle != nil {
			angle = alg.DegToRad * *la.InnerAngle
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
		fmt.Println("THREE.FBXLoader: Unknown light type ", la.LightType, ", defaulting to a THREE.PointLight.")
		model = NewPointLight(color, intensity, 0, 1)
	}
	if la.CastShadows {
		model.SetCastShadow(true)
	}
	return model
}

func (l *Loader) createMesh(relationships ConnectionSet, geometryMap map[IDType]Geometry, materialMap map[IDType]Material) Model {
	var model Model
	var geometry Geometry
	var materials = []*Material{}
	// get geometry and materials(s) from connections
	for _, child := range relationships.children {
		if g, ok := geometryMap[child.ID]; ok {
			geometry = g
		}
		if m, ok := materialMap[child.ID]; ok {
			materials = append(materials, &m)
		}
	}
	if len(materials) == 0 {
		mp := NewMeshPhong()
		mp.color = Color{0.8, 0.8, 0.8}
		materials = []*Material{&mp}
	}
	if len(geometry.Color) > 0 {
		for _, m := range materials {
			m.vertexColors = VertexColors
		}
	}
	if geometry.FBX_Deformer != nil {
		for _, m := range materials {
			m.skinning = true
		}

		fmt.Println("New Skinned Mesh")
		model = NewSkinnedMesh(&geometry, materials)
	} else {
		fmt.Println("New Mesh")
		model = NewMesh(&geometry, materials)
	}
	return model
}

// parse the model node for transform details and apply them to the model
func (l *Loader) setModelTransforms(model Model, modelNode *Node) {
	var td = TransformData{}
	if v, ok := modelNode.props["RotationOrder"]; ok {
		v2 := EulerOrder(v.Payload.(int))
		td.eulerOrder = &v2
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
	fmt.Println("model:", model)
	model.applyMatrix(generateTransform(td))
}

func (l *Loader) createCurve(relationships ConnectionSet, geometryMap map[IDType]Geometry) *Curve {
	var geometry Geometry
	for i := len(relationships.children) - 1; i >= 0; i-- {
		child := relationships.children[i]
		if geo, ok := geometryMap[child.ID]; ok {
			geometry = geo
			break
		}
	}
	// FBX does not list materials for Nurbs lines, so we'll just put our own in here.
	material := NewLineBasicMaterial(Color{0.2, 0, 1}, 1, 1)
	return NewLine(&geometry, material)
}

func (l *Loader) setLookAtProperties(model Model, modelNode *Node) {
	if _, ok := modelNode.props["LookAtProperty"]; ok {
		children := l.connections[model.ID()].children
		for _, child := range children {
			if child.Property == "LookAtProperty" {
				lookAtTarget := l.tree.Objects["Model"][child.ID]
				if prop, ok := lookAtTarget.props["Lcl_Translation"]; ok {
					pos := prop.Payload.([]float64)
					// DirectionalLight, SpotLight
					if tger, ok := model.(Light); ok {
						tger.SetTarget(floatgeom.Point3{pos[0], pos[1], pos[2]})
						//sceneGraph.add(model.target)
					} // else { // Cameras and other Object3Ds
					//	model.lookAt(floatgeom.Point3{pos[0], pos[1], pos[2]})
					//}
				}
			}
		}
	}
}

func (l *Loader) bindSkeleton(skeletons map[IDType]Skeleton, geometryMap map[IDType]Geometry, modelMap map[IDType]Model) {
	bindMatrices := l.parsePoseNodes()
	for _, skeleton := range skeletons {
		for _, parent := range l.connections[skeleton.ID].parents {
			if _, ok := geometryMap[parent.ID]; ok {
				for _, geoConnParent := range l.connections[parent.ID].parents {
					if model, ok := modelMap[geoConnParent.ID]; ok {
						model.BindSkeleton(NewSkeleton(skeleton.bones), bindMatrices[geoConnParent.ID])
					}
				}
			}
		}
	}
}

func (l *Loader) parsePoseNodes() map[IDType]mgl64.Mat4 {
	var bindMatrices = map[IDType]mgl64.Mat4{}
	if BindPoseNode, ok := l.tree.Objects["Pose"]; ok {
		for _, v := range BindPoseNode {
			if v.attrType == "BindPose" {
				poseNodes, ok := v.props["PoseNode"]
				if !ok {
					continue
				}
				for _, n := range poseNodes.Payload.([]*Node) {
					bindMatrices[n.ID] = n.props["Matrix"].Payload.(mgl64.Mat4)
				}
			}
		}
	}
	return bindMatrices
}

// Tree holds a representation of the FBX data, returned by the TextParser ( FBX ASCII format)
// and BinaryParser( FBX Binary format)
type Tree struct {
	Objects map[string]map[IDType]*Node
}

func NewTree() *Tree {
	return &Tree{
		Objects: make(map[string]map[IDType]*Node),
	}
}

var fbxVersionMatch = regexp.MustCompile("FBXVersion: (\\d+)")

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

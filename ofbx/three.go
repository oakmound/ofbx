package threefbx

type Loader struct{
	fbxTree ???
	connections ParsedConnections
	sceneGraph ???
}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Load(r io.Reader, textureDir string) (*Scene, error) {
	var err error
	if isBinary(r) {
		l.fbxTree, err = l.parseBinary(r)
	} else {
		l.fbxTree, err = l.parseASCII(r)
	}
	if err != nil {
		return err
	}
	if l.fbxTree.Objects.LayeredTexture != nil {
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
	deformers, err := l.parseDeformers()
	if err != nil {
		return nil, err
	}
	geometry, err := l.parseGeometry(deformers)
	return l.parseScene(deformers, geometry, materials)
}

type ParsedConnections map[int64]ConnectionSet

type ConnectionSet struct {
	parents []Connection
	children []Connection
}

func NewParsedConnections() ParsedConnections {
	return ParsedConnections(make(map[int64]ConnectionSet))
}

type Connection struct {
	To, From int64
	Relationship string
}

func (l *Loader) parseConnections() ParsedConnections {
	cns := NewParsedConnections()
	for _, cn := range l.fbxTree.connections {
		cns[cn.From].parents = append(cns[cn.From].parents, cn)
		cns[cn.To].children = append(cns[cn.To].children, cn)
	}
	return cns
}

type VideoNode struct {
	ContentType string
	Filename string
	Content io.Reader
}

func (l *Loader) parseImages() (map[int64]io.Reader, error) {
	fnms := make(map[int64]string)
	inBlobs := make(map[string]io.Reader)
	outBlobs := make(map[int64]io.Reader)
	var err error
	for id, v := range l.fbxTree.Videos {
		fnms[id] = v.Filename
		if v.Content != nil {
			inBlobs[v.Filename], err = l.parseImage(l.fbxTree.Videos[nodeID])
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
	for id, txn := range l.fbxTree.Objects.Texture {
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
	RepeatWrapping Wrapping = iota
	ClampToEdgeWrapping Wrapping = iota
)

type Texture struct {
	ID int64
	name string
	wrapS Wrapping
	wrapT Wrapping
	repeat floatgeom.Point2
	content io.Reader
}

func (l *Loader) parseTexture(tx TextureNode, images map[int64]io.Reader) (Texture, error) {
	r, err := l.loadTexture(tx, images)
	return Texture{
		ID: tx.ID,
		name: tx.attrName,
		wrapS: tx.WrapModeU,
		wrapT: tx.WrapModeV,
		repeat: tx.Scaling,
		content: r,
	}, err
}

func (l *Loader) parseMaterials(txs map[int64]Texture) map[int64]Material {
	mm := make(map[int64]Material)
	for id, mn := range l.fbxTree.Objects.Material {
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

func (l *Loader) loadTexture(tx TextureNode, images map[int64]io.Reader) (io.Reader, error) {
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
	BumpFactor *float64
	Diffuse *[3]float32
	DiffuseColor *[3]float32
	DisplacementFactor *float64
	Emissive *[3]float32
	EmissiveColor *[3]float32
	EmissiveFactor *float64
	Opacity *float64
	ReflectionFactor *float64
	Shininess *float64
	Specular *[3]float32
	SpecularColor *[3]float32
}

type MaterialParameters struct {
	BumpFactor float64
	Diffuse Color
	DisplacementFactor float64
	Emissive Color
	EmissiveFactor float64
	Opacity float64
	ReflectionFactor float64
	Shininess float64
	Specular Color
	Transparent bool

	bumpMap Texture
	normalMap Texture
	specularMap Texture
	emissiveMap Texture
	diffuseMap Texture
	alphaMap Texture
	displacementMap Texture
	envMap Texture
} 

func (l *Loader) parseParameters(mn MaterialNode, txs map[int64]Texture, id int64) MaterialParameters {
	parameters := MaterialParameters{}

	if materialNode.BumpFactor != nil  {
		parameters.BumpFactor = *materialNode.BumpFactor
	}
	if materialNode.Diffuse != nil  {
		parameters.Diffuse.R = (*MaterialNode.Diffuse)[0]
		parameters.Diffuse.G = (*MaterialNode.Diffuse)[1]
		parameters.Diffuse.B = (*MaterialNode.Diffuse)[2]
	} else if materialNode.DiffuseColor != nil  {
		// The blender exporter exports diffuse here instead of in materialNode.Diffuse
		parameters.Diffuse.R = (*MaterialNode.DiffuseColor)[0]
		parameters.Diffuse.G = (*MaterialNode.DiffuseColor)[1]
		parameters.Diffuse.B = (*MaterialNode.DiffuseColor)[2]
	}
	if materialNode.DisplacementFactor != nil  {
		parameters.displacementScale = *materialNode.DisplacementFactor;
	}
	if materialNode.Emissive != nil  {
		parameters.Emissive.R = (*MaterialNode.Emissive)[0]
		parameters.Emissive.G = (*MaterialNode.Emissive)[1]
		parameters.Emissive.B = (*MaterialNode.Emissive)[2]
	} else if materialNode.EmissiveColor != nil  {
		// The blender exporter exports emissive color here instead of in materialNode.Emissive
		parameters.Emissive.R = (*MaterialNode.EmissiveColor)[0]
		parameters.Emissive.G = (*MaterialNode.EmissiveColor)[1]
		parameters.Emissive.B = (*MaterialNode.EmissiveColor)[2]
	}
	if materialNode.EmissiveFactor != nil  {
		parameters.EmissiveFactor = *materialNode.EmissiveFactor
	}
	if materialNode.Opacity != nil  {
		parameters.Opacity = *materialNode.Opacity
	}
	if parameters.opacity < 1.0 != nil  {
		parameters.Transparent = true;
	}
	if materialNode.ReflectionFactor != nil  {
		parameters.ReflectionFactor = *materialNode.ReflectionFactor;
	}
	if materialNode.Shininess != nil  {
		parameters.Shininess = *materialNode.Shininess;
	}
	if materialNode.Specular != nil  {
		parameters.Specular.R = (*MaterialNode.Specular)[0]
		parameters.Specular.G = (*MaterialNode.Specular)[1]
		parameters.Specular.B = (*MaterialNode.Specular)[2]
	} else if materialNode.SpecularColor != nil  {
		// The blender exporter exports specular color here instead of in materialNode.Specular
		parameters.Specular.R = (*MaterialNode.SpecularColor)[0]
		parameters.Specular.G = (*MaterialNode.SpecularColor)[1]
		parameters.Specular.B = (*MaterialNode.SpecularColor)[2]
	}

	for _, child := range l.connections[id].children {
		// TODO: Remember to throw away layered things and use the first layer's
		// ID for layered textures
		txt := txs[child.id]
		switch child.Relationship  {
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
				parameters.envMap.mapping = EquirectangularReflectionMapping;
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
				fmt.Printf("%s map is not supported in three.js, skipping texture.\n", child.Relationship )
		}
	}
	return parameters
}

type Skeleton struct {
	ID int64
}
type MorphTarget struct {
	ID int64
}

func (l *Loader) parseDeformers() (map[int64]Skeleton, map[int64]MorphTarget) {
	skeletons := make(map[int64]Skeleton)
	morphTargets := make(map[int64]MorphTarget)
	for id, dn := range l.fbxTree.Objects.Deformer {
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
			mt.rawTargets = l.parseMorphTargets(relationships, l.fbxTree.Objects.Deformer)
			if len(relationships.parents) > 1 {
				fmt.Println("morph target attached to more than one geometry is not supported.")
			}
			morphTargets[id] = mt
		}
	}
	return skeletons, morphTargets
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
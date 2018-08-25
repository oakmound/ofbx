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

type Relationship int 

const (
	Parent Relationship = iota
	Child Relationship = iota
)

type Connection struct {
	To, From int64
	Relationship 
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
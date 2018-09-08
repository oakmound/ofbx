package threefbx

type Node struct {
	ID int64
	attrName string
	attrType string
	name string

	singleProperty bool
	propertyList []Property
}

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

// Parse single nodes in FBXTree.Objects.Deformer
// The top level skeleton node has type 'Skin' and sub nodes have type 'Cluster'
// Each skin node represents a skeleton and each cluster node represents a bone
parseSkeleton: function ( relationships, deformerNodes ) {
	var rawBones = [];
	relationships.children.forEach( function ( child ) {
		var boneNode = deformerNodes[ child.ID ];
		if ( boneNode.attrType !== 'Cluster' ) return;
		var rawBone = {
			ID: child.ID,
			indices: [],
			weights: [],
			transform: new THREE.Matrix4().fromArray( boneNode.Transform.a ),
			transformLink: new THREE.Matrix4().fromArray( boneNode.TransformLink.a ),
			linkMode: boneNode.Mode,
		};
		if ( 'Indexes' in boneNode ) {
			rawBone.indices = boneNode.Indexes.a;
			rawBone.weights = boneNode.Weights.a;
		}
		rawBones.push( rawBone );
	} );
	return {
		rawBones: rawBones,
		bones: []
	};
},
// The top level morph deformer node has type "BlendShape" and sub nodes have type "BlendShapeChannel"
parseMorphTargets: function ( relationships, deformerNodes ) {
	var rawMorphTargets = [];
	for ( var i = 0; i < relationships.children.length; i ++ ) {
		if ( i === 8 ) {
			console.warn( 'FBXLoader: maximum of 8 morph targets supported. Ignoring additional targets.' );
			break;
		}
		var child = relationships.children[ i ];
		var morphTargetNode = deformerNodes[ child.ID ];
		var rawMorphTarget = {
			name: morphTargetNode.attrName,
			initialWeight: morphTargetNode.DeformPercent,
			id: morphTargetNode.id,
			fullWeights: morphTargetNode.FullWeights.a
		};
		if ( morphTargetNode.attrType !== 'BlendShapeChannel' ) return;
		var targetRelationships = connections.get( parseInt( child.ID ) );
		targetRelationships.children.forEach( function ( child ) {
			if ( child.relationship === undefined ) rawMorphTarget.geoID = child.ID;
		} );
		rawMorphTargets.push( rawMorphTarget );
	}
	return rawMorphTargets;
},
// create the main THREE.Group() to be returned by the loader
parseScene: function ( deformers, geometryMap, materialMap ) {
	sceneGraph = new THREE.Group();
	var modelMap = this.parseModels( deformers.skeletons, geometryMap, materialMap );
	var modelNodes = fbxTree.Objects.Model;
	var self = this;
	modelMap.forEach( function ( model ) {
		var modelNode = modelNodes[ model.ID ];
		self.setLookAtProperties( model, modelNode );
		var parentConnections = connections.get( model.ID ).parents;
		parentConnections.forEach( function ( connection ) {
			var parent = modelMap.get( connection.ID );
			if ( parent !== undefined ) parent.add( model );
		} );
		if ( model.parent === null ) {
			sceneGraph.add( model );
		}
	} );
	this.bindSkeleton( deformers.skeletons, geometryMap, modelMap );
	this.createAmbientLight();
	this.setupMorphMaterials();
	var animations = new AnimationParser().parse();
	// if all the models where already combined in a single group, just return that
	if ( sceneGraph.children.length === 1 && sceneGraph.children[ 0 ].isGroup ) {
		sceneGraph.children[ 0 ].animations = animations;
		sceneGraph = sceneGraph.children[ 0 ];
	}
	sceneGraph.animations = animations;
},
// parse nodes in FBXTree.Objects.Model
parseModels: function ( skeletons, geometryMap, materialMap ) {
	var modelMap = new Map();
	var modelNodes = fbxTree.Objects.Model;
	for ( var nodeID in modelNodes ) {
		var id = parseInt( nodeID );
		var node = modelNodes[ nodeID ];
		var relationships = connections.get( id );
		var model = this.buildSkeleton( relationships, skeletons, id, node.attrName );
		if ( ! model ) {
			switch ( node.attrType ) {
				case 'Camera':
					model = this.createCamera( relationships );
					break;
				case 'Light':
					model = this.createLight( relationships );
					break;
				case 'Mesh':
					model = this.createMesh( relationships, geometryMap, materialMap );
					break;
				case 'NurbsCurve':
					model = this.createCurve( relationships, geometryMap );
					break;
				case 'LimbNode': // usually associated with a Bone, however if a Bone was not created we'll make a Group instead
				case 'Null':
				default:
					model = new THREE.Group();
					break;
			}
			model.name = THREE.PropertyBinding.sanitizeNodeName( node.attrName );
			model.ID = id;
		}
		this.setModelTransforms( model, node );
		modelMap.set( id, model );
	}
	return modelMap;
},
buildSkeleton: function ( relationships, skeletons, id, name ) {
	var bone = null;
	relationships.parents.forEach( function ( parent ) {
		for ( var ID in skeletons ) {
			var skeleton = skeletons[ ID ];
			skeleton.rawBones.forEach( function ( rawBone, i ) {
				if ( rawBone.ID === parent.ID ) {
					var subBone = bone;
					bone = new THREE.Bone();
					bone.matrixWorld.copy( rawBone.transformLink );
					// set name and id here - otherwise in cases where "subBone" is created it will not have a name / id
					bone.name = THREE.PropertyBinding.sanitizeNodeName( name );
					bone.ID = id;
					skeleton.bones[ i ] = bone;
					// In cases where a bone is shared between multiple meshes
					// duplicate the bone here and and it as a child of the first bone
					if ( subBone !== null ) {
						bone.add( subBone );
					}
				}
			} );
		}
	} );
	return bone;
},
// create a THREE.PerspectiveCamera or THREE.OrthographicCamera
createCamera: function ( relationships ) {
	var model;
	var cameraAttribute;
	relationships.children.forEach( function ( child ) {
		var attr = fbxTree.Objects.NodeAttribute[ child.ID ];
		if ( attr !== undefined ) {
			cameraAttribute = attr;
		}
	} );
	if ( cameraAttribute === undefined ) {
		model = new THREE.Object3D();
	} else {
		var type = 0;
		if ( cameraAttribute.CameraProjectionType !== undefined && cameraAttribute.CameraProjectionType.value === 1 ) {
			type = 1;
		}
		var nearClippingPlane = 1;
		if ( cameraAttribute.NearPlane !== undefined ) {
			nearClippingPlane = cameraAttribute.NearPlane.value / 1000;
		}
		var farClippingPlane = 1000;
		if ( cameraAttribute.FarPlane !== undefined ) {
			farClippingPlane = cameraAttribute.FarPlane.value / 1000;
		}
		var width = window.innerWidth;
		var height = window.innerHeight;
		if ( cameraAttribute.AspectWidth !== undefined && cameraAttribute.AspectHeight !== undefined ) {
			width = cameraAttribute.AspectWidth.value;
			height = cameraAttribute.AspectHeight.value;
		}
		var aspect = width / height;
		var fov = 45;
		if ( cameraAttribute.FieldOfView !== undefined ) {
			fov = cameraAttribute.FieldOfView.value;
		}
		var focalLength = cameraAttribute.FocalLength ? cameraAttribute.FocalLength.value : null;
		switch ( type ) {
			case 0: // Perspective
				model = new THREE.PerspectiveCamera( fov, aspect, nearClippingPlane, farClippingPlane );
				if ( focalLength !== null ) model.setFocalLength( focalLength );
				break;
			case 1: // Orthographic
				model = new THREE.OrthographicCamera( - width / 2, width / 2, height / 2, - height / 2, nearClippingPlane, farClippingPlane );
				break;
			default:
				console.warn( 'THREE.FBXLoader: Unknown camera type ' + type + '.' );
				model = new THREE.Object3D();
				break;
		}
	}
	return model;
},
// Create a THREE.DirectionalLight, THREE.PointLight or THREE.SpotLight
createLight: function ( relationships ) {
	var model;
	var lightAttribute;
	relationships.children.forEach( function ( child ) {
		var attr = fbxTree.Objects.NodeAttribute[ child.ID ];
		if ( attr !== undefined ) {
			lightAttribute = attr;
		}
	} );
	if ( lightAttribute === undefined ) {
		model = new THREE.Object3D();
	} else {
		var type;
		// LightType can be undefined for Point lights
		if ( lightAttribute.LightType === undefined ) {
			type = 0;
		} else {
			type = lightAttribute.LightType.value;
		}
		var color = 0xffffff;
		if ( lightAttribute.Color !== undefined ) {
			color = new THREE.Color().fromArray( lightAttribute.Color.value );
		}
		var intensity = ( lightAttribute.Intensity === undefined ) ? 1 : lightAttribute.Intensity.value / 100;
		// light disabled
		if ( lightAttribute.CastLightOnObject !== undefined && lightAttribute.CastLightOnObject.value === 0 ) {
			intensity = 0;
		}
		var distance = 0;
		if ( lightAttribute.FarAttenuationEnd !== undefined ) {
			if ( lightAttribute.EnableFarAttenuation !== undefined && lightAttribute.EnableFarAttenuation.value === 0 ) {
				distance = 0;
			} else {
				distance = lightAttribute.FarAttenuationEnd.value;
			}
		}
		// TODO: could this be calculated linearly from FarAttenuationStart to FarAttenuationEnd?
		var decay = 1;
		switch ( type ) {
			case 0: // Point
				model = new THREE.PointLight( color, intensity, distance, decay );
				break;
			case 1: // Directional
				model = new THREE.DirectionalLight( color, intensity );
				break;
			case 2: // Spot
				var angle = Math.PI / 3;
				if ( lightAttribute.InnerAngle !== undefined ) {
					angle = THREE.Math.degToRad( lightAttribute.InnerAngle.value );
				}
				var penumbra = 0;
				if ( lightAttribute.OuterAngle !== undefined ) {
				// TODO: this is not correct - FBX calculates outer and inner angle in degrees
				// with OuterAngle > InnerAngle && OuterAngle <= Math.PI
				// while three.js uses a penumbra between (0, 1) to attenuate the inner angle
					penumbra = THREE.Math.degToRad( lightAttribute.OuterAngle.value );
					penumbra = Math.max( penumbra, 1 );
				}
				model = new THREE.SpotLight( color, intensity, distance, angle, penumbra, decay );
				break;
			default:
				console.warn( 'THREE.FBXLoader: Unknown light type ' + lightAttribute.LightType.value + ', defaulting to a THREE.PointLight.' );
				model = new THREE.PointLight( color, intensity );
				break;
		}
		if ( lightAttribute.CastShadows !== undefined && lightAttribute.CastShadows.value === 1 ) {
			model.castShadow = true;
		}
	}
	return model;
},
createMesh: function ( relationships, geometryMap, materialMap ) {
	var model;
	var geometry = null;
	var material = null;
	var materials = [];
	// get geometry and materials(s) from connections
	relationships.children.forEach( function ( child ) {
		if ( geometryMap.has( child.ID ) ) {
			geometry = geometryMap.get( child.ID );
		}
		if ( materialMap.has( child.ID ) ) {
			materials.push( materialMap.get( child.ID ) );
		}
	} );
	if ( materials.length > 1 ) {
		material = materials;
	} else if ( materials.length > 0 ) {
		material = materials[ 0 ];
	} else {
		material = new THREE.MeshPhongMaterial( { color: 0xcccccc } );
		materials.push( material );
	}
	if ( 'color' in geometry.attributes ) {
		materials.forEach( function ( material ) {
			material.vertexColors = THREE.VertexColors;
		} );
	}
	if ( geometry.FBX_Deformer ) {
		materials.forEach( function ( material ) {
			material.skinning = true;
		} );
		model = new THREE.SkinnedMesh( geometry, material );
	} else {
		model = new THREE.Mesh( geometry, material );
	}
	return model;
},
createCurve: function ( relationships, geometryMap ) {
	var geometry = relationships.children.reduce( function ( geo, child ) {
		if ( geometryMap.has( child.ID ) ) geo = geometryMap.get( child.ID );
		return geo;
	}, null );
	// FBX does not list materials for Nurbs lines, so we'll just put our own in here.
	var material = new THREE.LineBasicMaterial( { color: 0x3300ff, linewidth: 1 } );
	return new THREE.Line( geometry, material );
},
// parse the model node for transform details and apply them to the model
setModelTransforms: function ( model, modelNode ) {
	var transformData = {};
	if ( 'RotationOrder' in modelNode ) transformData.eulerOrder = parseInt( modelNode.RotationOrder.value );
	if ( 'Lcl_Translation' in modelNode ) transformData.translation = modelNode.Lcl_Translation.value;
	if ( 'RotationOffset' in modelNode ) transformData.rotationOffset = modelNode.RotationOffset.value;
	if ( 'Lcl_Rotation' in modelNode ) transformData.rotation = modelNode.Lcl_Rotation.value;
	if ( 'PreRotation' in modelNode ) transformData.preRotation = modelNode.PreRotation.value;
	if ( 'PostRotation' in modelNode ) transformData.postRotation = modelNode.PostRotation.value;
	if ( 'Lcl_Scaling' in modelNode ) transformData.scale = modelNode.Lcl_Scaling.value;
	var transform = generateTransform( transformData );
	model.applyMatrix( transform );
},
setLookAtProperties: function ( model, modelNode ) {
	if ( 'LookAtProperty' in modelNode ) {
		var children = connections.get( model.ID ).children;
		children.forEach( function ( child ) {
			if ( child.relationship === 'LookAtProperty' ) {
				var lookAtTarget = fbxTree.Objects.Model[ child.ID ];
				if ( 'Lcl_Translation' in lookAtTarget ) {
					var pos = lookAtTarget.Lcl_Translation.value;
					// DirectionalLight, SpotLight
					if ( model.target !== undefined ) {
						model.target.position.fromArray( pos );
						sceneGraph.add( model.target );
					} else { // Cameras and other Object3Ds
						model.lookAt( new THREE.Vector3().fromArray( pos ) );
					}
				}
			}
		} );
	}
},
bindSkeleton: function ( skeletons, geometryMap, modelMap ) {
	var bindMatrices = this.parsePoseNodes();
	for ( var ID in skeletons ) {
		var skeleton = skeletons[ ID ];
		var parents = connections.get( parseInt( skeleton.ID ) ).parents;
		parents.forEach( function ( parent ) {
			if ( geometryMap.has( parent.ID ) ) {
				var geoID = parent.ID;
				var geoRelationships = connections.get( geoID );
				geoRelationships.parents.forEach( function ( geoConnParent ) {
					if ( modelMap.has( geoConnParent.ID ) ) {
						var model = modelMap.get( geoConnParent.ID );
						model.bind( new THREE.Skeleton( skeleton.bones ), bindMatrices[ geoConnParent.ID ] );
					}
				} );
			}
		} );
	}
},
parsePoseNodes: function () {
	var bindMatrices = {};
	if ( 'Pose' in fbxTree.Objects ) {
		var BindPoseNode = fbxTree.Objects.Pose;
		for ( var nodeID in BindPoseNode ) {
			if ( BindPoseNode[ nodeID ].attrType === 'BindPose' ) {
				var poseNodes = BindPoseNode[ nodeID ].PoseNode;
				if ( Array.isArray( poseNodes ) ) {
					poseNodes.forEach( function ( poseNode ) {
						bindMatrices[ poseNode.Node ] = new THREE.Matrix4().fromArray( poseNode.Matrix.a );
					} );
				} else {
					bindMatrices[ poseNodes.Node ] = new THREE.Matrix4().fromArray( poseNodes.Matrix.a );
				}
			}
		}
	}
	return bindMatrices;
},
// Parse ambient color in FBXTree.GlobalSettings - if it's not set to black (default), create an ambient light
createAmbientLight: function () {
	if ( 'GlobalSettings' in fbxTree && 'AmbientColor' in fbxTree.GlobalSettings ) {
		var ambientColor = fbxTree.GlobalSettings.AmbientColor.value;
		var r = ambientColor[ 0 ];
		var g = ambientColor[ 1 ];
		var b = ambientColor[ 2 ];
		if ( r !== 0 || g !== 0 || b !== 0 ) {
			var color = new THREE.Color( r, g, b );
			sceneGraph.add( new THREE.AmbientLight( color, 1 ) );
		}
	}
},
setupMorphMaterials: function () {
	sceneGraph.traverse( function ( child ) {
		if ( child.isMesh ) {
			if ( child.geometry.morphAttributes.position || child.geometry.morphAttributes.normal ) {
				var uuid = child.uuid;
				var matUuid = child.material.uuid;
				// if a geometry has morph targets, it cannot share the material with other geometries
				var sharedMat = false;
				sceneGraph.traverse( function ( child ) {
					if ( child.isMesh ) {
						if ( child.material.uuid === matUuid && child.uuid !== uuid ) sharedMat = true;
					}
				} );
				if ( sharedMat === true ) child.material = child.material.clone();
				child.material.morphTargets = true;
			}
		}
	} );
},
};
// FBXTree holds a representation of the FBX data, returned by the TextParser ( FBX ASCII format)
// and BinaryParser( FBX Binary format)
function FBXTree() {}
FBXTree.prototype = {
constructor: FBXTree,
add: function ( key, val ) {
	this[ key ] = val;
},
};


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

function getFbxVersion( text ) {
	var versionRegExp = /FBXVersion: (\d+)/;
	var match = text.match( versionRegExp );
	if ( match ) {
		var version = parseInt( match[ 1 ] );
		return version;
	}
	throw new Error( 'THREE.FBXLoader: Cannot find the version number for the file given.' );
}
// Converts FBX ticks into real time seconds.
function convertFBXTimeToSeconds( time ) {
	return time / 46186158000;
}
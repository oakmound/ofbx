package threefbx

import "github.com/go-gl/mathgl/mgl64"

type DrawMode int

const (
	TrianglesDrawMode     DrawMode = iota
	TriangleStripDrawMode DrawMode = iota
	TriangleFanDrawMode   DrawMode = iota
)

type Mesh struct {
	geometery *Geometry
	materials []*Material
	drawMode  DrawMode
}

func NewMesh(geometery *Geometry, materials []*Material) *Mesh {

	m := Mesh{
		geometery: geometery,
		materials: materials,
		drawMode:  TrianglesDrawMode,
	}

	// m.updateMorphTargets() //Currently think that morph targets does nothing we care about in the scope of FBX

	return &m
}

type SkinnedMesh struct {
	*Mesh
	bindMatrix        mgl64.Mat4
	bindMatrixInverse mgl64.Mat4
	bones             []Model
	skeleton          Skeleton
}

func NewSkinnedMesh(geometery *Geometry, materials []*Material) *SkinnedMesh {

	sm := SkinnedMesh{
		Mesh:  NewMesh(geometery, materials),
		bones: make([]Model, len(geometery.bones)),
	}

	for i, b := range sm.geometery.bones {
		sm.bones[i] = b.(*BoneModel).Copy()
	}

	// sm.skeleton =
	// sm.skeleton = geometery.FBX_Deformer
// var skeleton = new Skeleton( bones );


	updateMatrixWorld(true)

	
	// this.bind( skeleton, this.matrixWorld );
	// this.normalizeSkinWeights();

	return &sm
}

bind: function ( skeleton, bindMatrix ) {

	this.skeleton = skeleton;

	if ( bindMatrix === undefined ) {

		this.updateMatrixWorld( true );

		this.skeleton.calculateInverses();

		bindMatrix = this.matrixWorld;

	}

	this.bindMatrix.copy( bindMatrix );
	this.bindMatrixInverse.getInverse( bindMatrix );

},



// normalizeSkinWeights: function () {

// 	var scale, i;

// 	if ( this.geometry && this.geometry.isGeometry ) {

// 		for ( i = 0; i < this.geometry.skinWeights.length; i ++ ) {

// 			var sw = this.geometry.skinWeights[ i ];

// 			scale = 1.0 / sw.manhattanLength();

// 			if ( scale !== Infinity ) {

// 				sw.multiplyScalar( scale );

// 			} else {

// 				sw.set( 1, 0, 0, 0 ); // do something reasonable

// 			}

// 		}

// 	} else if ( this.geometry && this.geometry.isBufferGeometry ) {

// 		var vec = new Vector4();

// 		var skinWeight = this.geometry.attributes.skinWeight;

// 		for ( i = 0; i < skinWeight.count; i ++ ) {

// 			vec.x = skinWeight.getX( i );
// 			vec.y = skinWeight.getY( i );
// 			vec.z = skinWeight.getZ( i );
// 			vec.w = skinWeight.getW( i );

// 			scale = 1.0 / vec.manhattanLength();

// 			if ( scale !== Infinity ) {

// 				vec.multiplyScalar( scale );

// 			} else {

// 				vec.set( 1, 0, 0, 0 ); // do something reasonable

// 			}

// 			skinWeight.setXYZW( i, vec.x, vec.y, vec.z, vec.w );

// 		}

// 	}

// },

// updateMatrixWorld: function ( force ) {

// 	Mesh.prototype.updateMatrixWorld.call( this, force );

// 	if ( this.bindMode === 'attached' ) {

// 		this.bindMatrixInverse.getInverse( this.matrixWorld );

// 	} else if ( this.bindMode === 'detached' ) {

// 		this.bindMatrixInverse.getInverse( this.bindMatrix );

// 	} else {

// 		console.warn( 'THREE.SkinnedMesh: Unrecognized bindMode: ' + this.bindMode );

// 	}

// },

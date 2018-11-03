package threefbx

import (
	"github.com/go-gl/mathgl/mgl64"
)

type DrawMode int

const (
	TrianglesDrawMode     DrawMode = iota
	TriangleStripDrawMode DrawMode = iota
	TriangleFanDrawMode   DrawMode = iota
)

type Mesh struct {
	*baseModel
	geometry  *Geometry
	materials []*Material
	drawMode  DrawMode
}

func NewMesh(geometry *Geometry, materials []*Material) *Mesh {

	m := Mesh{
		baseModel: &baseModel{},
		geometry:  geometry,
		materials: materials,
		drawMode:  TrianglesDrawMode,
	}

	// m.updateMorphTargets() //Currently think that morph targets does nothing we care about in the scope of FBX

	return &m
}

type SkinnedMesh struct {
	*Mesh
	bound             bool
	bindMatrix        mgl64.Mat4
	bindMatrixInverse mgl64.Mat4
	bones             []Model
	skeleton          *Skeleton
}

func NewSkinnedMesh(geometry *Geometry, materials []*Material) *SkinnedMesh {

	sm := SkinnedMesh{
		Mesh:  NewMesh(geometry, materials),
		bones: make([]Model, len(geometry.FBX_Deformer.bones)),
	}

	for i, b := range sm.geometry.FBX_Deformer.bones {
		sm.bones[i] = b.Copy()
	}

	// sm.skeleton =
	// sm.skeleton = geometery.FBX_Deformer
	// var skeleton = new Skeleton( bones );

	sm.updateMatrixWorld(true)
	sm.bind(geometry.FBX_Deformer, sm.matrixWorld) //TODO: originally this.matrixWorld from js base
	sm.normalizeSkinWeights()
	return &sm
}

func (sm *SkinnedMesh) updateMatrixWorld(force bool) {
	sm.Mesh.updateMatrixWorld(force)
	if sm.bound {
		sm.bindMatrixInverse = sm.matrixWorld.Inv()
	} else {
		sm.bindMatrixInverse = sm.bindMatrix.Inv()
	}
}

func (sm *SkinnedMesh) bind(skeleton *Skeleton, bindMatrix mgl64.Mat4) {
	sm.skeleton = skeleton
	if bindMatrix == (mgl64.Mat4{}) {
		sm.skeleton.calculateInverses()
		sm.bindMatrix = sm.matrixWorld
		sm.bindMatrixInverse = sm.bindMatrix.Inv()
		return
	}
	sm.bindMatrix = bindMatrix
	sm.bindMatrixInverse = bindMatrix.Inv()
}

func (sm *SkinnedMesh) normalizeSkinWeights() {
	//we think that geometry in fbx is only buffergeometries

	// if ( this.geometry && this.geometry.isGeometry ) {
	// 	for ( i = 0; i < this.geometry.skinWeights.length; i ++ ) {
	// 		var sw = this.geometry.skinWeights[ i ];
	// 		scale = 1.0 / sw.manhattanLength();
	// 		if ( scale !== Infinity ) {
	// 			sw.multiplyScalar( scale );
	// 		} else {
	// 			sw.set( 1, 0, 0, 0 ); // do something reasonable
	// 		}
	// 	}
	// } else if ( this.geometry && this.geometry.isBufferGeometry ) {

	skinWeight := sm.geometry.skinWeight
	for i, sw := range skinWeight {
		sum := sw[0] + sw[1] + sw[2] + sw[3]
		sm.geometry.skinWeight[i] = [4]float64{sw[0] / sum, sw[1] / sum, sw[2] / sum, sw[3] / sum}
	}
}

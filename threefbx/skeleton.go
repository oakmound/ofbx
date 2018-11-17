package threefbx

import "github.com/go-gl/mathgl/mgl64"

type Skeleton struct {
	ID         IDType
	geometryID IDType
	// Todo: instead of rawBones and Bones,
	// if rawBones isn't used after it is 'refined'
	// into bones, have a 'processed' bool?
	rawBones     []Bone
	bones        []BoneModel //Should this be bonemodel?
	boneInverses []mgl64.Mat4
}

// calculateInverses taken from https://github.com/mrdoob/three.js/blob/c570b9bd95cf94829715b2cd3a8b128e37768a9c/src/objects/Skeleton.js
func (s *Skeleton) calculateInverses() {
	s.boneInverses = make([]mgl64.Mat4, len(s.bones))
	for i, b := range s.bones {
		s.boneInverses[i] = b.matrixWorld.Inv()
	}
}

func NewSkeleton(bones []BoneModel) *Skeleton {
	sk := &Skeleton{}
	sk.bones = bones
	return sk
}

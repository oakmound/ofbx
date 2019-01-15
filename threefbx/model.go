package threefbx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/oakmound/oak/alg/floatgeom"
)

type Model interface {
	setParent(Model)
	Parent() Model

	Children() []Model
	AddChild(child Model)

	SetName(name string)
	Name() string

	SetID(id IDType)
	ID() IDType

	IsGroup() bool

	MatrixWorld() mgl64.Mat4

	updateMatrixWorld(bool)
	updateMatrix()
	applyMatrix(mgl64.Mat4)

	BindSkeleton(s *Skeleton, mat mgl64.Mat4)
}

type Camera interface {
	Model
	SetFocalLength(int)
}

type ModelCopyable interface {
	Copy() Model
}

type baseModel struct {
	name string
	id   IDType

	parent   Model
	children []Model

	Position   floatgeom.Point3
	quaternion floatgeom.Point4
	scale      floatgeom.Point3

	matrixWorld            mgl64.Mat4
	matrix                 mgl64.Mat4
	matrixWorldNeedsUpdate bool
	matrixAutoUpdate       bool

	skeleton       *Skeleton
	skeletonMatrix mgl64.Mat4
}

// copy expects the real, non base model that this new base model will be
// put into as it's 'wrapping' argument
func (bm *baseModel) copy(wrapping Model) *baseModel {
	bm2 := &baseModel{
		name:     bm.name,
		id:       bm.id,
		parent:   bm.parent,
		children: make([]Model, len(bm.children)),
	}
	for i, c := range bm.children {
		if c2, ok := c.(ModelCopyable); ok {
			c3 := c2.Copy()
			c3.setParent(wrapping)
			bm2.children[i] = c3
		} else {
			fmt.Println(" Tried to copy an uncopiable model, this would normally be an error. #TODO")
		}
	}
	return bm2
}

func SearchModelsByName(root Model, name string) (Model, error) {

	nameFields := strings.Split(root.Name(), "\x00") // TODO: Figure out whether we should be stripping the \x00 \x01 from all names (this shows up currently as a space and smiley which is then postpended with the class name)
	if len(nameFields) != 0 && nameFields[0] == name {
		return root, nil
	}
	for _, c := range root.Children() {
		m, err := SearchModelsByName(c, name)
		if err == nil {
			return m, nil
		}
	}
	return nil, errors.New("Not found")
}

func (bm *baseModel) MatrixWorld() mgl64.Mat4 {
	return bm.matrixWorld
}

func (bm *baseModel) updateMatrixWorld(force bool) {
	if bm.matrixAutoUpdate {
		bm.updateMatrix()
	}

	if bm.matrixWorldNeedsUpdate || force {
		if bm.parent == nil {
			bm.matrixWorld = bm.matrix
		} else {
			bm.matrixWorld = bm.parent.MatrixWorld().Mul4(bm.matrixWorld)
		}
		bm.matrixWorldNeedsUpdate = false
	}

	for _, c := range bm.children {
		c.updateMatrixWorld(force)
	}
}

func (bm *baseModel) updateMatrix() {
	bm.matrix = composeMat(bm.Position, bm.quaternion, bm.scale)
}

func (bm *baseModel) applyMatrix(m2 mgl64.Mat4) {
	bm.matrix = m2.Mul4(bm.matrix)
	//var eul Euler
	bm.Position, _, bm.scale = decomposeMat(bm.matrix)
	//bm.quaternion = eul.ToQuaternion()
}

func (bm *baseModel) BindSkeleton(s *Skeleton, mat mgl64.Mat4) {
	bm.skeleton = s
	bm.skeletonMatrix = mat
}

func (bm *baseModel) Parent() Model {
	return bm.parent
}
func (bm *baseModel) setParent(parent Model) {
	bm.parent = parent
}
func (bm *baseModel) Children() []Model {
	return bm.children
}
func (bm *baseModel) AddChild(ch Model) {
	bm.children = append(bm.children, ch)
	// Note could warn here if child already has parent
	ch.setParent(bm)
}
func (bm *baseModel) SetName(name string) {
	bm.name = name
}
func (bm *baseModel) Name() string {
	return bm.name
}
func (bm *baseModel) SetID(id IDType) {
	bm.id = id
}
func (bm *baseModel) ID() IDType {
	return bm.id
}
func (bm *baseModel) IsGroup() bool {
	panic("baseModel called as full model")
}

func NewModelGroup() *ModelGroup {
	return &ModelGroup{
		baseModel: &baseModel{},
	}
}

type ModelGroup struct {
	*baseModel
}

func (mg *ModelGroup) IsGroup() bool {
	return true
}

func NewPerspectiveCamera(int, int, int, int) *PerspectiveCamera {
	// Not implemented
	return &PerspectiveCamera{
		baseModel: &baseModel{},
	}
}

func (pc *PerspectiveCamera) SetFocalLength(int) {}

type PerspectiveCamera struct {
	*baseModel
}

func (pc *PerspectiveCamera) IsGroup() bool {
	return false
}

func NewOrthographicCamera(int, int, int, int, int, int) *OrthographicCamera {
	// Not implemented
	return &OrthographicCamera{
		baseModel: &baseModel{},
	}
}

func (pc *OrthographicCamera) SetFocalLength(int) {}

type OrthographicCamera struct {
	*baseModel
}

func (oc *OrthographicCamera) IsGroup() bool {
	return false
}

type BoneModel struct {
	*baseModel
	matrixWorld mgl64.Mat4
	Indices     []int
	Weights     []float64
	Transform   mgl64.Mat4
}

func NewBoneModel() *BoneModel {
	return &BoneModel{
		baseModel: &baseModel{},
	}
}

func (bm *BoneModel) IsGroup() bool {
	return false
}

func (bm *BoneModel) Copy() *BoneModel {
	out := &BoneModel{}
	out.baseModel = bm.baseModel.copy(out)
	out.matrixWorld = bm.matrixWorld
	return out
}

type Content struct {
	Meshes        []*Mesh
	SkinnedMeshes []*SkinnedMesh
	Bones         map[string]*BoneModel
}

func newContent() Content {
	return Content{
		Meshes:        make([]*Mesh, 0),
		SkinnedMeshes: make([]*SkinnedMesh, 0),
		Bones:         make(map[string]*BoneModel),
	}
}

func ModelContent(m Model) Content { //TODO: Consider having flag for removal of class names for object's names ie "LeftHandIndex1 ☺Model" becomes "LeftHandIndex1" and "T ☺AnimCurveNode" becomes "T"
	// Parse out Meshes as field
	c := newContent()
	//fmt.Printf("%v:%T \n", m, m)
	switch v := m.(type) {
	case *BoneModel:
		c.Bones[v.Name()] = v
	case *Mesh:
		c.Meshes = []*Mesh{v}
	case *SkinnedMesh:
		c.SkinnedMeshes = []*SkinnedMesh{v}
	}
	for _, child := range m.Children() {
		cnt := ModelContent(child)
		c.Meshes = append(c.Meshes, cnt.Meshes...)
		c.SkinnedMeshes = append(c.SkinnedMeshes, cnt.SkinnedMeshes...)
		for name, b := range cnt.Bones {
			// Note: there's no duplicate checking here-- hypothetically
			// no bones will share names. If things start to go wonky
			// consider adding checks here
			c.Bones[name] = b
		}
	}
	return c
}

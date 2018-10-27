package threefbx

import "github.com/go-gl/mathgl/mgl64"

type Model interface {
	setParent(Model)
	Parent() Model
	Children() []Model
	AddChild(child Model)

	SetAnimations([]Animation)

	SetName(name string)
	Name() string

	SetID(id int)
	ID() int

	IsGroup() bool
}

type Camera interface {
	Model
	SetFocalLength(int)
}

type ModelCopyable interface{
	Copy() Model
}


type baseModel struct {
	name string
	id   int

	parent   Model
	children []Model

	animations []Animation
}

func (bm *baseModel) copy() *baseModel{
	bm2 := baseModel{
		name: bm.name,
		id : bm.id,
		parent: bm.parent,
		children: make([]Model, len(bm.children)),
		animations: make([]Animation, len(bm.animations))
	}
	for i , c := range bm.children{
		if c2, ok :=  c.(ModelCopyable); ok{
			c3 := c2.Copy()
			c3.parent = bm2
			bm2.children[i] = c3
		}
		else{
			fmt.Println(" Tried to copy an uncopiable model, this would normally be an error. #TODO")
		}
	}
	for i , a := range bm.animations{
		bm2.animations[i] = a.Copy()
	}
	return bm2
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
func (bm *baseModel) SetAnimations(anims []Animation) {
	bm.animations = anims
}
func (bm *baseModel) SetName(name string) {
	bm.name = name
}
func (bm *baseModel) Name() string {
	return bm.name
}
func (bm *baseModel) SetID(id int) {
	bm.id = id
}
func (bm *baseModel) ID() int {
	return bm.id
}
func (bm *baseModel) IsGroup() bool {
	panic("baseModel called as full model")
}

type ModelGroup struct {
	*baseModel
}

func (mg *ModelGroup) IsGroup() bool {
	return true
}

func NewPerspectiveCamera(int, int, int, int) *PerspectiveCamera {
	// Not implemented
	return &PerspectiveCamera{}
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
	return &OrthographicCamera{}
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
}

func (bm *BoneModel) IsGroup() bool {
	return false
}

func (bm *BoneModel) Copy() *BoneModel {
	out := BoneModel{
		baseModel: bm.baseModel.copy()
		matrixWorld: bm.matrixWorld
	}
	return out
}

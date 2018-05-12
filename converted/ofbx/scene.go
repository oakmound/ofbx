package ofbx

type Scene struct{}

type IScene struct{}

func (is *IScene) getRootElement() *IElement {
	return nil
}
func (is *IScene) getRoot() *Object {
	return nil
}
func (is *IScene) getTakeInfo(name string) *TakeInfo {
	return nil
}
func (is *IScene) getSceneFrameRate() float32 {
	return 0
}
func (is *IScene) getMesh(int index) []Mesh {
	return nil
}
func (is *IScene) getAnimationStack(index int) []AnimationStack {
	return nil
}
func (is *IScene) getAllObjects() []Object {
	return nil
}

func load(data []byte) *Iscene {
	return nil
}

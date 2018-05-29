package ofbx

type Connection struct {
	ConnectionType
	typ      Type
	from, to uint64
	property DataView
}

type ConnectionType int

// Connection Types
const (
	OBJECT_OBJECT   ConnectionType = iota
	OBJECT_PROPERTY ConnectionType = iota
)

type ObjectPair struct {
	element *Element
	object  *Object
}

//const GlobalSettings* getGlobalSettings() const override { return &m_settings; }

type Scene struct {
	m_root_element     *Element
	m_root             *Root
	m_scene_frame_rate float32 // = -1
	m_settings         GlobalSettings
	m_object_map       map[uint64]ObjectPair // Slice or map?
	m_all_objects      []*Object
	m_meshes           []*Mesh
	m_animation_stacks []*AnimationStack
	m_connections      []Connection
	m_data             []byte
	m_take_infos       []TakeInfo
}

func (s *Scene) getRootElement() *Element {
	return s.m_root_element
}
func (s *Scene) getRoot() *Object {
	return s.m_root
}
func (s *Scene) getTakeInfo(name string) *TakeInfo {
	for _, info := range m_take_infos {
		if info.name == name {
			return &info
		}
	}
	return nil
}
func (s *Scene) getSceneFrameRate() float32 {
	return s.m_scene_frame_rate
}
func (s *Scene) getMesh(index int) *Mesh {
	//assert(index >= 0);
	//assert(index < m_meshes.size());
	return m_meshes[index]
}
func (s *Scene) getAnimationStack(index int) *AnimationStack {
	//assert(index >= 0);
	//assert(index < m_animation_stacks.size());
	return m_animation_stacks[index]

}
func (s *Scene) getAllObjects() []Object {
	return s.m_all_objects
}

func load(data []byte) *Scene {
	s := &Scene{}
	s.m_data = make([]byte, len(data))
	copy(s.m_data, data)

	root, err := tokenize(s.m_data)
	if err != nil {
		root, err = tokenizeText(s.m_data)
		if err != nil {
			return nil
		}
	}

	scene.m_root_element = root.getValue()
	//assert(scene.m_root_element);

	// This was commented out already I didn't do it
	//if (parseTemplates(*root.getValue()).isError()) return nil
	if !parseConnections(*root.getValue(), scene.get()) {
		return nil
	}
	if !parseTakes(scene.get()) {
		return nil
	}
	if !parseObjects(*root.getValue(), scene.get()) {
		return nil
	}
	parseGlobalSettings(*root.getValue(), scene.get())

	return scene.release()
}

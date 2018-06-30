package ofbx

import (
	"fmt"
	"io"
)

type Connection struct {
	typ      ConnectionType
	from, to uint64
	property string
}

type ConnectionType int

// Connection Types
const (
	OBJECT_OBJECT   ConnectionType = iota
	OBJECT_PROPERTY ConnectionType = iota
)

type ObjectPair struct {
	element *Element
	object  Obj
}

//const GlobalSettings* getGlobalSettings() const override { return &m_settings; }

type Scene struct {
	m_root_element     *Element
	m_root             *Node
	m_scene_frame_rate float32 // = -1
	m_settings         GlobalSettings
	m_object_map       map[uint64]ObjectPair // Slice or map?
	m_all_objects      []Obj
	m_meshes           []*Mesh
	m_animation_stacks []*AnimationStack
	m_connections      []Connection
	m_data             []byte
	m_take_infos       []TakeInfo
}

func (s *Scene) getRootElement() *Element {
	return s.m_root_element
}
func (s *Scene) getRoot() Obj {
	return s.m_root
}
func (s *Scene) getTakeInfo(name string) *TakeInfo {
	for _, info := range s.m_take_infos {
		if info.name.String() == name {
			return &info
		}
	}
	return nil
}
func (s *Scene) getSceneFrameRate() float32 {
	return s.m_scene_frame_rate
}
func (s *Scene) getMesh(index int) *Mesh {
	return s.m_meshes[index]
}

func (s *Scene) GetAnimationStacks() []*AnimationStack {
	return s.m_animation_stacks
}

func (s *Scene) GetAnimationStack(index int) *AnimationStack {
	return s.m_animation_stacks[index]

}
func (s *Scene) getAllObjects() []Obj {
	return s.m_all_objects
}

func Load(r io.Reader) (*Scene, error) {
	s := &Scene{}
	s.m_object_map = make(map[uint64]ObjectPair)
	fmt.Println("Starting tokenize")
	root, err := tokenize(r)
	fmt.Println("Tokenize completed")
	if err != nil {
		//fmt.Println("Starting TokenizeText")
		//TODO: reimplement
		// root, err = tokenizeText(r)
		// fmt.Println("TokenizeText completed")
		if err != nil {
			return nil, err
		}
	}

	s.m_root_element = root
	//assert(scene.m_root_element);

	// This was commented out already I didn't do it
	//if (parseTemplates(*root.getValue()).isError()) return nil
	fmt.Println("Starting parse connection")
	if ok, err := parseConnection(root, s); !ok {
		return nil, err
	}
	fmt.Println("Starting parse takes")
	if ok, err := parseTakes(s); !ok {
		return nil, err
	}
	fmt.Println("Starting parse objects")
	if ok, err := parseObjects(root, s); !ok {
		return nil, err
	}
	fmt.Println("Parsing global settings")
	parseGlobalSettings(root, s)

	return s, nil
}

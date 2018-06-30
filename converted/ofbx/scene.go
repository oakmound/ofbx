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
	RootElement     *Element
	RootNode        *Node
	frameRate       float32 // = -1
	settings        Settings
	objectMap       map[uint64]ObjectPair // Slice or map?
	Objects         []Obj
	meshes          []*Mesh
	AnimationStacks []*AnimationStack
	connections     []Connection
	takeInfos       []TakeInfo
}

func (s *Scene) String() {
	st := "Scene: "
	st += "element=" + s.RootElement.String()
	st += ", root=" + s.RootNode.String()
	st += ", frameRate=" + fmt.Sprintf("%f", s.frameRate)
	st += ", setttings=" + s.settings.String()
	st += ", objects=" + fmt.Sprint(s.Objects)
	st += ", meshes=" + fmt.Sprint(s.meshes)
	st += ", animations=" + fmt.Sprint(s.AnimationStacks)
	st += ", connections=" + fmt.Sprint(s.connections)
	st += ", takeInfos=" + fmt.Sprint(s.takeInfos)
}

func (s *Scene) getTakeInfo(name string) *TakeInfo {
	for _, info := range s.takeInfos {
		if info.name.String() == name {
			return &info
		}
	}
	return nil
}
func (s *Scene) getSceneFrameRate() float32 {
	return s.frameRate
}
func (s *Scene) getMesh(index int) *Mesh {
	return s.meshes[index]
}

func Load(r io.Reader) (*Scene, error) {
	s := &Scene{}
	s.objectMap = make(map[uint64]ObjectPair)
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

	s.RootElement = root

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

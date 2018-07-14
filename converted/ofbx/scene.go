package ofbx

import (
	"fmt"
	"io"
)

type ObjectPair struct {
	element *Element
	object  Obj
}

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

var (
	printRecursiveObjects = false
)

func (s *Scene) String() string {
	if s == nil {
		return "nil Scene"
	}
	st := "Scene: " + "\n"
	// if s.RootElement != nil {
	// 	st += "element=" + s.RootElement.String()
	// }
	// if s.RootNode != nil {
	// 	st += "root=" + s.RootNode.String() + "\n"
	// }
	st += "frameRate=" + fmt.Sprintf("%f", s.frameRate) + "\n"
	st += "setttings=" + fmt.Sprintf("%+v", s.settings) + "\n"
	if s.Objects != nil {
		st += "objects=" + "\n" //perhaps this should output the object types?
	}
	for _, o := range s.Objects {
		st += o.stringPrefix("\t") + "\n"
	}

	if s.meshes != nil {
		st += "meshes=" + fmt.Sprint(s.meshes) + "\n"
	}
	// if s.AnimationStacks != nil {
	// 	st += "animations=" + fmt.Sprint(s.AnimationStacks) + "\n"
	// }
	if len(s.connections) > 0 {
		st += "connections=" + "\n"
		for _, c := range s.connections {
			st += "\t" + c.String() + "\n"
		}
	}
	if len(s.takeInfos) > 0 {
		st += "takeInfos=" + "\n"
		for _, tk := range s.takeInfos {
			st += "\t" + tk.String()
		}
	}
	return st
}

func (s *Scene) Geometries() []*Geometry {
	out := make([]*Geometry, 0)
	for _, o := range s.Objects {
		elem := o.Element()
		if elem == nil {
			continue
		}
		if elem.ID.String() == "Geometry" {
			out = append(out, o.(*Geometry))
		}
	}
	return out
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
	root, err := tokenize(r)
	if err != nil {
		//TODO: reimplement
		// root, err = tokenizeText(r)
		if err != nil {
			return nil, err
		}
	}

	s.RootElement = root

	if ok, err := parseConnection(root, s); !ok {
		return nil, err
	}
	if ok, err := parseTakes(s); !ok {
		return nil, err
	}
	if ok, err := parseObjects(root, s); !ok {
		return nil, err
	}
	parseGlobalSettings(root, s)

	return s, nil
}

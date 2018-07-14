package ofbx

import "fmt"

type AnimationStack struct {
	Object
	Layers []*AnimationLayer
}

func NewAnimationStack(scene *Scene, element *Element) *AnimationStack {
	o := *NewObject(scene, element)
	return &AnimationStack{o, []*AnimationLayer{}}
}

func (as *AnimationStack) Type() Type {
	return ANIMATION_STACK
}

func (as *AnimationStack) getLayer(index int) *AnimationLayer {
	recv := resolveObjectLinkIndex(as, index)
	if recv == nil {
		return nil
	}
	al, _ := recv.(*AnimationLayer)
	return al
}

func (as *AnimationStack) getAllLayers() []*AnimationLayer {
	objs := resolveAllObjectLinks(as)
	als := []*AnimationLayer{}
	for _, obj := range objs {
		if obj == nil {
			continue
		}
		al, ok := obj.(*AnimationLayer)
		if ok {
			als = append(als, al)
		}
	}
	return als
}

func (as *AnimationStack) postProcess() bool {
	as.Layers = as.getAllLayers()
	return true
}

func (as *AnimationStack) String() string {
	return as.stringPrefix("")
}

func (as *AnimationStack) stringPrefix(prefix string) string {
	s := prefix + "AnimationStack:" + fmt.Sprintf("%v", as.ID())
	if len(as.Layers) > 0 {
		for _, lay := range as.Layers {
			s += "\n" + lay.stringPrefix(prefix+"\t")
		}
	} else {
		s += "<empty>"
	}
	return s + "\n"
}

type AnimationLayer struct {
	Object
	CurveNodes []*AnimationCurveNode
}

func NewAnimationLayer(scene *Scene, element *Element) *AnimationLayer {
	o := *NewObject(scene, element)
	return &AnimationLayer{o, nil}
}

func (as *AnimationLayer) Type() Type {
	return ANIMATION_LAYER
}

func (as *AnimationLayer) getCurveNode(bone Obj, property string) *AnimationCurveNode {
	for _, node := range as.CurveNodes {
		if node.boneLinkProp == property && node.Bone == bone {
			return node
		}
	}
	return nil
}

func (as *AnimationLayer) String() string {
	return as.stringPrefix("")
}

func (as *AnimationLayer) stringPrefix(prefix string) string {
	s := prefix + "AnimationLayer:" + fmt.Sprintf("%v", as.ID())
	if len(as.CurveNodes) != 0 {
		for _, curve := range as.CurveNodes {
			s += "\n" + curve.stringPrefix(prefix+"\t")
		}
	} else {
		s += "<empty>"
	}
	return s

}

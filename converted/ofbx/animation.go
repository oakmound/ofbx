package ofbx

import "fmt"

// An AnimationStack is collection of 1 to n AnimationLayers along with possibily some properties
type AnimationStack struct {
	Object
	Layers []*AnimationLayer
}

// NewAnimationStack creates a new empty stack
func NewAnimationStack(scene *Scene, element *Element) *AnimationStack {
	o := *NewObject(scene, element)
	return &AnimationStack{o, []*AnimationLayer{}}
}

// Type returns the Animation_stack type
func (as *AnimationStack) Type() Type {
	return ANIMATION_STACK
}

// getLayer returns a requested animationlayer from the stack. TODO: Verify that we can fully remove this.
func (as *AnimationStack) getLayer(index int) *AnimationLayer {
	recv := resolveObjectLinkIndex(as, index)
	if recv == nil {
		return nil
	}
	al, _ := recv.(*AnimationLayer)
	return al
}

// getAllLayers returns gets all linked layers and creates a slice out of them
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

// postProcess places the processed AnimationLayers into the slice on AnimationStack
func (as *AnimationStack) postProcess() bool {
	as.Layers = as.getAllLayers()
	return true
}

// String returns a pretty print version of AnimationStack
func (as *AnimationStack) String() string {
	return as.stringPrefix("")
}

// stringPrefix returns a string with formatting and indentation based on the given prefix
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

// AnimationLayer is a collection of AnimationCurveNodes along with possibily some properties
type AnimationLayer struct {
	Object
	CurveNodes []*AnimationCurveNode
}

// NewAnimationLayer creates a new AnimationLayer with no curvenodes
func NewAnimationLayer(scene *Scene, element *Element) *AnimationLayer {
	o := *NewObject(scene, element)
	return &AnimationLayer{o, nil}
}

// Type returns the type of Animation_layer
func (as *AnimationLayer) Type() Type {
	return ANIMATION_LAYER
}

// getCuurveNode gets the first curvenode with the given bone and property
func (as *AnimationLayer) getCurveNode(bone Obj, property string) *AnimationCurveNode {
	for _, node := range as.CurveNodes {
		if node.BoneLinkProp == property && node.Bone == bone {
			return node
		}
	}
	return nil
}

// String pretty formats the AnimationLayers Curve Nodes
func (as *AnimationLayer) String() string {
	return as.stringPrefix("")
}

// stringPrefiox pretty formats the AnimationLayers Curve Nodes along with formatting based on the given prefix
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

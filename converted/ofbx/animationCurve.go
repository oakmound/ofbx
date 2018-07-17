package ofbx

import (
	"fmt"
	"strings"

	"github.com/oakmound/oak/alg/floatgeom"
)

// CurveMode details how the values in the CurveNode should be interperted and used (translation,rotation,scale)
type CurveMode int

const (
	// TRANSLATION is a CurveMode where the curve applies a Translation
	TRANSLATION CurveMode = iota
	// ROTATION is a CurveMode where the curve applies a rotation
	ROTATION CurveMode = iota
	// SCALE is a CurveMode where the curve applies a scaling effect
	SCALE CurveMode = iota
)

// AnimationCurve are a mapping of key frame times to a set of data. Basic need is to have Times mapping with Values while disregarding AttributeFlags and Data
type AnimationCurve struct {
	Object
	Times        []int64
	Values       []float32
	AttrFlags    []int64
	AttrData     []float32
	AttrRefCount []int64
}

// NewAnimationCurve creates a new stub AnimationCurve
func NewAnimationCurve(scene *Scene, element *Element) *AnimationCurve {
	o := *NewObject(scene, element)
	return &AnimationCurve{Object: o}
}

// Type returns the type ANIMATION_CURVE
func (ac *AnimationCurve) Type() Type {
	return ANIMATION_CURVE
}

// String pretty prints out the AnimationCurve
func (ac *AnimationCurve) String() string {
	return ac.stringPrefix("")
}

// stringPrefix pretty prints out the AnimationCurve while respecting the prefix indentation
func (ac *AnimationCurve) stringPrefix(prefix string) string {

	strs := make([]string, len(ac.Times))
	for i := 0; i < len(ac.Times); i++ {
		strs[i] = fmt.Sprintf("%d:%f", ac.Times[i], ac.Values[i])
	}
	return prefix + "AnimCurve: " + strings.Join(strs, ",") + " "
}

// AnimationCurveNode is a mapping of Curve to a property
type AnimationCurveNode struct {
	Object
	curves       [3]Curve
	Bone         Obj
	boneLinkProp string
	mode         CurveMode
}

// Curve is a connection Linkage for an AnimationCurve
type Curve struct {
	curve      *AnimationCurve
	connection *Connection
}

// String pretty formats the Curve
func (c *Curve) String() string {
	s := c.curve.String() + " "
	//s += c.connection.String()
	return s
}

// NewAnimationCurveNode creates a new AnimationCurveNode with just the base object
func NewAnimationCurveNode(s *Scene, e *Element) *AnimationCurveNode {
	acn := AnimationCurveNode{}
	obj := *NewObject(s, e)
	acn.Object = obj
	return &acn
}

//Type returns ANIMATION_CURVE_NODE
func (acn *AnimationCurveNode) Type() Type {
	return ANIMATION_CURVE_NODE
}

func (acn *AnimationCurveNode) getNodeLocalTransform(time float64) floatgeom.Point3 {
	fbxTime := secondsToFbxTime(time)

	getCoord := func(curve *Curve, fbxTime int64) float32 {
		if curve.curve == nil {
			return 0.0
		}

		times := curve.curve.Times
		values := curve.curve.Values
		count := len(times)

		if fbxTime < times[0] {
			fbxTime = times[0]
		}
		if fbxTime > times[count-1] {
			fbxTime = times[count-1]
		}
		for i := 1; i < count; i++ {
			if times[i] >= fbxTime {
				t := float32(float64(fbxTime-times[i-1]) / float64(times[i]-times[i-1]))
				return values[i-1]*(1-t) + values[i]*t
			}
		}
		return values[0]
	}

	return floatgeom.Point3{
		float64(getCoord(&acn.curves[0], fbxTime)),
		float64(getCoord(&acn.curves[1], fbxTime)),
		float64(getCoord(&acn.curves[2], fbxTime)),
	}
}

// String pretty formats the AnimationCurveNode
func (acn *AnimationCurveNode) String() string {
	return acn.stringPrefix("")
}

// stringPrefix pretty formats the AnimationCurveNode while also respecting indentation dictated by the prefix
func (acn *AnimationCurveNode) stringPrefix(prefix string) string {
	s := prefix + "AnimationCurveNode: "
	if acn.Bone != nil {
		s += "boneID=" + fmt.Sprintf("%v ", acn.Bone.ID())
	}
	s += "bone_link_property=\"" + acn.boneLinkProp + "\""
	s += " mode=" + fmt.Sprintf("%d", acn.mode)
	for i, curve := range acn.curves {
		s += "\n" + prefix + "\t"
		switch i {
		case 0:
			s += " X: "
		case 1:
			s += " Y: "
		case 2:
			s += " Z: "
		}
		s += curve.String()
	}

	return s
}

package ofbx

import (
	"fmt"
	"strings"
	"time"

	"github.com/oakmound/oak/v2/alg/floatgeom"
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

const (
	BoneTranslate = "Lcl Translation"
	BoneRotate    = "Lcl Rotation"
	BoneScale     = "Lcl Scaling"
)

// AnimationCurve are a mapping of key frame times to a set of data. Basic need is to have Times mapping with Values while disregarding AttributeFlags and Data
type AnimationCurve struct {
	Object
	Times        []time.Duration
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
		strs[i] = fmt.Sprintf("%v:%f", ac.Times[i], ac.Values[i])
	}
	return prefix + "AnimCurve: " + strings.Join(strs, ",") + " "
}

// AnimationCurveNode is a mapping of Curve to a property
type AnimationCurveNode struct {
	Object
	Curves       [3]Curve
	Bone         Obj
	BoneLinkProp string
	mode         CurveMode
}

// Curve is a connection Linkage for an AnimationCurve
type Curve struct {
	Curve      *AnimationCurve
	connection *Connection
}

// String pretty formats the Curve
func (c *Curve) String() string {
	s := c.Curve.String() + " "
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

func (acn *AnimationCurveNode) getNodeLocalTransform(t float64) floatgeom.Point3 {
	fbxTime := fbxTimetoStdTime(secondsToFbxTime(t))

	getCoord := func(curve *Curve, fbxTime time.Duration) float32 {
		if curve.Curve == nil {
			return 0.0
		}

		times := curve.Curve.Times
		values := curve.Curve.Values
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
		float64(getCoord(&acn.Curves[0], fbxTime)),
		float64(getCoord(&acn.Curves[1], fbxTime)),
		float64(getCoord(&acn.Curves[2], fbxTime)),
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
	s += "bone_link_property=\"" + acn.BoneLinkProp + "\""
	s += " mode=" + fmt.Sprintf("%d", acn.mode)

	// NOTE: an animation curve node which specify the focal length property has
	// exactly one curve (FocalLength), rather than three (X,Y,Z).
	if strings.HasPrefix(acn.Object.name, "FocalLength") {
		s += "\n" + prefix + "\t"
		s += " FocalLength: "
		s += acn.Curves[0].String()
		return s
	}

	for i, curve := range acn.Curves {
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

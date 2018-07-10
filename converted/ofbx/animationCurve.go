package ofbx

import (
	"fmt"
	"strings"

	"github.com/oakmound/oak/alg/floatgeom"
)

type CurveMode int

const (
	TRANSLATION CurveMode = iota
	ROTATION    CurveMode = iota
	SCALE       CurveMode = iota
)

type AnimationCurve struct {
	Object
	times  []int64
	values []float32
}

func NewAnimationCurve(scene *Scene, element *Element) *AnimationCurve {
	o := *NewObject(scene, element)
	return &AnimationCurve{o, nil, nil}
}

func (ac *AnimationCurve) Type() Type {
	return ANIMATION_CURVE
}

func (ac *AnimationCurve) getKeyTime() []int64 {
	return ac.times
}

func (ac *AnimationCurve) getKeyValue() []float32 {
	return ac.values
}

func (ac *AnimationCurve) String() string {
	return ac.stringPrefix("")
}

func (ac *AnimationCurve) stringPrefix(prefix string) string {

	strs := make([]string, len(ac.times))
	for i := 0; i < len(ac.times); i++ {
		strs[i] = fmt.Sprintf("%d:%f", ac.times[i], ac.values[i])
	}
	return prefix + "AnimCurve: " + strings.Join(strs, ",") + " "
}

type AnimationCurveNode struct {
	Object
	curves             [3]Curve
	bone               Obj
	bone_link_property string
	mode               CurveMode
}

type Curve struct {
	curve      *AnimationCurve
	connection *Connection
}

func (c *Curve) String() string {
	s := c.curve.String() + " "
	s += c.connection.String()
	return s
}

func NewAnimationCurveNode(s *Scene, e *Element) *AnimationCurveNode {

	acn := AnimationCurveNode{}
	obj := *NewObject(s, e)
	acn.Object = obj
	return &acn
}

func (acn *AnimationCurveNode) Type() Type {
	return ANIMATION_CURVE_NODE
}

func (acn *AnimationCurveNode) getNodeLocalTransform(time float64) floatgeom.Point3 {
	fbx_time := secondsToFbxTime(time)

	getCoord := func(curve *Curve, fbx_time int64) float32 {
		if curve.curve == nil {
			return 0.0
		}

		times := curve.curve.getKeyTime()
		values := curve.curve.getKeyValue()
		count := len(times)

		if fbx_time < times[0] {
			fbx_time = times[0]
		}
		if fbx_time > times[count-1] {
			fbx_time = times[count-1]
		}
		for i := 1; i < count; i++ {
			if times[i] >= fbx_time {
				t := float32(float64(fbx_time-times[i-1]) / float64(times[i]-times[i-1]))
				return values[i-1]*(1-t) + values[i]*t
			}
		}
		return values[0]
	}

	return floatgeom.Point3{
		float64(getCoord(&acn.curves[0], fbx_time)),
		float64(getCoord(&acn.curves[1], fbx_time)),
		float64(getCoord(&acn.curves[2], fbx_time)),
	}
}

func (acn *AnimationCurveNode) getBone() Obj {
	return acn.bone
}

func (acn *AnimationCurveNode) String() string {
	return acn.stringPrefix("")
}
func (acn *AnimationCurveNode) stringPrefix(prefix string) string {
	s := prefix + "AnimationCurveNode: "
	if printRecursiveObjects {
		s += "\n\tbone=" + acn.bone.stringPrefix(prefix) + "\n"
	}
	s += "bone_link_property=" + acn.bone_link_property
	s += " mode=" + fmt.Sprintf("%d", acn.mode)
	s += acn.Object.stringPrefix(prefix)
	s += prefix + "curves="
	for _, curve := range acn.curves {
		s += "\n" + prefix + "\t" + curve.String()
	}

	return s
}

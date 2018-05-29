package ofbx

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
	return &AnimationCurve{}
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

type AnimationCurveNode struct {
	Object
	curves             [3]Curve
	bone               *Object
	bone_link_property string
	mode               CurveMode
}

type Curve struct {
	curve      *AnimationCurve
	connection *Connection
}

func (acn *AnimationCurveNode) Type() Type {
	return ANIMATION_CURVE_NODE
}

func (acn *AnimationCurveNode) getNodeLocalTransform(time float64) Vec3 {
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

	return Vec3{
		float64(getCoord(&acn.curves[0], fbx_time)),
		float64(getCoord(&acn.curves[1], fbx_time)),
		float64(getCoord(&acn.curves[2], fbx_time)),
	}
}

func (acn *AnimationCurveNode) getBone() *Object {
	return acn.bone
}

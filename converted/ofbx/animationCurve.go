package ofbx

type AnimationCurve struct {
	Object
}

func NewAnimationCurve(scene *Scene, element *IElement) *AnimationCurve {
	return nil
}

func (ac *AnimationCurve) Type() Type {
	return ANIMATION_CURVE
}

// 200sc note: this may be the length of the next two functions, i.e. unneeded
func (ac *AnimationCurve) getKeyCount() int {
	return 0
}

func (ac *AnimationCurve) getKeyTime() *int64 {
	return nil
}

func (ac *AnimationCurve) getKeyValue() *float32 {
	return nil
}

type AnimationCurveNode struct {
	Object
}

func (acn *AnimationCurveNode) Type() Type {
	return ANIMATION_CURVE_NODE
}

func (acn *AnimationCurveNode) getNodeLocalTransform(time float64) Vec3 {
	return Vec3{}
}

func (acn *AnimationCurveNode) getBone() *Object {
	return nil
}

struct AnimationCurveImpl : AnimationCurve
{
	AnimationCurveImpl(const Scene& _scene, const IElement& _element)
		: AnimationCurve(_scene, _element)
	{
	}

	int getKeyCount() const override { return (int)times.size(); }
	const int64* getKeyTime() const override { return &times[0]; }
	const float* getKeyValue() const override { return &values[0]; }

	std::vector<int64> times;
	std::vector<float> values;
	Type getType() const override { return Type::ANIMATION_CURVE; }
};


struct AnimationCurveNodeImpl : AnimationCurveNode
{
	AnimationCurveNodeImpl(const Scene& _scene, const IElement& _element)
		: AnimationCurveNode(_scene, _element)
	{
	}


	const Object* getBone() const override
	{
		return bone;
	}


	Vec3 getNodeLocalTransform(double time) const override
	{
		int64 fbx_time = secondsToFbxTime(time);

		auto getCoord = [](const Curve& curve, int64 fbx_time) {
			if (!curve.curve) return 0.0f;

			const int64* times = curve.curve.getKeyTime();
			const float* values = curve.curve.getKeyValue();
			int count = curve.curve.getKeyCount();

			if (fbx_time < times[0]) fbx_time = times[0];
			if (fbx_time > times[count - 1]) fbx_time = times[count - 1];
			for (int i = 1; i < count; ++i)
			{
				if (times[i] >= fbx_time)
				{
					float t = float(double(fbx_time - times[i - 1]) / double(times[i] - times[i - 1]));
					return values[i - 1] * (1 - t) + values[i] * t;
				}
			}
			return values[0];
		};

		return {getCoord(curves[0], fbx_time), getCoord(curves[1], fbx_time), getCoord(curves[2], fbx_time)};
	}


	struct Curve
	{
		const AnimationCurve* curve = nullptr;
		const Scene::Connection* connection = nullptr;
	};


	Curve curves[3];
	Object* bone = nullptr;
	DataView bone_link_property;
	Type getType() const override { return Type::ANIMATION_CURVE_NODE; }
	enum Mode
	{
		TRANSLATION,
		ROTATION,
		SCALE
	} mode = TRANSLATION;
};
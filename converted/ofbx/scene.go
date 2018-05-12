package ofbx

type Scene struct{}
struct Scene : IScene
{
	struct Connection
	{
		enum Type
		{
			OBJECT_OBJECT,
			OBJECT_PROPERTY
		};

		Type type;
		uint64 from;
		uint64 to;
		DataView property;
	};

	struct ObjectPair
	{
		const Element* element;
		Object* object;
	};


	int getAnimationStackCount() const override { return (int)m_animation_stacks.size(); }
	int getMeshCount() const override { return (int)m_meshes.size(); }
	float getSceneFrameRate() const override { return m_scene_frame_rate; }
	const GlobalSettings* getGlobalSettings() const override { return &m_settings; }

	const Object* const* getAllObjects() const override { return m_all_objects.empty() ? nullptr : &m_all_objects[0]; }


	int getAllObjectCount() const override { return (int)m_all_objects.size(); }


	const AnimationStack* getAnimationStack(int index) const override
	{
		assert(index >= 0);
		assert(index < m_animation_stacks.size());
		return m_animation_stacks[index];
	}


	const Mesh* getMesh(int index) const override
	{
		assert(index >= 0);
		assert(index < m_meshes.size());
		return m_meshes[index];
	}


	const TakeInfo* getTakeInfo(const char* name) const override
	{
		for (const TakeInfo& info : m_take_infos)
		{
			if (info.name == name) return &info;
		}
		return nullptr;
	}


	const IElement* getRootElement() const override { return m_root_element; }
	const Object* getRoot() const override { return m_root; }


	void destroy() override { delete this; }


	~Scene()
	{
		for (auto iter : m_object_map)
		{
			delete iter.second.object;
		}
		
		deleteElement(m_root_element);
	}


	Element* m_root_element = nullptr;
	Root* m_root = nullptr;
	float m_scene_frame_rate = -1;
	GlobalSettings m_settings;
	std::unordered_map<uint64, ObjectPair> m_object_map;
	std::vector<Object*> m_all_objects;
	std::vector<Mesh*> m_meshes;
	std::vector<AnimationStack*> m_animation_stacks;
	std::vector<Connection> m_connections;
	std::vector<uint8> m_data;
	std::vector<TakeInfo> m_take_infos;
};

type IScene struct{}

func (is *IScene) getRootElement() *IElement {
	return nil
}
func (is *IScene) getRoot() *Object {
	return nil
}
func (is *IScene) getTakeInfo(name string) *TakeInfo {
	return nil
}
func (is *IScene) getSceneFrameRate() float32 {
	return 0
}
func (is *IScene) getMesh(int index) []Mesh {
	return nil
}
func (is *IScene) getAnimationStack(index int) []AnimationStack {
	return nil
}
func (is *IScene) getAllObjects() []Object {
	return nil
}

func load(data []byte) *Iscene {
	std::unique_ptr<Scene> scene = std::make_unique<Scene>();
	scene.m_data.resize(size);
	memcpy(&scene.m_data[0], data, size);
	OptionalError<Element*> root = tokenize(&scene.m_data[0], size);
	if (root.isError())
	{
		Error::s_message = "";
		root = tokenizeText(&scene.m_data[0], size);
		if (root.isError()) return nullptr;
	}

	scene.m_root_element = root.getValue();
	assert(scene.m_root_element);

	//if (parseTemplates(*root.getValue()).isError()) return nullptr;
	if(!parseConnections(*root.getValue(), scene.get())) return nullptr;
	if(!parseTakes(scene.get())) return nullptr;
	if(!parseObjects(*root.getValue(), scene.get())) return nullptr;
	parseGlobalSettings(*root.getValue(), scene.get());

	return scene.release();
}

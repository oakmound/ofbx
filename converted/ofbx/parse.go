package ofbx


static bool parseObjects(const Element& root, Scene* scene)
{
	const Element* objs = findChild(root, "Objects");
	if (!objs) return true;

	scene.m_root = new Root(*scene, root);
	scene.m_root.id = 0;
	scene.m_object_map[0] = {&root, scene.m_root};

	const Element* object = objs.child;
	while (object)
	{
		if (!isLong(object.first_property))
		{
			Error::s_message = "Invalid";
			return false;
		}

		uint64 id = object.first_property.value.touint64();
		scene.m_object_map[id] = {object, nullptr};
		object = object.sibling;
	}

	for (auto iter : scene.m_object_map)
	{
		OptionalError<Object*> obj = nullptr;

		if (iter.second.object == scene.m_root) continue;

		if (iter.second.element.id == "Geometry")
		{
			Property* last_prop = iter.second.element.first_property;
			while (last_prop.next) last_prop = last_prop.next;
			if (last_prop && last_prop.value == "Mesh")
			{
				obj = parseGeometry(*scene, *iter.second.element);
			}
		}
		else if (iter.second.element.id == "Material")
		{
			obj = parseMaterial(*scene, *iter.second.element);
		}
		else if (iter.second.element.id == "AnimationStack")
		{
			obj = parse<AnimationStackImpl>(*scene, *iter.second.element);
			if (!obj.isError())
			{
				AnimationStackImpl* stack = (AnimationStackImpl*)obj.getValue();
				scene.m_animation_stacks.push_back(stack);
			}
		}
		else if (iter.second.element.id == "AnimationLayer")
		{
			obj = parse<AnimationLayerImpl>(*scene, *iter.second.element);
		}
		else if (iter.second.element.id == "AnimationCurve")
		{
			obj = parseAnimationCurve(*scene, *iter.second.element);
		}
		else if (iter.second.element.id == "AnimationCurveNode")
		{
			obj = parse<AnimationCurveNodeImpl>(*scene, *iter.second.element);
		}
		else if (iter.second.element.id == "Deformer")
		{
			IElementProperty* class_prop = iter.second.element.getProperty(2);

			if (class_prop)
			{
				if (class_prop.getValue() == "Cluster")
					obj = parseCluster(*scene, *iter.second.element);
				else if (class_prop.getValue() == "Skin")
					obj = parse<SkinImpl>(*scene, *iter.second.element);
			}
		}
		else if (iter.second.element.id == "NodeAttribute")
		{
			obj = parseNodeAttribute(*scene, *iter.second.element);
		}
		else if (iter.second.element.id == "Model")
		{
			IElementProperty* class_prop = iter.second.element.getProperty(2);

			if (class_prop)
			{
				if (class_prop.getValue() == "Mesh")
				{
					obj = parseMesh(*scene, *iter.second.element);
					if (!obj.isError())
					{
						Mesh* mesh = (Mesh*)obj.getValue();
						scene.m_meshes.push_back(mesh);
						obj = mesh;
					}
				}
				else if (class_prop.getValue() == "LimbNode")
					obj = parseLimbNode(*scene, *iter.second.element);
				else if (class_prop.getValue() == "Null")
					obj = parse<NullImpl>(*scene, *iter.second.element);
				else if (class_prop.getValue() == "Root")
					obj = parse<NullImpl>(*scene, *iter.second.element);
			}
		}
		else if (iter.second.element.id == "Texture")
		{
			obj = parseTexture(*scene, *iter.second.element);
		}

		if (obj.isError()) return false;

		scene.m_object_map[iter.first].object = obj.getValue();
		if (obj.getValue())
		{
			scene.m_all_objects.push_back(obj.getValue());
			obj.getValue().id = iter.first;
		}
	}

	for (const Scene::Connection& con : scene.m_connections)
	{
		Object* parent = scene.m_object_map[con.to].object;
		Object* child = scene.m_object_map[con.from].object;
		if (!child) continue;
		if (!parent) continue;

		switch (child.getType())
		{
			case Object::Type::NODE_ATTRIBUTE:
				if (parent.node_attribute)
				{
					Error::s_message = "Invalid node attribute";
					return false;
				}
				parent.node_attribute = (NodeAttribute*)child;
				break;
			case Object::Type::ANIMATION_CURVE_NODE:
				if (parent.isNode())
				{
					AnimationCurveNodeImpl* node = (AnimationCurveNodeImpl*)child;
					node.bone = parent;
					node.bone_link_property = con.property;
				}
				break;
		}

		switch (parent.getType())
		{
			case Object::Type::MESH:
			{
				MeshImpl* mesh = (MeshImpl*)parent;
				switch (child.getType())
				{
					case Object::Type::GEOMETRY:
						if (mesh.geometry)
						{
							Error::s_message = "Invalid mesh";
							return false;
						}
						mesh.geometry = (Geometry*)child;
						break;
					case Object::Type::MATERIAL: mesh.materials.push_back((Material*)child); break;
				}
				break;
			}
			case Object::Type::SKIN:
			{
				SkinImpl* skin = (SkinImpl*)parent;
				if (child.getType() == Object::Type::CLUSTER)
				{
					ClusterImpl* cluster = (ClusterImpl*)child;
					skin.clusters.push_back(cluster);
					if (cluster.skin)
					{
						Error::s_message = "Invalid cluster";
						return false;
					}
					cluster.skin = skin;
				}
				break;
			}
			case Object::Type::MATERIAL:
			{
				MaterialImpl* mat = (MaterialImpl*)parent;
				if (child.getType() == Object::Type::TEXTURE)
				{
					Texture::TextureType type = Texture::COUNT;
					if (con.property == "NormalMap")
						type = Texture::NORMAL;
					else if (con.property == "DiffuseColor")
						type = Texture::DIFFUSE;
					if (type == Texture::COUNT) break;

					if (mat.textures[type])
					{
						break;// This may happen for some models (eg. 2 normal maps in use)
						Error::s_message = "Invalid material";
						return false;
					}

					mat.textures[type] = (Texture*)child;
				}
				break;
			}
			case Object::Type::GEOMETRY:
			{
				GeometryImpl* geom = (GeometryImpl*)parent;
				if (child.getType() == Object::Type::SKIN) geom.skin = (Skin*)child;
				break;
			}
			case Object::Type::CLUSTER:
			{
				ClusterImpl* cluster = (ClusterImpl*)parent;
				if (child.getType() == Object::Type::LIMB_NODE || child.getType() == Object::Type::MESH || child.getType() == Object::Type::NULL_NODE)
				{
					if (cluster.link)
					{
						Error::s_message = "Invalid cluster";
						return false;
					}

					cluster.link = child;
				}
				break;
			}
			case Object::Type::ANIMATION_LAYER:
			{
				if (child.getType() == Object::Type::ANIMATION_CURVE_NODE)
				{
					((AnimationLayerImpl*)parent).curve_nodes.push_back((AnimationCurveNodeImpl*)child);
				}
			}
			break;
			case Object::Type::ANIMATION_CURVE_NODE:
			{
				AnimationCurveNodeImpl* node = (AnimationCurveNodeImpl*)parent;
				if (child.getType() == Object::Type::ANIMATION_CURVE)
				{
					if (!node.curves[0].curve)
					{
						node.curves[0].connection = &con;
						node.curves[0].curve = (AnimationCurve*)child;
					}
					else if (!node.curves[1].curve)
					{
						node.curves[1].connection = &con;
						node.curves[1].curve = (AnimationCurve*)child;
					}
					else if (!node.curves[2].curve)
					{
						node.curves[2].connection = &con;
						node.curves[2].curve = (AnimationCurve*)child;
					}
					else
					{
						Error::s_message = "Invalid animation node";
						return false;
					}
				}
				break;
			}
		}
	}

	for (auto iter : scene.m_object_map)
	{
		Object* obj = iter.second.object;
		if (!obj) continue;
		if(obj.getType() == Object::Type::CLUSTER)
		{
			if (!((ClusterImpl*)iter.second.object).postprocess())
			{
				Error::s_message = "Failed to postprocess cluster";
				return false;
			}
		}
	}

	return true;
}

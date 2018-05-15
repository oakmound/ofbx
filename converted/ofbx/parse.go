package ofbx

func parseTexture(scene *Scene, element *Element) *Object {
	texture := NewTexture(scene, element)
	texture_filename := findChild(element, "FileName")
	if texture_filename && texture_filename.first_property {
		texture.filename = texture_filename.first_property.value
	}
	texture_relative_filename := findChild(element, "RelativeFilename")
	if texture_relative_filename && texture_relative_filename.first_property {
		texture.relative_filename = texture_relative_filename.first_property.value
	}
	return texture
}

func parseLimbNode(scene *Scene, element *Element) (*Object, error) {
	if (element.first_property == nil
		|| !element.first_property.next == nil
		|| !element.first_property.next.next == nil
		|| element.first_property.next.next.value != "LimbNode") {
		return nil, errors.New("Invalid limb node")
	}
	return NewLimpNode(scene, element)
}

func parseMesh(scene *Scene, element *Element) (*Object, error) {
	if (element.first_property == nil
		|| element.first_property.next == nil 
		|| element.first_property.next.next == nil
		|| element.first_property.next.next.value != "Mesh") {
		return nil, errors.New("Invalid mesh")
	}
	return NewMesh(scene, element)
}

func parseMaterial(scene *Scene, element *Element) *Object {
	material := NewMaterial(scene, element)
	prop := findChild(element, "Properties70")
	material.diffuse_color = Color{1,1,1}
	if prop != nil {
		prop = prop.child
	}
	for prop != nil {
		if prop.id == "P" && prop.first_property {
			if prop.first_property.value == "DiffuseColor" {
				material.diffuse_color.r = float32(prop.getProperty(4).getValue().toDouble())
				material.diffuse_color.g = float32(prop.getProperty(5).getValue().toDouble())
				material.diffuse_color.b = float32(prop.getProperty(6).getValue().toDouble())
			}
		}
		prop = prop.sibling
	}
	return material
}

static OptionalError<Object*> parseAnimationCurve(const Scene& scene, const Element& element) {
	std::unique_ptr<AnimationCurveImpl> curve = std::make_unique<AnimationCurveImpl>(scene, element);

	const Element* times = findChild(element, "KeyTime");
	const Element* values = findChild(element, "KeyValueFloat");

	if (times && times.first_property) {
		curve.times.resize(times.first_property.getCount());
		if (!times.first_property.getValues(&curve.times[0], (int)curve.times.size() * sizeof(curve.times[0]))) {
			return Error("Invalid animation curve");
		}
	}

	if (values && values.first_property) {
		curve.values.resize(values.first_property.getCount());
		if (!values.first_property.getValues(&curve.values[0], (int)curve.values.size() * sizeof(curve.values[0]))) {
			return Error("Invalid animation curve");
		}
	}

	if (curve.times.size() != curve.values.size()) return Error("Invalid animation curve");

	return curve.release();
}

static bool parseConnections(const Element& root, Scene* scene) {
	assert(scene);

	const Element* connections = findChild(root, "Connections");
	if (!connections) return true;

	const Element* connection = connections.child;
	while (connection) {
		if (!isString(connection.first_property)
			|| !isLong(connection.first_property.next)
			|| !isLong(connection.first_property.next.next)) {
			Error::s_message = "Invalid connection";
			return false;
		}

		Scene::Connection c;
		c.from = connection.first_property.next.value.touint64();
		c.to = connection.first_property.next.next.value.touint64();
		if (connection.first_property.value == "OO") {
			c.type = Scene::Connection::OBJECT_OBJECT;
		}
		else if (connection.first_property.value == "OP") {
			c.type = Scene::Connection::OBJECT_PROPERTY;
			if (!connection.first_property.next.next.next) {
				Error::s_message = "Invalid connection";
				return false;
			}
			c.property = connection.first_property.next.next.next.value;
		}
		else {
			assert(false);
			Error::s_message = "Not supported";
			return false;
		}
		scene.m_connections.push_back(c);

		connection = connection.sibling;
	}
	return true;
}

static bool parseTakes(Scene* scene) {
	const Element* takes = findChild((const Element&)*scene.getRootElement(), "Takes");
	if (!takes) return true;

	const Element* object = takes.child;
	while (object) {
		if (object.id == "Take") {
			if (!isString(object.first_property)) {
				Error::s_message = "Invalid name in take";
				return false;
			}

			TakeInfo take;
			take.name = object.first_property.value;
			const Element* filename = findChild(*object, "FileName");
			if (filename) {
				if (!isString(filename.first_property)) {
					Error::s_message = "Invalid filename in take";
					return false;
				}
				take.filename = filename.first_property.value;
			}
			const Element* local_time = findChild(*object, "LocalTime");
			if (local_time) {
				if (!isLong(local_time.first_property) || !isLong(local_time.first_property.next)) {
					Error::s_message = "Invalid local time in take";
					return false;
				}

				take.local_time_from = fbxTimeToSeconds(local_time.first_property.value.toint64());
				take.local_time_to = fbxTimeToSeconds(local_time.first_property.next.value.toint64());
			}
			const Element* reference_time = findChild(*object, "ReferenceTime");
			if (reference_time) {
				if (!isLong(reference_time.first_property) || !isLong(reference_time.first_property.next)) {
					Error::s_message = "Invalid reference time in take";
					return false;
				}

				take.reference_time_from = fbxTimeToSeconds(reference_time.first_property.value.toint64());
				take.reference_time_to = fbxTimeToSeconds(reference_time.first_property.next.value.toint64());
			}

			scene.m_take_infos.push_back(take);
		}

		object = object.sibling;
	}

	return true;
}

static void parseGlobalSettings(const Element& root, Scene* scene) {
	for (ofbx::Element* settings = root.child; settings; settings = settings.sibling) {
		if (settings.id == "GlobalSettings") {
			for (ofbx::Element* props70 = settings.child; props70; props70 = props70.sibling) {
				if (props70.id == "Properties70") {
					for (ofbx::Element* node = props70.child; node; node = node.sibling) {
						if (!node.first_property)
							continue;

#define get_property(name, field, type) if(node.first_property.value == name) \ { \
							ofbx::IElementProperty* prop = node.getProperty(4); \
							if (prop) \ { \
								ofbx::DataView value = prop.getValue(); \
								scene.m_settings.field = *(type*)value.begin; \
							} \
						}

						get_property("UpAxis", UpAxis, UpVector);
						get_property("UpAxisSign", UpAxisSign, int);
						get_property("FrontAxis", FrontAxis, FrontVector);
						get_property("FrontAxisSign", FrontAxisSign, int);
						get_property("CoordAxis", CoordAxis, CoordSystem);
						get_property("CoordAxisSign", CoordAxisSign, int);
						get_property("OriginalUpAxis", OriginalUpAxis, int);
						get_property("OriginalUpAxisSign", OriginalUpAxisSign, int);
						get_property("UnitScaleFactor", UnitScaleFactor, float);
						get_property("OriginalUnitScaleFactor", OriginalUnitScaleFactor, float);
						get_property("TimeSpanStart", TimeSpanStart, uint64);
						get_property("TimeSpanStop", TimeSpanStop, uint64);
						get_property("TimeMode", TimeMode, FrameRate);
						get_property("CustomFrameRate", CustomFrameRate, float);

#undef get_property

						scene.m_scene_frame_rate = getFramerateFromTimeMode(scene.m_settings.TimeMode, scene.m_settings.CustomFrameRate);
					}
					break;
				}
			}
			break;
		}
	}
}

func parseObjects(root *Element, scene *Scene) (bool, error) {
	objs := findChild(root, "Objects")
	if objs == nil {
		return true, nil
	}

	scene.m_root = NewRoot(*scene, root)
	scene.m_object_map[0] = ObjectPair{&root, scene.m_root}

	object := objs.child
	for object != nil {
		if !isLong(object.first_property) {
			return false, errors.New("Invalid")
		}

		id := object.first_property.value.touint64()
		scene.m_object_map[id] = ObjectPair{object, nullptr}
		object = object.sibling
	}

	for _, iter := range scene.m_object_map {
		var obj *Object

		if iter.second.object == scene.m_root {
			continue
		}

		if iter.second.element.id == "Geometry" {
			last_prop := iter.second.element.first_property
			for last_prop.next != nil {
				last_prop = last_prop.next
			} 
			if last_prop != nil && last_prop.value == "Mesh" {
				obj = parseGeometry(*scene, *iter.second.element)
			}
		} else if iter.second.element.id == "Material" {
			obj = parseMaterial(*scene, *iter.second.element)
		} else if iter.second.element.id == "AnimationStack" {
			obj = NewAnimationStack(*scene, *iter.second.element)
			if !obj.isError() {
				stack := obj.getValue().(*AnimationStackImpl)
				scene.m_animation_stacks.push_back(stack)
			}
		} else if iter.second.element.id == "AnimationLayer" {
			obj = NewAnimationLayer(*scene, *iter.second.element)
		} else if iter.second.element.id == "AnimationCurve" {
			obj = parseAnimationCurve(*scene, *iter.second.element)
		} else if iter.second.element.id == "AnimationCurveNode" {
			obj = NewAnimationCurveNode(*scene, *iter.second.element)
		} else if iter.second.element.id == "Deformer" {
			class_prop = iter.second.element.getProperty(2)
			if class_prop != nil {
				v := class_prop.getValue()
				if v == "Cluster" {
					obj = parseCluster(*scene, *iter.second.element)
				} else if v == "Skin" {
					obj = NewSkin(*scene, *iter.second.element)
				}
			}
		} else if iter.second.element.id == "NodeAttribute" {
			obj = parseNodeAttribute(*scene, *iter.second.element)
		} else if iter.second.element.id == "Model" {
			iter.second.element.getProperty(2)
			if class_prop != nil {
				v := class_prop.getValue()
				if v == "Mesh" {
					obj = parseMesh(*scene, *iter.second.element)
					if !obj.isError() {
						mesh = obj.getValue().(*Mesh)
						scene.m_meshes.push_back(mesh)
						obj = mesh
					}
				} else if v == "LimbNode" {
					obj, err = parseLimbNode(*scene, *iter.second.element)
					if err != nil {
						return false, err
					}
				} else if v == "Null" || v == "Root" {
					obj = NewNull(*scene, *iter.second.element)
				} 
			}
		} else if (iter.second.element.id == "Texture") {
			obj = parseTexture(*scene, *iter.second.element)
		}


		if obj.isError() {
			return false, nil // error?
		}

		val := obj.getValue()
		scene.m_object_map[iter.first].object = val 
		if val != nil {
			scene.m_all_objects.push_back(val)
			val.id = iter.first
		}
	}
	for _, con := range scene.m_connections {
		parent := scene.m_object_map[con.to].object
		child := scene.m_object_map[con.from].object
		if child == nil || parent == nil {
			continue
		}

		ctyp := child.getType()

		switch ctyp {
			case NODE_ATTRIBUTE:
				if parent.node_attribute {
					return false, errors.New("Invalid node attribute")
				}
				parent.node_attribute = child.(*NodeAttribute)
			case ANIMATION_CURVE_NODE:
				if parent.isNode() {
					node := child.(*AnimationCurveNode)
					node.bone = parent
					node.bone_link_property = con.property
				}
		}

		switch (parent.getType()) {
			case MESH: {
				mesh := parent.(*MeshImpl)
				switch ctyp {
					case GEOMETRY:
						if mesh.geometry != nil {
							return false, errors.New("Invalid mesh")
						}
						mesh.geometry = child.(*Geometry)
					case MATERIAL: 
						mesh.materials.push_back(child.(*Material))
				}
			}
			case SKIN: {
				skin := parent.(*Skin)
				if ctyp == CLUSTER {
					cluster := child.(*Cluster)
					skin.clusters.push_back(cluster)
					if cluster.skin != nil {
						return false, errors.New("Invalid cluster")
					}
					cluster.skin = skin
				}
			}
			case MATERIAL: {
				mat := parent.(*Material)
				if ctyp == TEXTURE {
					ttyp = COUNT
					if con.property == "NormalMap" {
						ttyp = NORMAL
					} else if con.property == "DiffuseColor"
						ttyp = DIFFUSE
					if ttyp == COUNT {
						break
					}
					if mat.textures[ttyp] != nil {
						break
					}
					mat.textures[ttyp] = child.(*Texture)
				}
			}
			case GEOMETRY:
				geom := parent.(*Geometry)
				if ctyp == SKIN { 
					geom.skin = child.(*Skin)
				}
			}
			case CLUSTER:
				cluster := parent.(*Cluster)
				if ctyp == LIMB_NODE || ctyp == MESH || ctyp == NULL_NODE {
					if cluster.link != nil {
						return false, errors.New("Invalid cluster")
					}
					cluster.link = child
				}
			}
			case ANIMATION_LAYER:
				if ctyp == ANIMATION_CURVE_NODE {
					(parent.(*AnimationLayer).curve_nodes.push_back(child.(*AnimationCurveNode))
				}
			}
			case ANIMATION_CURVE_NODE:
				node = parent.(*AnimationCurveNode)
				if ctyp == ANIMATION_CURVE {
					if !node.curves[0].curve == nil {
						node.curves[0].connection = &con
						node.curves[0].curve = child.(*AnimationCurve)
					} else if node.curves[1].curve == nil {
						node.curves[1].connection = &con
						node.curves[1].curve = child.(*AnimationCurve)
					} else if !node.curves[2].curve == nil {
						node.curves[2].connection = &con
						node.curves[2].curve = child.(*AnimationCurve)
					} else {
						return false, errors.New("Invalid animation node")
					}
				}
			}
		}
	}

	for _, iter := range scene.m_object_map {
		obj := iter.second.object
		if obj == nil {
			continue
		}
		if obj.getType() == CLUSTER {
			if !iter.second.object.(*ClusterImpl).postprocess() {
				return false, errors.New("Failed to postprocess cluster")
			}
		}
	}

	return true
}

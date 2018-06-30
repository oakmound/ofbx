package ofbx

import (
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
)

func parseTemplates(root *Element) {
	defs := findChildren(root, "Definitions")
	if defs == nil {
		return
	}

	templates := make(map[string]*Element)
	defs = defs[0].children
	for _, def := range defs {
		if def.id.String() == "ObjectType" {
			prop1 := def.first_property.value
			prop1Data, err := ioutil.ReadAll(prop1)
			if err != nil && err != io.EOF {
				fmt.Println(err)
				continue
			}
			subdefs := def.children
			for _, subdef := range subdefs {
				if subdef.id.String() == "PropertyTemplate" {
					prop2 := subdef.first_property.value
					prop2Data, err := ioutil.ReadAll(prop2)
					if err != nil && err != io.EOF {
						fmt.Println(err)
						continue
					}
					templates[string(prop1Data)+string(prop2Data)] = subdef
				}
			}

		}
	}
}

func parseBinaryArrayInt(property *Property) ([]int, error) {
	count := property.getCount()
	if count == 0 {
		return []int{}, nil
	}
	if !property.typ.IsArray() {
		return nil, errors.New("Invalid type")
	}
	return parseArrayRawInt(property)
}

func parseBinaryArrayFloat64(property *Property) ([]float64, error) {
	count := property.getCount()
	if count == 0 {
		return []float64{}, nil
	}
	if !property.typ.IsArray() {
		return nil, errors.New("Invalid type")
	}
	return parseArrayRawFloat64(property)
}

// This might not work??
func parseBinaryArrayFloat32(property *Property) ([]float32, error) {
	f64s, err := parseBinaryArrayFloat64(property)
	if err != nil {
		return nil, err
	}
	f32s := make([]float32, len(f64s))
	for i, f64 := range f64s {
		f32s[i] = float32(f64)
	}
	return f32s, nil
}
func parseBinaryArrayVec2(property *Property) ([]Vec2, error) {
	f64s, err := parseBinaryArrayFloat64(property)
	if err != nil {
		return nil, err
	}
	vs := make([]Vec2, len(f64s)/2)
	for i := 0; i < len(f64s); i += 2 {
		vs[i/2].X = f64s[i]
		vs[i/2].Y = f64s[i+1]
	}
	return vs, nil
}
func parseBinaryArrayVec3(property *Property) ([]Vec3, error) {
	f64s, err := parseBinaryArrayFloat64(property)
	if err != nil {
		return nil, err
	}
	vs := make([]Vec3, len(f64s)/3)
	// len(f64s) should probably be divisible by 3
	if len(f64s)%3 != 0 {
		fmt.Println("Vec3 binary array not made up of Vec3s")
	}
	for i := 0; (i + 2) < len(f64s); i += 3 {
		vs[i/3].X = f64s[i]
		vs[i/3].Y = f64s[i+1]
		vs[i/3].Z = f64s[i+2]
	}
	return vs, nil
}
func parseBinaryArrayVec4(property *Property) ([]Vec4, error) {
	f64s, err := parseBinaryArrayFloat64(property)
	if err != nil {
		return nil, err
	}
	vs := make([]Vec4, len(f64s)/4)
	for i := 0; i < len(f64s); i += 4 {
		vs[i/4].X = f64s[i]
		vs[i/4].Y = f64s[i+1]
		vs[i/4].Z = f64s[i+2]
		vs[i/4].w = f64s[i+3]
	}
	return vs, nil
}

func parseArrayRawInt(property *Property) ([]int, error) {
	if property.typ == 'd' || property.typ == 'f' {
		return nil, errors.New("Invalid type, expected i or l")
	}
	if property.encoding == 0 {
		return parseArrayRawIntEnd(property.value, property.count, property.typ.Size()), nil
	} else if property.encoding == 1 {
		zr, err := zlib.NewReader(&property.value.Reader)
		if err != nil {
			return nil, errors.Wrap(err, "New Reader failed")
		}
		defer zr.Close()
		return parseArrayRawIntEnd(zr, property.count, property.typ.Size()), nil
	}
	return nil, errors.New("Invalid encoding")
}

func parseArrayRawIntEnd(r io.Reader, ln int, elem_size int) []int {
	if elem_size == 4 {
		i32s := make([]int32, int(ln))
		binary.Read(r, binary.LittleEndian, i32s)
		out := make([]int, len(i32s))
		for i, f := range i32s {
			out[i] = int(f)
		}
		return out
	}
	i64s := make([]int64, int(ln))
	binary.Read(r, binary.LittleEndian, i64s)
	out := make([]int, len(i64s))
	for i, f := range i64s {
		out[i] = int(f)
	}
	return out
}

func parseArrayRawInt64(property *Property) ([]int64, error) {
	if property.typ == 'd' || property.typ == 'f' {
		return nil, errors.New("Invalid type, expected i or l")
	}
	if property.encoding == 0 {
		return parseArrayRawInt64End(property.value, property.count, property.typ.Size()), nil
	} else if property.encoding == 1 {
		zr, err := zlib.NewReader(&property.value.Reader)
		if err != nil {
			return nil, errors.Wrap(err, "New Reader failed")
		}
		defer zr.Close()
		return parseArrayRawInt64End(zr, property.count, property.typ.Size()), nil
	}
	return nil, errors.New("Invalid encoding")
}

func parseArrayRawInt64End(r io.Reader, ln int, elem_size int) []int64 {
	if elem_size == 4 {
		i32s := make([]int32, int(ln))
		binary.Read(r, binary.LittleEndian, i32s)
		out := make([]int64, len(i32s))
		for i, f := range i32s {
			out[i] = int64(f)
		}
		return out
	}
	out := make([]int64, int(ln))
	binary.Read(r, binary.LittleEndian, out)
	return out
}

func parseArrayRawFloat32(property *Property) ([]float32, error) {
	if property.typ == 'i' || property.typ == 'l' {
		return nil, errors.New("Invalid type, expected d or f")
	}
	if property.encoding == 0 {
		return parseArrayRawFloat32End(property.value, property.count, property.typ.Size()), nil
	} else if property.encoding == 1 {
		zr, err := zlib.NewReader(&property.value.Reader)
		if err != nil {
			return nil, errors.Wrap(err, "New Reader failed")
		}
		defer zr.Close()
		return parseArrayRawFloat32End(zr, property.count, property.typ.Size()), nil
	}
	return nil, errors.New("Invalid encoding")
}

func parseArrayRawFloat32End(r io.Reader, ln int, elem_size int) []float32 {
	if elem_size == 4 {
		out := make([]float32, int(ln))
		binary.Read(r, binary.LittleEndian, out)
		return out
	}
	f64s := make([]float64, int(ln))
	binary.Read(r, binary.LittleEndian, f64s)
	out := make([]float32, len(f64s))
	for i, f := range f64s {
		out[i] = float32(f)
	}
	return out
}

func parseArrayRawFloat64(property *Property) ([]float64, error) {
	if property.typ == 'i' || property.typ == 'l' {
		return nil, errors.New("Invalid type, expected d or f")
	}
	if property.encoding == 0 {
		return parseArrayRawFloat64End(property.value, property.count, property.typ.Size()), nil
	} else if property.encoding == 1 {
		zr, err := zlib.NewReader(&property.value.Reader)
		if err != nil {
			return nil, errors.Wrap(err, "New Reader failed")
		}
		defer zr.Close()
		return parseArrayRawFloat64End(zr, property.count, property.typ.Size()), nil
	}
	return nil, errors.New("Invalid encoding")
}

func parseArrayRawFloat64End(r io.Reader, ln int, elem_size int) []float64 {
	if elem_size == 4 {
		f32s := make([]float32, int(ln))
		binary.Read(r, binary.LittleEndian, f32s)
		out := make([]float64, len(f32s))
		for i, f := range f32s {
			out[i] = float64(f)
		}
		return out
	}
	out := make([]float64, int(ln))
	binary.Read(r, binary.LittleEndian, out)
	return out
}

func parseDoubleVecDataVec2(property *Property) ([]Vec2, error) {
	if property.typ == 'd' {
		return parseBinaryArrayVec2(property)
	}
	tmp, err := parseBinaryArrayFloat32(property)
	if err != nil {
		return nil, err
	}
	size := 2
	out_vec := make([]Vec2, len(tmp)/size)
	for i := 0; i < len(tmp); i += size {
		j := i / size
		out_vec[j].X = float64(tmp[i])
		out_vec[j].Y = float64(tmp[i+1])
	}
	return out_vec, nil
}

func parseDoubleVecDataVec3(property *Property) ([]Vec3, error) {
	if property.typ == 'd' {
		return parseBinaryArrayVec3(property)
	}
	tmp, err := parseBinaryArrayFloat32(property)
	if err != nil {
		return nil, err
	}
	size := 3
	out_vec := make([]Vec3, len(tmp)/size)
	for i := 0; i < len(tmp); i += size {
		j := i / size
		out_vec[j].X = float64(tmp[i])
		out_vec[j].Y = float64(tmp[i+1])
		out_vec[j].Z = float64(tmp[i+2])
	}
	return out_vec, nil
}

func parseDoubleVecDataVec4(property *Property) ([]Vec4, error) {
	if property.typ == 'd' {
		return parseBinaryArrayVec4(property)
	}
	tmp, err := parseBinaryArrayFloat32(property)
	if err != nil {
		return nil, err
	}
	size := 4
	out_vec := make([]Vec4, len(tmp)/size)
	for i := 0; i < len(tmp); i += size {
		j := i / size
		out_vec[j].X = float64(tmp[i])
		out_vec[j].Y = float64(tmp[i+1])
		out_vec[j].Z = float64(tmp[i+2])
		out_vec[j].w = float64(tmp[i+3])
	}
	return out_vec, nil
}

func parseVertexDataVec2(element *Element, name, index_name string) ([]Vec2, []int, VertexDataMapping, error) {
	idxs, mapping, dataProp, err := parseVertexDataInner(element, name, index_name)
	vcs, err := parseDoubleVecDataVec2(dataProp)
	return vcs, idxs, mapping, err
}

func parseVertexDataVec3(element *Element, name, index_name string) ([]Vec3, []int, VertexDataMapping, error) {
	idxs, mapping, dataProp, err := parseVertexDataInner(element, name, index_name)
	vcs, err := parseDoubleVecDataVec3(dataProp)
	return vcs, idxs, mapping, err
}

func parseVertexDataVec4(element *Element, name, index_name string) ([]Vec4, []int, VertexDataMapping, error) {

	idxs, mapping, dataProp, err := parseVertexDataInner(element, name, index_name)
	vcs, err := parseDoubleVecDataVec4(dataProp)
	return vcs, idxs, mapping, err
}

func parseVertexDataInner(element *Element, name, index_name string) ([]int, VertexDataMapping, *Property, error) {
	dataProp := findChildProperty(element, name)
	if dataProp == nil {
		return nil, 0, nil, errors.New("Invalid data element")
	}
	mappingProp := findChildProperty(element, "MappingInformationType")
	referenceProp := findChildProperty(element, "ReferenceInformationType")

	var idxs []int
	var mapping VertexDataMapping
	var err error

	if mappingProp != nil {
		s := mappingProp.value.String()
		if s == "ByPolygonVertex" {
			mapping = BY_POLYGON_VERTEX
		} else if s == "ByPolygon" {
			mapping = BY_POLYGON
		} else if s == "ByVertice" || s == "ByVertex" {
			mapping = BY_VERTEX
		} else {
			return nil, 0, nil, errors.New("Unable to parse mapping")
		}
	}
	if referenceProp != nil {
		if referenceProp.value.String() == "IndexToDirect" {
			indicesProp := findChildProperty(element, index_name)
			if indicesProp != nil {
				if idxs, err = parseBinaryArrayInt(indicesProp); err != nil {
					return nil, 0, nil, errors.New("Unable to parse indices")
				}
			}
		} else if referenceProp.value.String() != "Direct" {
			return nil, 0, nil, errors.New("Invalid properties")
		}
	}
	return idxs, mapping, dataProp, nil
}

func parseTexture(scene *Scene, element *Element) *Texture {
	texture := NewTexture(scene, element)
	textureFilenameProp := findChildProperty(element, "FileName")
	if textureFilenameProp != nil {
		texture.filename = textureFilenameProp.value
	}
	textureRelativeFilenameProp := findChildProperty(element, "RelativeFilename")
	if textureRelativeFilenameProp != nil {
		texture.relative_filename = textureRelativeFilenameProp.value
	}
	return texture
}

func parseLimbNode(scene *Scene, element *Element) (*Node, error) {
	if element.first_property == nil ||
		element.first_property.next == nil ||
		element.first_property.next.next == nil ||
		element.first_property.next.next.value.String() != "LimbNode" {
		return nil, errors.New("Invalid limb node")
	}
	return NewNode(scene, element, LIMB_NODE), nil
}

func parseMesh(scene *Scene, element *Element) (*Mesh, error) {
	if element.first_property == nil ||
		element.first_property.next == nil ||
		element.first_property.next.next == nil ||
		element.first_property.next.next.value.String() != "Mesh" {
		return nil, errors.New("Invalid mesh")
	}
	return NewMesh(scene, element), nil
}

func parseMaterial(scene *Scene, element *Element) *Material {
	material := NewMaterial(scene, element)
	props := findChildren(element, "Properties70")
	material.diffuse_color = Color{1, 1, 1}
	if len(props) == 0 {
		return material
	}
	props = props[0].children
	//For some reason materials inherit the last diffuse color in the property list?
	for i := len(props) - 1; i >= 0; i-- {
		prop := props[i]
		if prop.id.String() == "p" && prop.first_property != nil && prop.first_property.value.String() == "DiffuseColor" {
			material.diffuse_color.r = float32(prop.getProperty(4).getValue().toDouble())
			material.diffuse_color.g = float32(prop.getProperty(5).getValue().toDouble())
			material.diffuse_color.b = float32(prop.getProperty(6).getValue().toDouble())
			return material
		}
	}
	return material
}

func parseAnimationCurve(scene *Scene, element *Element) (*AnimationCurve, error) {
	curve := &AnimationCurve{}

	times := findChildProperty(element, "KeyTime")
	values := findChildProperty(element, "KeyValueFloat")

	if times != nil {
		var err error
		curve.times, err = times.getValuesInt64()
		if err != nil {
			return nil, errors.Wrap(err, "Invalid animation curve: times error")
		}
	}
	if values != nil {
		var err error
		curve.values, err = values.getValuesF32()
		if err != nil {
			return nil, errors.New("Invalid animation curve: values error")
		}
	}
	if len(curve.times) != len(curve.values) {
		return nil, errors.New("Invalid animation curve: len error")
	}
	return curve, nil
}

func parseConnection(root *Element, scene *Scene) (bool, error) {
	connections := findChildren(root, "Connections")
	if connections == nil {
		return true, nil
	}

	connections = connections[0].children
	for _, connection := range connections {
		if !isString(connection.first_property) ||
			!isLong(connection.first_property.next) ||
			!isLong(connection.first_property.next.next) {
			return false, errors.New("Invalid connection")
		}
		var c Connection
		c.from = connection.first_property.next.value.touint64()
		c.to = connection.first_property.next.next.value.touint64()
		if connection.first_property.value.String() == "OO" {
			c.typ = OBJECT_OBJECT
		} else if connection.first_property.value.String() == "OP" {
			c.typ = OBJECT_PROPERTY
			if connection.first_property.next.next.next == nil {
				return false, errors.New("Invalid connection")
			}
			c.property = connection.first_property.next.next.next.value.String()
		} else {
			return false, errors.New("Not supported")
		}
		scene.m_connections = append(scene.m_connections, c)
	}
	return true, nil
}

func parseTakes(scene *Scene) (bool, error) {
	takes := findChildren(scene.getRootElement(), "Takes")
	if takes == nil {
		return true, nil
	}

	objects := takes[0].children

	for _, object := range objects {
		if object.id.String() != "Take" {
			continue
		}
		if !isString(object.first_property) {
			return false, errors.New("Invalid name in take")
		}
		var take TakeInfo
		take.name = object.first_property.value
		filename := findChildProperty(object, "FileName")
		if filename != nil {
			if !isString(filename) {
				return false, errors.New("Invalid filename in take")
			}
			take.filename = filename.value
		}
		local_time := findChildProperty(object, "LocalTime")
		if local_time != nil {
			if !isLong(local_time) || !isLong(local_time.next) {
				return false, errors.New("Invalid local time in take")
			}

			take.local_time_from = fbxTimeToSeconds(local_time.value.toint64())
			take.local_time_to = fbxTimeToSeconds(local_time.next.value.toint64())
		}
		reference_time := findChildProperty(object, "ReferenceTime")
		if reference_time != nil {
			if !isLong(reference_time) || !isLong(reference_time.next) {
				return false, errors.New("Invalid reference time in take")
			}
			take.reference_time_from = fbxTimeToSeconds(reference_time.value.toint64())
			take.reference_time_to = fbxTimeToSeconds(reference_time.next.value.toint64())
		}
		scene.m_take_infos = append(scene.m_take_infos, take)
	}
	return true, nil
}

func parseGlobalSettings(root *Element, scene *Scene) {

	for _, settings := range root.children {
		if settings.id.String() != "GlobalSettings" {
			continue
		}
		for _, props70 := range settings.children {
			if props70.id.String() != "Properties70" {
				continue
			}
			for _, node := range props70.children {
				if node.first_property != nil {
					continue
				}
				if node.first_property.value.String() == "UpAxis" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.UpAxis = UpVector(int(value.toInt32()))
					}
				}

				if node.first_property.value.String() == "UpAxisSign" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.UpAxisSign = int(value.toInt32())
					}
				}

				if node.first_property.value.String() == "FrontAxis" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.FrontAxis = FrontVector(int(value.toInt32()))
					}
				}

				if node.first_property.value.String() == "FrontAxisSign" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.FrontAxisSign = int(value.toInt32())
					}
				}

				if node.first_property.value.String() == "CoordAxis" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.CoordAxis = CoordSystem(int(value.toInt32()))
					}
				}

				if node.first_property.value.String() == "CoordAxisSign" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.CoordAxisSign = int(value.toInt32())
					}
				}

				if node.first_property.value.String() == "OriginalUpAxis" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.OriginalUpAxis = int(value.toInt32())
					}
				}

				if node.first_property.value.String() == "OriginalUpAxisSign" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.OriginalUpAxisSign = int(value.toInt32())
					}
				}

				if node.first_property.value.String() == "UnitScaleFactor" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.UnitScaleFactor = value.toFloat()
					}
				}

				if node.first_property.value.String() == "OriginalUnitScaleFactor" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.OriginalUnitScaleFactor = value.toFloat()
					}
				}

				if node.first_property.value.String() == "TimeSpanStart" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.TimeSpanStart = value.touint64()
					}
				}

				if node.first_property.value.String() == "TimeSpanStop" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.TimeSpanStop = value.touint64()
					}
				}

				if node.first_property.value.String() == "TimeMode" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.TimeMode = FrameRate(int(value.toInt32()))
					}
				}

				if node.first_property.value.String() == "CustomFrameRate" {
					prop := node.getProperty(4)
					if prop != nil {
						value := prop.getValue()
						scene.m_settings.CustomFrameRate = value.toFloat()
					}
				}

				scene.m_scene_frame_rate = GetFramerateFromTimeMode(scene.m_settings.TimeMode, scene.m_settings.CustomFrameRate)
			}
			break

		}
		break

	}
}

func parseObjects(root *Element, scene *Scene) (bool, error) {
	fmt.Println("Starting object Parse")
	objs := findChildren(root, "Objects")
	if objs == nil {
		return true, nil
	}
	scene.m_root = NewNode(scene, root, ROOT)
	scene.m_object_map[0] = ObjectPair{root, scene.m_root}

	objs = objs[0].children
	for _, object := range objs {
		if !isLong(object.first_property) {
			return false, errors.New("Invalid")
		}
		id := object.first_property.value.touint64()
		scene.m_object_map[id] = ObjectPair{object, nil}
	}

	fmt.Println("Iterating through the object map")
	for k, iter := range scene.m_object_map {
		var obj Obj
		var err error
		if iter.object == scene.m_root {
			continue
		}
		fmt.Println("Printing for ", iter.element.id.String())

		if iter.element.id.String() == "Geometry" {
			last_prop := iter.element.first_property
			for last_prop.next != nil {
				last_prop = last_prop.next
			}
			if last_prop != nil && last_prop.value.String() == "Mesh" {
				obj, err = parseGeometry(scene, iter.element)
				if err != nil {
					return false, err
				}
			}
		} else if iter.element.id.String() == "Material" {
			obj = parseMaterial(scene, iter.element)
		} else if iter.element.id.String() == "AnimationStack" {
			obj = NewAnimationStack(scene, iter.element)
			stack := obj.(*AnimationStack)
			scene.m_animation_stacks = append(scene.m_animation_stacks, stack)
		} else if iter.element.id.String() == "AnimationLayer" {
			obj = NewAnimationLayer(scene, iter.element)
		} else if iter.element.id.String() == "AnimationCurve" {
			obj, err = parseAnimationCurve(scene, iter.element)
			if err != nil {
				return false, err
			}
		} else if iter.element.id.String() == "AnimationCurveNode" {
			obj = NewAnimationCurveNode(scene, iter.element)
		} else if iter.element.id.String() == "Deformer" {
			class_prop := iter.element.getProperty(2)
			if class_prop != nil {
				v := class_prop.getValue().String()
				if v == "Cluster" {
					obj, err = parseCluster(scene, iter.element)
					if err != nil {
						return false, err
					}
				} else if v == "Skin" {
					obj = NewSkin(scene, iter.element)
				}
			}
		} else if iter.element.id.String() == "NodeAttribute" {
			obj, err = parseNodeAttribute(scene, iter.element)
			if err != nil {
				return false, err
			}
		} else if iter.element.id.String() == "Model" {
			class_prop := iter.element.getProperty(2)
			if class_prop != nil {
				v := class_prop.getValue().String()
				if v == "Mesh" {
					obj, err = parseMesh(scene, iter.element)
					if err != nil {
						mesh := obj.(*Mesh)
						scene.m_meshes = append(scene.m_meshes, mesh)
						obj = mesh
					}
				} else if v == "LimbNode" {
					obj, err = parseLimbNode(scene, iter.element)
					if err != nil {
						return false, err
					}
				} else if v == "Null" || v == "Root" {
					obj = NewNode(scene, iter.element, NULL_NODE)
				}
			}
		} else if iter.element.id.String() == "Texture" {
			obj = parseTexture(scene, iter.element)
		}

		scene.m_object_map[k] = ObjectPair{iter.element, obj}
		if obj != nil {
			scene.m_all_objects = append(scene.m_all_objects, obj)
			obj.SetID(k)
		}
	}

	fmt.Println("Parsing connections")
	for _, con := range scene.m_connections {
		parent := scene.m_object_map[con.to].object
		child := scene.m_object_map[con.from].object
		if child == nil || parent == nil {
			continue
		}

		ctyp := child.Type()

		switch ctyp {
		case NODE_ATTRIBUTE:
			if parent.Node_attribute() != nil {
				return false, errors.New("Invalid node attribute")
			}
			parent.SetNodeAttribute(child) //previously asserted that the child was a nodeattribute
		case ANIMATION_CURVE_NODE:
			if parent.IsNode() {
				node := child.(*AnimationCurveNode)
				node.bone = parent
				node.bone_link_property = con.property
			}
		}

		switch parent.Type() {
		case MESH:
			{
				mesh := parent.(*Mesh)
				switch ctyp {
				case GEOMETRY:
					if mesh.geometry != nil {
						return false, errors.New("Invalid mesh")
					}
					mesh.geometry = child.(*Geometry)
				case MATERIAL:
					mesh.materials = append(mesh.materials, (child.(*Material)))
				}
			}
		case SKIN:
			{
				skin := parent.(*Skin)
				if ctyp == CLUSTER {
					cluster := child.(*Cluster)
					skin.clusters = append(skin.clusters, cluster)
					if cluster.skin != nil {
						return false, errors.New("Invalid cluster")
					}
					cluster.skin = skin
				}
			}
		case MATERIAL:
			mat := parent.(*Material)
			if ctyp == TEXTURE {
				ttyp := TextureCOUNT
				if con.property == "NormalMap" {
					ttyp = NORMAL
				} else if con.property == "DiffuseColor" {
					ttyp = DIFFUSE
				}
				if ttyp == TextureCOUNT {
					break
				}
				if mat.textures[ttyp] != nil {
					break
				}
				mat.textures[ttyp] = child.(*Texture)
			}
		case GEOMETRY:
			geom := parent.(*Geometry)
			if ctyp == SKIN {
				geom.skin = child.(*Skin)
			}
		case CLUSTER:
			cluster := parent.(*Cluster)
			if ctyp == LIMB_NODE || ctyp == MESH || ctyp == NULL_NODE {
				if cluster.link != nil {
					return false, errors.New("Invalid cluster")
				}
				cluster.link = child
			}

		case ANIMATION_LAYER:
			if ctyp == ANIMATION_CURVE_NODE {
				p := parent.(*AnimationLayer)
				p.curve_nodes = append(p.curve_nodes, child.(*AnimationCurveNode))
			}

		case ANIMATION_CURVE_NODE:
			node := parent.(*AnimationCurveNode)
			if ctyp == ANIMATION_CURVE {
				if node.curves[0].curve == nil {
					node.curves[0].connection = &con
					node.curves[0].curve = child.(*AnimationCurve)
				} else if node.curves[1].curve == nil {
					node.curves[1].connection = &con
					node.curves[1].curve = child.(*AnimationCurve)
				} else if node.curves[2].curve == nil {
					node.curves[2].connection = &con
					node.curves[2].curve = child.(*AnimationCurve)
				} else {
					return false, errors.New("Invalid animation node")
				}
			}
		}
	}

	fmt.Println("Parsing clusters?")
	for _, iter := range scene.m_object_map {
		obj := iter.object
		if obj == nil {
			continue
		}
		if obj.Type() == CLUSTER {
			if !iter.object.(*Cluster).postProcess() {
				return false, errors.New("Failed to postprocess cluster")
			}
		}
	}

	return true, nil
}

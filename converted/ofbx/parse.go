package ofbx

import (
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/oakmound/oak/alg/floatgeom"
	"github.com/pkg/errors"
)

func parseTemplates(root *Element) {
	defs := findChildren(root, "Definitions")
	if defs == nil {
		return
	}

	templates := make(map[string]*Element)
	defs = defs[0].Children
	for _, def := range defs {
		if def.ID.String() == "ObjectType" {
			prop1 := def.getProperty(0).value
			prop1Data, err := ioutil.ReadAll(prop1)
			if err != nil && err != io.EOF {
				//fmt.Println(err)
				continue
			}
			subdefs := def.Children
			for _, subdef := range subdefs {
				if subdef.ID.String() == "PropertyTemplate" {
					prop2 := subdef.getProperty(0).value
					prop2Data, err := ioutil.ReadAll(prop2)
					if err != nil && err != io.EOF {
						//fmt.Println(err)
						continue
					}
					templates[string(prop1Data)+string(prop2Data)] = subdef
				}
			}

		}
	}
}

func parseBinaryArrayInt(property *Property) ([]int, error) {
	count := property.Count
	if count == 0 {
		return []int{}, nil
	}
	if !property.Type.IsArray() {
		return nil, errors.New("Invalid type")
	}
	return parseArrayRawInt(property)
}

func parseBinaryArrayFloat64(property *Property) ([]float64, error) {
	count := property.Count
	if count == 0 {
		return []float64{}, nil
	}
	if !property.Type.IsArray() {
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
func parseBinaryArrayVec2(property *Property) ([]floatgeom.Point2, error) {
	f64s, err := parseBinaryArrayFloat64(property)
	if err != nil {
		return nil, err
	}
	vs := make([]floatgeom.Point2, len(f64s)/2)
	for i := 0; i < len(f64s); i += 2 {
		vs[i/2][0] = f64s[i]
		vs[i/2][1] = f64s[i+1]
	}
	return vs, nil
}
func parseBinaryArrayVec3(property *Property) ([]floatgeom.Point3, error) {
	f64s, err := parseBinaryArrayFloat64(property)
	if err != nil {
		return nil, err
	}
	vs := make([]floatgeom.Point3, len(f64s)/3)
	// len(f64s) should probably be divisible by 3
	if len(f64s)%3 != 0 {
		//fmt.Println("Vec3 binary array not made up of floatgeom.Point3s")
	}
	for i := 0; (i + 2) < len(f64s); i += 3 {
		vs[i/3][0] = f64s[i]
		vs[i/3][1] = f64s[i+1]
		vs[i/3][2] = f64s[i+2]
	}
	return vs, nil
}
func parseBinaryArrayVec4(property *Property) ([]floatgeom.Point4, error) {
	f64s, err := parseBinaryArrayFloat64(property)
	if err != nil {
		return nil, err
	}
	vs := make([]floatgeom.Point4, len(f64s)/4)
	for i := 0; i < len(f64s); i += 4 {
		vs[i/4][0] = f64s[i]
		vs[i/4][1] = f64s[i+1]
		vs[i/4][2] = f64s[i+2]
		vs[i/4][3] = f64s[i+3]
	}
	return vs, nil
}

func parseArrayRawInt(property *Property) ([]int, error) {
	if property.Type == 'd' || property.Type == 'f' {
		return nil, errors.New("Invalid type, expected i or l")
	}
	if property.Encoding == 0 {
		return parseArrayRawIntEnd(property.value, property.Count, property.Type.Size()), nil
	} else if property.Encoding == 1 {
		zr, err := zlib.NewReader(&property.value.Reader)
		if err != nil {
			return nil, errors.Wrap(err, "New Reader failed")
		}
		defer zr.Close()
		return parseArrayRawIntEnd(zr, property.Count, property.Type.Size()), nil
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
	if property.Type == 'd' || property.Type == 'f' {
		return nil, errors.New("Invalid type, expected i or l")
	}
	if property.Encoding == 0 {
		return parseArrayRawInt64End(property.value, property.Count, property.Type.Size()), nil
	} else if property.Encoding == 1 {
		zr, err := zlib.NewReader(&property.value.Reader)
		if err != nil {
			return nil, errors.Wrap(err, "New Reader failed")
		}
		defer zr.Close()
		return parseArrayRawInt64End(zr, property.Count, property.Type.Size()), nil
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
	if property.Type == 'i' || property.Type == 'l' {
		return nil, errors.New("Invalid type, expected d or f")
	}
	if property.Encoding == 0 {
		return parseArrayRawFloat32End(property.value, property.Count, property.Type.Size()), nil
	} else if property.Encoding == 1 {
		zr, err := zlib.NewReader(&property.value.Reader)
		if err != nil {
			return nil, errors.Wrap(err, "New Reader failed")
		}
		defer zr.Close()
		return parseArrayRawFloat32End(zr, property.Count, property.Type.Size()), nil
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
	if property.Type == 'i' || property.Type == 'l' {
		return nil, errors.New("Invalid type, expected d or f")
	}
	if property.Encoding == 0 {
		return parseArrayRawFloat64End(property.value, property.Count, property.Type.Size()), nil
	} else if property.Encoding == 1 {
		zr, err := zlib.NewReader(&property.value.Reader)
		if err != nil {
			return nil, errors.Wrap(err, "New Reader failed")
		}
		defer zr.Close()
		return parseArrayRawFloat64End(zr, property.Count, property.Type.Size()), nil
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

func parseDoubleVecDataVec2(property *Property) ([]floatgeom.Point2, error) {
	if property.Type == 'd' {
		return parseBinaryArrayVec2(property)
	}
	tmp, err := parseBinaryArrayFloat32(property)
	if err != nil {
		return nil, err
	}
	size := 2
	out_vec := make([]floatgeom.Point2, len(tmp)/size)
	for i := 0; i < len(tmp); i += size {
		j := i / size
		out_vec[j][0] = float64(tmp[i])
		out_vec[j][1] = float64(tmp[i+1])
	}
	return out_vec, nil
}

func parseDoubleVecDataVec3(property *Property) ([]floatgeom.Point3, error) {
	if property.Type == 'd' {
		return parseBinaryArrayVec3(property)
	}
	tmp, err := parseBinaryArrayFloat32(property)
	if err != nil {
		return nil, err
	}
	size := 3
	out_vec := make([]floatgeom.Point3, len(tmp)/size)
	for i := 0; i < len(tmp); i += size {
		j := i / size
		out_vec[j][0] = float64(tmp[i])
		out_vec[j][1] = float64(tmp[i+1])
		out_vec[j][2] = float64(tmp[i+2])
	}
	return out_vec, nil
}

func parseDoubleVecDataVec4(property *Property) ([]floatgeom.Point4, error) {
	if property.Type == 'd' {
		return parseBinaryArrayVec4(property)
	}
	tmp, err := parseBinaryArrayFloat32(property)
	if err != nil {
		return nil, err
	}
	size := 4
	out_vec := make([]floatgeom.Point4, len(tmp)/size)
	for i := 0; i < len(tmp); i += size {
		j := i / size
		out_vec[j][0] = float64(tmp[i])
		out_vec[j][1] = float64(tmp[i+1])
		out_vec[j][2] = float64(tmp[i+2])
		out_vec[j][3] = float64(tmp[i+3])
	}
	return out_vec, nil
}

func parseVertexDataVec2(element *Element, name, index_name string) ([]floatgeom.Point2, []int, VertexDataMapping, error) {
	idxs, mapping, dataProp, err := parseVertexDataInner(element, name, index_name)
	vcs, err := parseDoubleVecDataVec2(dataProp)
	return vcs, idxs, mapping, err
}

func parseVertexDataVec3(element *Element, name, index_name string) ([]floatgeom.Point3, []int, VertexDataMapping, error) {
	idxs, mapping, dataProp, err := parseVertexDataInner(element, name, index_name)
	vcs, err := parseDoubleVecDataVec3(dataProp)
	return vcs, idxs, mapping, err
}

func parseVertexDataVec4(element *Element, name, index_name string) ([]floatgeom.Point4, []int, VertexDataMapping, error) {
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

	if len(mappingProp) != 0 {
		s := mappingProp[0].value.String()
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
	if len(referenceProp) != 0 {
		if referenceProp[0].value.String() == "IndexToDirect" {
			indicesProp := findChildProperty(element, index_name)
			if len(indicesProp) != 0 {
				if idxs, err = parseBinaryArrayInt(indicesProp[0]); err != nil {
					return nil, 0, nil, errors.New("Unable to parse indices")
				}
			}
		} else if referenceProp[0].value.String() != "Direct" {
			return nil, 0, nil, errors.New("Invalid properties")
		}
	}
	return idxs, mapping, dataProp[0], nil
}

func parseTexture(scene *Scene, element *Element) *Texture {
	texture := NewTexture(scene, element)
	textureFilenameProp := findSingleChildProperty(element, "FileName")
	if textureFilenameProp != nil {
		texture.filename = textureFilenameProp.value
	}
	textureRelativeFilenameProp := findSingleChildProperty(element, "RelativeFilename")
	if textureRelativeFilenameProp != nil {
		texture.relative_filename = textureRelativeFilenameProp.value
	}
	return texture
}

func parseLimbNode(scene *Scene, element *Element) (*Node, error) {
	if prop := element.getProperty(2); prop == nil || prop.value.String() != "LimbNode" {
		return nil, errors.New("Invalid limb node")
	}
	return NewNode(scene, element, LIMB_NODE), nil
}

func parseMesh(scene *Scene, element *Element) (*Mesh, error) {
	if prop := element.getProperty(2); prop == nil || prop.value.String() != "Mesh" {
		return nil, errors.New("Invalid mesh")
	}
	return NewMesh(scene, element), nil
}

func parseMaterial(scene *Scene, element *Element) *Material {
	material := NewMaterial(scene, element)
	elems := findChildren(element, "Properties70")
	material.DiffuseColor = Color{1, 1, 1}
	if len(elems) == 0 {
		return material
	}
	elems = elems[0].Children
	// Todo: reflection / struct tags for these types of values
	for _, elem := range elems {
		if elem.getProperty(0) == nil {
			continue
		}
		v := elem.getProperty(0).value.String()
		// Commented out cases are things I (200sc) think might exist
		// but haven't seen
		switch v {
		case "EmissiveColor":
			material.EmissiveColor.R = float32(elem.getProperty(4).value.toDouble())
			material.EmissiveColor.G = float32(elem.getProperty(5).value.toDouble())
			material.EmissiveColor.B = float32(elem.getProperty(6).value.toDouble())
		case "AmbientColor":
			material.AmbientColor.R = float32(elem.getProperty(4).value.toDouble())
			material.AmbientColor.G = float32(elem.getProperty(5).value.toDouble())
			material.AmbientColor.B = float32(elem.getProperty(6).value.toDouble())
		case "DiffuseColor":
			material.DiffuseColor.R = float32(elem.getProperty(4).value.toDouble())
			material.DiffuseColor.G = float32(elem.getProperty(5).value.toDouble())
			material.DiffuseColor.B = float32(elem.getProperty(6).value.toDouble())
		case "TransparentColor":
			material.TransparentColor.R = float32(elem.getProperty(4).value.toDouble())
			material.TransparentColor.G = float32(elem.getProperty(5).value.toDouble())
			material.TransparentColor.B = float32(elem.getProperty(6).value.toDouble())
		case "SpecularColor":
			material.SpecularColor.R = float32(elem.getProperty(4).value.toDouble())
			material.SpecularColor.G = float32(elem.getProperty(5).value.toDouble())
			material.SpecularColor.B = float32(elem.getProperty(6).value.toDouble())
		case "ReflectionColor":
			material.ReflectionColor.R = float32(elem.getProperty(4).value.toDouble())
			material.ReflectionColor.G = float32(elem.getProperty(5).value.toDouble())
			material.ReflectionColor.B = float32(elem.getProperty(6).value.toDouble())
		case "EmissiveFactor":
			material.EmissiveFactor = elem.getProperty(4).value.toDouble()
		// case "AmbientFactor":
		case "DiffuseFactor":
			material.DiffuseFactor = elem.getProperty(4).value.toDouble()
		// case "TransparentFactor":
		case "SpecularFactor":
			material.SpecularFactor = elem.getProperty(4).value.toDouble()
		case "ReflectionFactor":
			material.ReflectionFactor = elem.getProperty(4).value.toDouble()
		case "Shininess":
			material.Shininess = elem.getProperty(4).value.toDouble()
		case "ShininessExponent":
			material.ShininessExponent = elem.getProperty(4).value.toDouble()
		}
	}
	return material
}

func parseAnimationCurve(scene *Scene, element *Element) (*AnimationCurve, error) {
	curve := &AnimationCurve{}
	var err error
	if attrFlags := findSingleChildProperty(element, "KeyAttrFlags"); attrFlags != nil {
		curve.AttrFlags, err = attrFlags.getValuesInt64()
		if err != nil {
			return nil, errors.Wrap(err, "Invalid animation curve: attrFlags error")
		}
	}
	if attrData := findSingleChildProperty(element, "KeyAttrDataFloat"); attrData != nil {
		curve.AttrData, err = attrData.getValuesF32()
		if err != nil {
			return nil, errors.Wrap(err, "Invalid animation curve: attrFlags error")
		}
	}
	if attrRefCt := findSingleChildProperty(element, "KeyAttrRefCount"); attrRefCt != nil {
		curve.AttrRefCount, err = attrRefCt.getValuesInt64()
		if err != nil {
			return nil, errors.Wrap(err, "Invalid animation curve: attrFlags error")
		}
	}
	if times := findSingleChildProperty(element, "KeyTime"); times != nil {
		curve.Times, err = times.getValuesInt64()
		if err != nil {
			return nil, errors.Wrap(err, "Invalid animation curve: times error")
		}
	}
	if values := findSingleChildProperty(element, "KeyValueFloat"); values != nil {
		curve.Values, err = values.getValuesF32()
		if err != nil {
			return nil, errors.New("Invalid animation curve: values error")
		}
	}
	if len(curve.Times) != len(curve.Values) {
		return nil, errors.New("Invalid animation curve: len error")
	}
	return curve, nil
}

func parseConnection(root *Element, scene *Scene) (bool, error) {
	connections := findChildren(root, "Connections")
	if connections == nil {
		return true, nil
	}

	connections = connections[0].Children
	for _, connection := range connections {
		prop0 := connection.getProperty(0)
		prop1 := connection.getProperty(1)
		prop2 := connection.getProperty(2)
		if !isString(prop0) ||
			!isLong(prop1) ||
			!isLong(prop2) {
			return false, errors.New("Invalid connection")
		}
		var c Connection
		c.from = prop1.value.touint64()
		c.to = prop2.value.touint64()
		if prop0.value.String() == "OO" {
			c.typ = OBJECT_OBJECT
		} else if prop0.value.String() == "OP" {
			c.typ = OBJECT_PROPERTY
			if prop3 := connection.getProperty(3); prop3 != nil {
				c.property = prop3.value.String()
			} else {
				return false, errors.New("Invalid connection")
			}
		} else {
			return false, errors.New("Not supported")
		}
		scene.connections = append(scene.connections, c)
	}
	return true, nil
}

func parseTakes(scene *Scene) (bool, error) {
	takes := findChildren(scene.RootElement, "Takes")
	if takes == nil {
		return true, nil
	}

	objects := takes[0].Children

	for _, object := range objects {
		if object.ID.String() != "Take" {
			continue
		}
		if !isString(object.getProperty(0)) {
			return false, errors.New("Invalid name in take")
		}
		var take TakeInfo
		take.name = object.getProperty(0).value
		filename := findSingleChildProperty(object, "FileName")
		if filename != nil {
			if !isString(filename) {
				return false, errors.New("Invalid filename in take")
			}
			take.filename = filename.value
		}
		local_time := findChildProperty(object, "LocalTime")
		if len(local_time) != 0 {
			if !isLong(local_time[0]) || len(local_time) < 2 || !isLong(local_time[1]) {
				return false, errors.New("Invalid local time in take")
			}

			take.local_time_from = fbxTimeToSeconds(local_time[0].value.toint64())
			take.local_time_to = fbxTimeToSeconds(local_time[1].value.toint64())
		}
		reference_time := findChildProperty(object, "ReferenceTime")
		if len(reference_time) != 0 {
			if !isLong(reference_time[0]) || len(reference_time) < 2 || !isLong(reference_time[1]) {
				return false, errors.New("Invalid reference time in take")
			}
			take.reference_time_from = fbxTimeToSeconds(reference_time[0].value.toint64())
			take.reference_time_to = fbxTimeToSeconds(reference_time[1].value.toint64())
		}
		scene.takeInfos = append(scene.takeInfos, take)
	}
	return true, nil
}

func parseGlobalSettings(root *Element, scene *Scene) {

	for _, settings := range root.Children {
		if settings.ID.String() != "GlobalSettings" {
			continue
		}
		for _, props70 := range settings.Children {
			if props70.ID.String() != "Properties70" {
				continue
			}
			for _, node := range props70.Children {
				p := node.getProperty(0)
				if p == nil {
					continue
				}
				prop4 := node.getProperty(4)
				if prop4 == nil {
					continue
				}
				value := prop4.value

				switch p.value.String() {
				case "UpAxis":
					scene.settings.UpAxis = UpVector(int(value.toInt32()))
				case "UpAxisSign":
					scene.settings.UpAxisSign = int(value.toInt32())
				case "FrontAxis":
					scene.settings.FrontAxis = FrontVector(int(value.toInt32()))
				case "FrontAxisSign":
					scene.settings.FrontAxisSign = int(value.toInt32())
				case "CoordAxis":
					scene.settings.CoordAxis = CoordSystem(int(value.toInt32()))
				case "CoordAxisSign":
					scene.settings.CoordAxisSign = int(value.toInt32())
				case "OriginalUpAxis":
					scene.settings.OriginalUpAxis = int(value.toInt32())
				case "OriginalUpAxisSign":
					scene.settings.OriginalUpAxisSign = int(value.toInt32())
				case "UnitScaleFactor":
					scene.settings.UnitScaleFactor = value.toFloat()
				case "OriginalUnitScaleFactor":
					scene.settings.OriginalUnitScaleFactor = value.toFloat()
				case "TimeSpanStart":
					scene.settings.TimeSpanStart = value.touint64()
				case "TimeSpanStop":
					scene.settings.TimeSpanStop = value.touint64()
				case "TimeMode":
					scene.settings.TimeMode = FrameRate(int(value.toInt32()))
				case "CustomFrameRate":
					scene.settings.CustomFrameRate = value.toFloat()
				}
			}
			break
		}
		break
	}
	scene.frameRate = GetFramerateFromTimeMode(scene.settings.TimeMode, scene.settings.CustomFrameRate)
}

func parseObjects(root *Element, scene *Scene) (bool, error) {
	//fmt.Println("Starting object Parse")
	objs := findChildren(root, "Objects")
	if objs == nil {
		return true, nil
	}
	scene.RootNode = NewNode(scene, root, ROOT)
	scene.objectMap[0] = ObjectPair{root, scene.RootNode}

	objs = objs[0].Children
	for _, object := range objs {
		if !isLong(object.getProperty(0)) {
			return false, errors.New("Invalid")
		}
		id := object.getProperty(0).value.touint64()
		scene.objectMap[id] = ObjectPair{object, nil}
	}

	//fmt.Println("Iterating through the object map")
	for k, iter := range scene.objectMap {
		var obj Obj
		var err error
		if iter.object == scene.RootNode {
			continue
		}
		//fmt.Println("Printing for ", iter.element.id.String())

		if iter.element.ID.String() == "Geometry" {
			last_prop := iter.element.getProperty(len(iter.element.Properties) - 1)
			if last_prop != nil && last_prop.value.String() == "Mesh" {
				obj, err = parseGeometry(scene, iter.element)
				if err != nil {
					return false, err
				}
			}
		} else if iter.element.ID.String() == "Material" {
			obj = parseMaterial(scene, iter.element)
		} else if iter.element.ID.String() == "AnimationStack" {
			obj = NewAnimationStack(scene, iter.element)
			stack := obj.(*AnimationStack)
			scene.AnimationStacks = append(scene.AnimationStacks, stack)
		} else if iter.element.ID.String() == "AnimationLayer" {
			obj = NewAnimationLayer(scene, iter.element)
		} else if iter.element.ID.String() == "AnimationCurve" {
			obj, err = parseAnimationCurve(scene, iter.element)
			if err != nil {
				return false, err
			}
		} else if iter.element.ID.String() == "AnimationCurveNode" {
			obj = NewAnimationCurveNode(scene, iter.element)
		} else if iter.element.ID.String() == "Deformer" {
			class_prop := iter.element.getProperty(2)
			if class_prop != nil {
				v := class_prop.value.String()
				if v == "Cluster" {
					obj, err = parseCluster(scene, iter.element)
					if err != nil {
						return false, err
					}
				} else if v == "Skin" {
					obj = NewSkin(scene, iter.element)
				}
			}
		} else if iter.element.ID.String() == "NodeAttribute" {
			obj, err = parseNodeAttribute(scene, iter.element)
			if err != nil {
				return false, err
			}
		} else if iter.element.ID.String() == "Model" {
			class_prop := iter.element.getProperty(2)
			if class_prop != nil {
				v := class_prop.value.String()
				if v == "Mesh" {
					obj, err = parseMesh(scene, iter.element)
					if err == nil {
						mesh := obj.(*Mesh)
						scene.Meshes = append(scene.Meshes, mesh)
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
		} else if iter.element.ID.String() == "Texture" {
			obj = parseTexture(scene, iter.element)
		}

		scene.objectMap[k] = ObjectPair{iter.element, obj}
		if obj != nil {
			scene.Objects = append(scene.Objects, obj)
			obj.SetID(k)
		}
	}

	//fmt.Println("Parsing connections")
	for _, con := range scene.connections {
		con := con
		parent := scene.objectMap[con.to].object
		child := scene.objectMap[con.from].object
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
				node.Bone = parent
				node.boneLinkProp = con.property
			}
		}

		switch parent.Type() {
		case MESH:
			mesh := parent.(*Mesh)
			switch ctyp {
			case GEOMETRY:
				if mesh.Geometry != nil {
					return false, errors.New("Invalid mesh")
				}
				mesh.Geometry = child.(*Geometry)
			case MATERIAL:
				mesh.Materials = append(mesh.Materials, (child.(*Material)))
			}
		case SKIN:
			skin := parent.(*Skin)
			if ctyp == CLUSTER {
				cluster := child.(*Cluster)
				skin.clusters = append(skin.clusters, cluster)
				if cluster.Skin != nil {
					return false, errors.New("Invalid cluster")
				}
				cluster.Skin = skin
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
				if mat.Textures[ttyp] != nil {
					break
				}
				mat.Textures[ttyp] = child.(*Texture)
			}
		case GEOMETRY:
			geom := parent.(*Geometry)
			if ctyp == SKIN {
				geom.Skin = child.(*Skin)
			}
		case CLUSTER:
			cluster := parent.(*Cluster)
			if ctyp == LIMB_NODE || ctyp == MESH || ctyp == NULL_NODE {
				if cluster.Link != nil {
					return false, errors.New("Invalid cluster")
				}
				cluster.Link = child
			}

		case ANIMATION_LAYER:
			if ctyp == ANIMATION_CURVE_NODE {
				p := parent.(*AnimationLayer)
				p.CurveNodes = append(p.CurveNodes, child.(*AnimationCurveNode))
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

	for _, iter := range scene.objectMap {
		obj := iter.object
		if obj == nil {
			continue
		}
		if ppr, ok := obj.(NeedsPostProcessing); ok {
			if !ppr.postProcess() {
				return false, errors.New("Failed to postprocess object" + fmt.Sprintf("%v", obj.ID()))
			}
		}
	}

	return true, nil
}

// NeedsPostProcessing note objects that require post processing
type NeedsPostProcessing interface {
	postProcess() bool
}

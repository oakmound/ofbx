package ofbx

import (
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"errors"

	"github.com/oakmound/oak/v4/alg/floatgeom"
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
	//if len(f64s)%3 != 0 {
	//fmt.Println("Vec3 binary array not made up of floatgeom.Point3s")
	//}
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
			return nil, fmt.Errorf("New Reader failed: %w", err)
		}
		defer zr.Close()
		return parseArrayRawIntEnd(zr, property.Count, property.Type.Size()), nil
	}
	return nil, errors.New("Invalid encoding")
}

func parseArrayRawIntEnd(r io.Reader, ln int, elemSize int) []int {
	if elemSize == 4 {
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
			return nil, fmt.Errorf("New Reader failed: %w", err)
		}
		defer zr.Close()
		return parseArrayRawInt64End(zr, property.Count, property.Type.Size()), nil
	}
	return nil, errors.New("Invalid encoding")
}

func parseArrayRawInt64End(r io.Reader, ln int, elemSize int) []int64 {
	if elemSize == 4 {
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
			return nil, fmt.Errorf("New Reader failed: %w", err)
		}
		defer zr.Close()
		return parseArrayRawFloat32End(zr, property.Count, property.Type.Size()), nil
	}
	return nil, errors.New("Invalid encoding")
}

func parseArrayRawFloat32End(r io.Reader, ln int, elemSize int) []float32 {
	if elemSize == 4 {
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
			return nil, fmt.Errorf("New Reader failed: %w", err)
		}
		defer zr.Close()
		return parseArrayRawFloat64End(zr, property.Count, property.Type.Size()), nil
	}
	return nil, errors.New("Invalid encoding")
}

func parseArrayRawFloat64End(r io.Reader, ln int, elemSize int) []float64 {
	if elemSize == 4 {
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
	outVec := make([]floatgeom.Point2, len(tmp)/size)
	for i := 0; i < len(tmp); i += size {
		j := i / size
		outVec[j][0] = float64(tmp[i])
		outVec[j][1] = float64(tmp[i+1])
	}
	return outVec, nil
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
	outVec := make([]floatgeom.Point3, len(tmp)/size)
	for i := 0; i < len(tmp); i += size {
		j := i / size
		outVec[j][0] = float64(tmp[i])
		outVec[j][1] = float64(tmp[i+1])
		outVec[j][2] = float64(tmp[i+2])
	}
	return outVec, nil
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
	outVec := make([]floatgeom.Point4, len(tmp)/size)
	for i := 0; i < len(tmp); i += size {
		j := i / size
		outVec[j][0] = float64(tmp[i])
		outVec[j][1] = float64(tmp[i+1])
		outVec[j][2] = float64(tmp[i+2])
		outVec[j][3] = float64(tmp[i+3])
	}
	return outVec, nil
}

func parseVertexDataVec2(element *Element, name, idxName string) ([]floatgeom.Point2, []int, VertexDataMapping, error) {
	idxs, mapping, dataProp, err := parseVertexDataInner(element, name, idxName)
	if err != nil {
		return nil, nil, mapping, err
	}
	vcs, err := parseDoubleVecDataVec2(dataProp)
	return vcs, idxs, mapping, err
}

func parseVertexDataVec3(element *Element, name, idxName string) ([]floatgeom.Point3, []int, VertexDataMapping, error) {
	idxs, mapping, dataProp, err := parseVertexDataInner(element, name, idxName)
	if err != nil {
		return nil, nil, mapping, err
	}
	vcs, err := parseDoubleVecDataVec3(dataProp)
	return vcs, idxs, mapping, err
}

func parseVertexDataVec4(element *Element, name, idxName string) ([]floatgeom.Point4, []int, VertexDataMapping, error) {
	idxs, mapping, dataProp, err := parseVertexDataInner(element, name, idxName)
	if err != nil {
		return nil, nil, mapping, err
	}
	vcs, err := parseDoubleVecDataVec4(dataProp)
	return vcs, idxs, mapping, err
}

func parseVertexDataInner(element *Element, name, idxName string) ([]int, VertexDataMapping, *Property, error) {
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
		var ok bool
		mapping, ok = vtxDataMapFromStrs[mappingProp[0].value.String()]
		if !ok {
			return nil, 0, nil, errors.New("Unable to parse mapping")
		}
	}
	if len(referenceProp) != 0 {
		if referenceProp[0].value.String() == "IndexToDirect" {
			indicesProp := findChildProperty(element, idxName)
			if len(indicesProp) != 0 {
				if idxs, err = parseBinaryArrayInt(indicesProp[0]); err != nil {
					return nil, 0, nil, errors.New("Unable to parse indices")
				}
			} else {
				// just use indicies in order.
			}
		} else if referenceProp[0].value.String() != "Direct" {
			return nil, 0, nil, errors.New("Invalid properties")
		}
	}
	return idxs, mapping, dataProp[0], nil
}

func parseTexture(scene *Scene, element *Element) *Texture {
	texture := NewTexture(scene, element)
	assignSingleChildProperty(element, "FileName", texture.filename)
	assignSingleChildProperty(element, "RelativeFilename", texture.relativeFilename)
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
		p := elem.getProperty(0)
		if p == nil {
			continue
		}
		v := p.value.String()
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
	curve := NewAnimationCurve(scene, element)
	var err error
	if attrFlags := findSingleChildProperty(element, "KeyAttrFlags"); attrFlags != nil {
		curve.AttrFlags, err = attrFlags.getValuesInt64()
		if err != nil {
			return nil, fmt.Errorf("Invalid animation curve: attrFlags error: %w", err)
		}
	}
	if attrData := findSingleChildProperty(element, "KeyAttrDataFloat"); attrData != nil {
		curve.AttrData, err = attrData.getValuesF32()
		if err != nil {
			return nil, fmt.Errorf("Invalid animation curve: attrFlags error: %w", err)
		}
	}
	if attrRefCt := findSingleChildProperty(element, "KeyAttrRefCount"); attrRefCt != nil {
		curve.AttrRefCount, err = attrRefCt.getValuesInt64()
		if err != nil {
			return nil, fmt.Errorf("Invalid animation curve: attrFlags error: %w", err)
		}
	}
	if times := findSingleChildProperty(element, "KeyTime"); times != nil {
		intTimes, err := times.getValuesInt64()
		if err != nil {
			return nil, fmt.Errorf("Invalid animation curve: times error: %w", err)
		}
		curve.Times = make([]time.Duration, len(intTimes))
		for i, v := range intTimes {
			curve.Times[i] = fbxTimetoStdTime(v)
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
		c.From = prop1.value.touint64()
		c.To = prop2.value.touint64()
		if prop0.value.String() == "OO" {
			c.Typ = ObjectConn
		} else if prop0.value.String() == "OP" {
			c.Typ = PropConn
			if prop3 := connection.getProperty(3); prop3 != nil {
				c.Property = prop3.value.String()
			} else {
				return false, errors.New("Invalid connection")
			}
		} else {
			return false, errors.New("Not supported")
		}
		scene.Connections = append(scene.Connections, c)
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
		localTime := findChildProperty(object, "LocalTime")
		if len(localTime) != 0 {
			if !isLong(localTime[0]) || len(localTime) < 2 || !isLong(localTime[1]) {
				return false, errors.New("Invalid local time in take")
			}

			take.localTimeFrom = fbxTimeToSeconds(localTime[0].value.toint64())
			take.localTimeTo = fbxTimeToSeconds(localTime[1].value.toint64())
		}
		refTime := findChildProperty(object, "ReferenceTime")
		if len(refTime) != 0 {
			if !isLong(refTime[0]) || len(refTime) < 2 || !isLong(refTime[1]) {
				return false, errors.New("Invalid reference time in take")
			}
			take.refTimeFrom = fbxTimeToSeconds(refTime[0].value.toint64())
			take.refTimeTo = fbxTimeToSeconds(refTime[1].value.toint64())
		}
		scene.TakeInfos = append(scene.TakeInfos, take)
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
					scene.Settings.UpAxis = UpVector(int(value.toInt32()))
				case "UpAxisSign":
					scene.Settings.UpAxisSign = int(value.toInt32())
				case "FrontAxis":
					scene.Settings.FrontAxis = FrontVector(int(value.toInt32()))
				case "FrontAxisSign":
					scene.Settings.FrontAxisSign = int(value.toInt32())
				case "CoordAxis":
					scene.Settings.CoordAxis = CoordSystem(int(value.toInt32()))
				case "CoordAxisSign":
					scene.Settings.CoordAxisSign = int(value.toInt32())
				case "OriginalUpAxis":
					scene.Settings.OriginalUpAxis = int(value.toInt32())
				case "OriginalUpAxisSign":
					scene.Settings.OriginalUpAxisSign = int(value.toInt32())
				case "UnitScaleFactor":
					scene.Settings.UnitScaleFactor = value.toFloat()
				case "OriginalUnitScaleFactor":
					scene.Settings.OriginalUnitScaleFactor = value.toFloat()
				case "TimeSpanStart":
					scene.Settings.TimeSpanStart = value.touint64()
				case "TimeSpanStop":
					scene.Settings.TimeSpanStop = value.touint64()
				case "TimeMode":
					scene.Settings.TimeMode = FrameRate(int(value.toInt32()))
				case "CustomFrameRate":
					scene.Settings.CustomFrameRate = value.toFloat()
				}
			}
			break
		}
		break
	}
	scene.FrameRate = GetFramerateFromTimeMode(scene.Settings.TimeMode, scene.Settings.CustomFrameRate)
}

func parseObjects(root *Element, scene *Scene) (bool, error) {
	//fmt.Println("Starting object Parse")
	objs := findChildren(root, "Objects")
	if objs == nil {
		return true, nil
	}
	scene.RootNode = NewNode(scene, root, ROOT)
	scene.ObjectMap[0] = scene.RootNode

	objs = objs[0].Children
	for _, elem := range objs {
		if !isLong(elem.getProperty(0)) {
			return false, errors.New("Invalid")
		}
		id := elem.getProperty(0).value.touint64()

		var obj Obj
		var err error
		// This shouldn't happen?
		// Original library had a check like this but it seems nonsensical
		if id == 0 {
			continue
		}
		switch elem.ID.String() {
		case "Geometry":
			lastProp := elem.getProperty(len(elem.Properties) - 1)
			if lastProp != nil && lastProp.value.String() == "Mesh" {
				obj, err = parseGeometry(scene, elem)
				if err != nil {
					return false, err
				}
			}
		case "Material":
			obj = parseMaterial(scene, elem)
		case "AnimationStack":
			obj = NewAnimationStack(scene, elem)
			stack := obj.(*AnimationStack)
			scene.AnimationStacks = append(scene.AnimationStacks, stack)
		case "AnimationLayer":
			obj = NewAnimationLayer(scene, elem)
		case "AnimationCurve":
			obj, err = parseAnimationCurve(scene, elem)
			if err != nil {
				return false, err
			}
		case "AnimationCurveNode":
			obj = NewAnimationCurveNode(scene, elem)
		case "Deformer":
			classProp := elem.getProperty(2)
			if classProp != nil {
				v := classProp.value.String()
				if v == "Cluster" {
					obj, err = parseCluster(scene, elem)
					if err != nil {
						return false, err
					}
				} else if v == "Skin" {
					obj = NewSkin(scene, elem)
				}
			}
		case "NodeAttribute":
			obj, err = parseNodeAttribute(scene, elem)
			if err != nil {
				return false, err
			}
		case "Model":
			classProp := elem.getProperty(2)
			if classProp != nil {
				v := classProp.value.String()
				if v == "Mesh" {
					obj, err = parseMesh(scene, elem)
					if err == nil {
						mesh := obj.(*Mesh)
						scene.Meshes = append(scene.Meshes, mesh)
						obj = mesh
					}
				} else if v == "LimbNode" {
					obj, err = parseLimbNode(scene, elem)
					if err != nil {
						return false, err
					}
				} else if v == "Null" || v == "Root" {
					obj = NewNode(scene, elem, NULL_NODE)
				}
			}
		case "Texture":
			obj = parseTexture(scene, elem)
		}

		scene.ObjectMap[id] = obj
		if obj != nil {
			obj.SetID(id)
		}
	}

	//fmt.Println("Parsing connections")
	for _, con := range scene.Connections {
		con := con
		parent := scene.ObjectMap[con.To]
		child := scene.ObjectMap[con.From]
		if child == nil || parent == nil {
			continue
		}

		ctyp := child.Type()

		switch ctyp {
		case NODE_ATTRIBUTE:
			if parent.NodeAttribute() != nil {
				return false, errors.New("Invalid node attribute")
			}
			parent.SetNodeAttribute(child) //previously asserted that the child was a nodeattribute
		case ANIMATION_CURVE_NODE:
			if parent.IsNode() {
				node := child.(*AnimationCurveNode)
				node.Bone = parent
				node.BoneLinkProp = con.Property
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
				skin.Clusters = append(skin.Clusters, cluster)
				if cluster.Skin != nil {
					return false, errors.New("Cluster assigned to multiple skins")
				}
				cluster.Skin = skin
			}
		case MATERIAL:
			mat := parent.(*Material)
			if ctyp == TEXTURE {
				ttyp := TextureCOUNT
				if con.Property == "NormalMap" {
					ttyp = NORMAL
				} else if con.Property == "DiffuseColor" {
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
				if node.Curves[0].Curve == nil {
					node.Curves[0].connection = &con
					node.Curves[0].Curve = child.(*AnimationCurve)
				} else if node.Curves[1].Curve == nil {
					node.Curves[1].connection = &con
					node.Curves[1].Curve = child.(*AnimationCurve)
				} else if node.Curves[2].Curve == nil {
					node.Curves[2].connection = &con
					node.Curves[2].Curve = child.(*AnimationCurve)
				} else {
					return false, errors.New("Invalid animation node")
				}
			}
		}
	}

	for _, obj := range scene.ObjectMap {
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

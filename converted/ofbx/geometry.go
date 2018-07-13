package ofbx

import (
	"errors"
	"fmt"

	"github.com/oakmound/oak/alg/floatgeom"
)

type VertexDataMapping int

const (
	BY_POLYGON_VERTEX = iota
	BY_POLYGON        = iota
	BY_VERTEX         = iota
)

const MaxUvs = 4

type Geometry struct {
	Object
	Skin *Skin

	Vertices, Normals, Tangents []floatgeom.Point3

	UVs                        [MaxUvs][]floatgeom.Point2
	Colors                     []floatgeom.Point4
	Materials, to_old_vertices []int
	to_new_vertices            []Vertex
	Faces                      [][]int
}

func (g *Geometry) String() string {
	return g.stringPrefix("")
}

func (g *Geometry) stringPrefix(prefix string) string {
	s := prefix + "Geometry:\n"
	s += g.Object.String() + "\n"
	if len(g.Vertices) != 0 {
		s += prefix + "Verts:"
		for i, v := range g.Vertices {
			if i != 0 {
				s += ","
			}
			s += fmt.Sprintf("%+v", v)
		}
		s += "\n"
	}
	if len(g.Normals) != 0 {
		s += prefix + "Norms:"
		for i, v := range g.Normals {
			if i != 0 {
				s += ","
			}
			s += fmt.Sprintf("%+v", v)
		}
		s += "\n"
	}
	if len(g.Tangents) != 0 {
		s += prefix + "Tangents:"
		for i, v := range g.Tangents {
			if i != 0 {
				s += ","
			}
			s += fmt.Sprintf("%+v", v)
		}
		s += "\n"
	}
	if len(g.Materials) != 0 {
		s += prefix + "Materials:"
		for i, v := range g.Materials {
			if i != 0 {
				s += ","
			}
			s += fmt.Sprintf("%v", v)
		}
		s += "\n"
	}
	if len(g.Colors) != 0 {
		s += prefix + "Colors:"
		for i, v := range g.Colors {
			if i != 0 {
				s += ","
			}
			s += fmt.Sprintf("%+v", v)
		}
		s += "\n"
	}

	if len(g.Faces) != 0 {
		s += prefix + "Faces:"
		for i, v := range g.Faces {
			if i != 0 {
				s += ", "
			}
			s += fmt.Sprintf("%v", v)
		}
		s += "\n"
	}

	s += prefix + "UVs:"
	for _, v := range g.UVs {
		if len(v) == 0 {
			continue
		}
		for i, v2 := range v {
			if i != 0 {
				s += ","
			}
			s += fmt.Sprintf("%+v", v2)
		}
		s += "\n"
	}
	return s
}

//Hey its a linked list of indices!.....
type Vertex struct {
	index int //should start as -1
	next  *Vertex
}

func add(nv *Vertex, index int) {
	nv.add(index)
}

func (nv *Vertex) add(index int) {
	if nv.index == -1 {
		//TODO: change this cuz we aint implementing it this way. Really its checking if the newvertex exists...
		nv.index = index
	} else if nv.next != nil {
		nv.next.add(index)
	} else {
		nv.next = &Vertex{-1, nil}
		nv.next.index = index
	}
}

func NewGeometry(scene *Scene, element *Element) *Geometry {
	g := &Geometry{}
	g.Object = *NewObject(scene, element)
	g.to_old_vertices = make([]int, 0)
	return g
}

func (g *Geometry) Type() Type {
	return GEOMETRY
}

func (g *Geometry) UVSMax() int {
	return MaxUvs
}

func (g *Geometry) triangulate(old_indices []int) []int {
	to_old := make([]int, 0)
	in_polygon_idx := 0
	for i := 0; i < len(old_indices); i++ {
		idx := old_indices[i]
		if idx < 0 {
			idx = (-idx) - 1
		}

		if in_polygon_idx <= 2 {
			g.to_old_vertices = append(g.to_old_vertices, idx)
			to_old = append(to_old, i)
		} else {
			g.to_old_vertices = append(g.to_old_vertices, old_indices[i-in_polygon_idx])
			to_old = append(to_old, i-in_polygon_idx)
			g.to_old_vertices = append(g.to_old_vertices, old_indices[i-1])
			to_old = append(to_old, i-1)
			g.to_old_vertices = append(g.to_old_vertices, idx)
			to_old = append(to_old, i)
		}
		in_polygon_idx++
		if old_indices[i] < 0 {
			in_polygon_idx = 0
		}
	}
	return to_old
}

func parseGeometry(scene *Scene, element *Element) (*Geometry, error) {
	if element.properties == nil {
		return nil, errors.New("Geometry invalid")
	}
	geom := NewGeometry(scene, element)

	verticesProp := findChildProperty(element, "Vertices")
	if len(verticesProp) == 0 {
		return nil, errors.New("Geometry Vertices Missing")
	}

	polysProp := findChildProperty(element, "PolygonVertexIndex")
	if len(polysProp) == 0 {
		return nil, errors.New("Geometry Indicies missing")
	}
	vertices, err := parseDoubleVecDataVec3(verticesProp[0])
	if err != nil {
		return nil, err
	}
	original_indices, err := parseBinaryArrayInt(polysProp[0])
	if err != nil {
		return nil, err
	}

	geom.Faces = make([][]int, 0)
	curFace := []int{}
	//Parse out the polygons. List of vertex references with a negative value indicating the last vertex of a face.
	for _, v := range original_indices {
		if v < 0 {
			curFace = append(curFace, (v*-1)-1)
			geom.Faces = append(geom.Faces, curFace)
			curFace = []int{}
		} else {
			curFace = append(curFace, v)
		}
	}

	to_old_indices := geom.triangulate(original_indices)
	geom.Vertices = make([]floatgeom.Point3, len(geom.to_old_vertices))

	for i, vIdx := range geom.to_old_vertices {
		v := vertices[vIdx]
		geom.Vertices[i] = v
	}

	geom.to_new_vertices = make([]Vertex, len(geom.Vertices))

	for i := 0; i < len(geom.to_old_vertices); i++ {
		old := geom.to_old_vertices[i]
		geom.to_new_vertices[old].add(i)
	}

	layerMaterialElements := findChildren(element, "LayerElementMaterial")
	if len(layerMaterialElements) > 0 {
		mappingProp := findChildProperty(layerMaterialElements[0], "MappingInformationType")
		referenceProp := findChildProperty(layerMaterialElements[0], "ReferenceInformationType")
		if len(mappingProp) == 0 || len(referenceProp) == 0 {
			return nil, errors.New("Invalid LayerElementMaterial")
		}
		var err error
		tmp := make([]int, 0)

		if mappingProp[0].value.String() == "ByPolygon" &&
			referenceProp[0].value.String() == "IndexToDirect" {
			geom.Materials = make([]int, len(geom.Vertices)/3)
			for i := 0; i < len(geom.Vertices)/3; i++ {
				geom.Materials[i] = -1
			}

			indiciesProp := findChildProperty(layerMaterialElements[0], "Materials")
			if indiciesProp == nil {
				return nil, errors.New("Invalid LayerElementMaterial")
			}

			tmp, err = parseBinaryArrayInt(indiciesProp[0])
			if err != nil {
				return nil, err
			}

			tmp_i := 0
			tri_count := 0
			insertIdx := 0
			for poly := 0; poly < len(tmp); {
				poly++
				tri_count, tmp_i = getTriCountFromPoly(original_indices, tmp_i)
				for i := 0; i < tri_count; i++ {
					geom.Materials[insertIdx] = tmp[poly]
					insertIdx++
				}
			}
		} else {
			if mappingProp[0].value.String() != "AllSame" {
				return nil, errors.New("Mapping not supported")
			}
		}

		for _, elem := range element.children {
			if elem.id.String() != "LayerElementUV" {
				continue
			}
			uv_index := 0
			if len(elem.properties) > 0 {
				uv_index = int(elem.properties[0].value.toInt32())
			}
			if uv_index >= 0 && uv_index < geom.UVSMax() {
				tmp, tmp_indices, mapping, err := parseVertexDataVec2(elem, "UV", "UVIndex")
				if err != nil {
					return nil, err
				}
				if tmp != nil && len(tmp) > 0 {
					//uvs = [4]floatgeom.Point2{} //resize(tmp_indices.empty() ? tmp.size() : tmp_indices.size());
					geom.UVs[uv_index] = splatVec2(mapping, tmp, tmp_indices, original_indices)
					remapVec2(&geom.UVs[uv_index], to_old_indices)
				}
			}

		}

		layerTangentElems := findChildren(element, "LayerElementTangents")
		if len(layerTangentElems) > 0 {
			tans := findChildren(layerTangentElems[0], "Tangents")
			var tmp []floatgeom.Point3
			var tmp_indices []int
			var mapping VertexDataMapping
			var err error
			if len(tans) > 0 {
				tmp, tmp_indices, mapping, err = parseVertexDataVec3(layerTangentElems[0], "Tangents", "TangentsIndex")
			} else {
				tmp, tmp_indices, mapping, err = parseVertexDataVec3(layerTangentElems[0], "Tangent", "TangentIndex")
			}
			if err != nil {
				return nil, err
			}
			if tmp != nil && len(tmp) > 0 {
				geom.Tangents = splatVec3(mapping, tmp, tmp_indices, original_indices)
				remapVec3(&geom.Tangents, to_old_indices)
			}
		}

		layerColorElems := findChildren(element, "LayerElementColor")
		if len(layerColorElems) > 0 {
			tmp, tmp_indices, mapping, err := parseVertexDataVec4(layerColorElems[0], "Colors", "ColorIndex")
			if err != nil {
				return nil, err
			}
			if len(tmp) > 0 {
				geom.Colors = splatVec4(mapping, tmp, tmp_indices, original_indices)
				remapVec4(&geom.Colors, to_old_indices)
			}
		}

		layerNormalElems := findChildren(element, "LayerElementNormal")
		if len(layerNormalElems) > 0 {
			tmp, tmp_indices, mapping, err := parseVertexDataVec3(layerNormalElems[0], "Normals", "NormalsIndex")
			if err != nil {
				return nil, err
			}
			if len(tmp) > 0 {
				geom.Normals = splatVec3(mapping, tmp, tmp_indices, original_indices)
				remapVec3(&geom.Normals, to_old_indices)
			}
		}
	}

	// Todo: undo / redo some work above to not require redoing vertices

	geom.Vertices = vertices

	return geom, nil
}

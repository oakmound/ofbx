package ofbx

import (
	"errors"
	"fmt"
)

type VertexDataMapping int

const (
	BY_POLYGON_VERTEX = iota
	BY_POLYGON        = iota
	BY_VERTEX         = iota
)

const s_uvs_max = 4

type Geometry struct {
	Object
	skin *Skin

	vertices, normals, tangents []Vec3

	uvs                        [s_uvs_max][]Vec2
	colors                     []Vec4
	materials, to_old_vertices []int
	to_new_vertices            []Vertex
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
	return s_uvs_max
}

func (g *Geometry) getVertices() []Vec3 {
	return g.vertices
}
func (g *Geometry) getVertexCount() int {
	return len(g.vertices)
}

func (g *Geometry) getNormals() []Vec3 {
	return g.normals
}

func (g *Geometry) getUVs() []Vec2 {
	return g.getUVsIndex(0)
}

func (g *Geometry) getUVsIndex(index int) []Vec2 {
	if index < 0 || index > len(g.uvs) {
		return nil
	}
	return g.uvs[index]
}

func (g *Geometry) getColors() []Vec4 {
	return g.colors
}

func (g *Geometry) getTangents() []Vec3 {
	return g.tangents
}

func (g *Geometry) getSkin() *Skin {
	return g.skin
}

func (g *Geometry) getMaterials() []int {
	return g.materials
}

func (g *Geometry) triangulate(old_indices []int) []int {
	to_old := make([]int, 0)
	in_polygon_idx := 0
	for i := 0; i < len(old_indices); i++ {
		idx := old_indices[i]
		if idx < 0 {
			idx = -idx - 1
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

//From CPP

func (g *Geometry) getType() Type {
	return g.Type()
}

func parseGeometry(scene *Scene, element *Element) (*Geometry, error) {
	if element.first_property == nil {
		return nil, errors.New("Geometry invalid")
	}
	geom := NewGeometry(scene, element)

	vertices_element := findChild(element, "Vertices")
	if vertices_element == nil || vertices_element.first_property == nil {
		return nil, errors.New("Geometry Vertices Missing")
	}

	polys_element := findChild(element, "PolygonVertexIndex")
	if polys_element == nil || polys_element.first_property == nil {
		return nil, errors.New("Geometry Indicies missing")
	}
	fmt.Println("Geometry parsing arrays")
	vertices, err := parseDoubleVecDataVec3(vertices_element.first_property)
	if err != nil {
		return nil, err
	}
	fmt.Println("Parsing binary int array")
	original_indices, err := parseBinaryArrayInt(polys_element.first_property)
	if err != nil {
		return nil, err
	}

	to_old_indices := geom.triangulate(original_indices)
	geom.vertices = make([]Vec3, len(geom.to_old_vertices))

	for i, vIdx := range geom.to_old_vertices {
		v := vertices[vIdx]
		geom.vertices[i] = v
	}

	geom.to_new_vertices = make([]Vertex, len(geom.vertices))

	for i := 0; i < len(geom.to_old_vertices); i++ {
		old := geom.to_old_vertices[i]
		geom.to_new_vertices[old].add(i)
	}

	layer_material_element := findChild(element, "LayerElementMaterial")
	if layer_material_element != nil {
		mapping_element := findChild(layer_material_element, "MappingInformationType")
		reference_element := findChild(layer_material_element, "ReferenceInformationType")
		if mapping_element == nil || reference_element == nil {
			return nil, errors.New("Invalid LayerElementMaterial")
		}
		var err error
		tmp := make([]int, 0)

		if mapping_element.first_property.value.String() == "ByPolygon" &&
			reference_element.first_property.value.String() == "IndexToDirect" {
			geom.materials = make([]int, len(geom.vertices)/3)
			for i := 0; i < len(geom.vertices)/3; i++ {
				geom.materials[i] = -1
			}

			indices_element := findChild(layer_material_element, "Materials")
			if indices_element == nil || indices_element.first_property == nil {
				return nil, errors.New("Invalid LayerElementMaterial")
			}

			tmp, err = parseBinaryArrayInt(indices_element.first_property)
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
					geom.materials[insertIdx] = tmp[poly]
					insertIdx++
				}
			}
		} else {
			if mapping_element.first_property.value.String() != "AllSame" {
				return nil, errors.New("Mapping not supported")
			}
		}

		layer_uv_element := findChild(element, "LayerElementUV")
		for layer_uv_element != nil {
			uv_index := 0
			if layer_uv_element.first_property != nil {
				uv_index = int(layer_uv_element.first_property.getValue().toInt32())
			}
			if uv_index >= 0 && uv_index < geom.UVSMax() {
				tmp, tmp_indices, mapping, err := parseVertexDataVec2(layer_uv_element, "UV", "UVIndex")
				if err != nil {
					return nil, err
				}
				if tmp != nil && len(tmp) > 0 {
					//uvs = [4]Vec2{} //resize(tmp_indices.empty() ? tmp.size() : tmp_indices.size());
					geom.uvs[uv_index] = splatVec2(mapping, tmp, tmp_indices, original_indices)
					remapVec2(&geom.uvs[uv_index], to_old_indices)
				}
			}
			layer_uv_element = layer_uv_element.sibling
			for layer_uv_element != nil && layer_uv_element.id.String() != "LayerElementUV" {
				layer_uv_element = layer_uv_element.sibling
			}

		}

		layer_tangent_element := findChild(element, "LayerElementTangents")
		if layer_tangent_element != nil {
			tans := findChild(layer_tangent_element, "Tangents")
			var tmp []Vec3
			var tmp_indices []int
			var mapping VertexDataMapping
			var err error
			if tans != nil {
				tmp, tmp_indices, mapping, err = parseVertexDataVec3(layer_tangent_element, "Tangents", "TangentsIndex")
			} else {
				tmp, tmp_indices, mapping, err = parseVertexDataVec3(layer_tangent_element, "Tangent", "TangentIndex")
			}
			if err != nil {
				return nil, err
			}
			if tmp != nil && len(tmp) > 0 {
				geom.tangents = splatVec3(mapping, tmp, tmp_indices, original_indices)
				remapVec3(&geom.tangents, to_old_indices)
			}
		}

		layer_color_element := findChild(element, "LayerElementColor")
		if layer_color_element != nil {
			tmp, tmp_indices, mapping, err := parseVertexDataVec4(layer_color_element, "Colors", "ColorIndex")
			if err != nil {
				return nil, err
			}
			if len(tmp) > 0 {
				geom.colors = splatVec4(mapping, tmp, tmp_indices, original_indices)
				remapVec4(&geom.colors, to_old_indices)
			}
		}

		layer_normal_element := findChild(element, "LayerElementNormal")
		if layer_normal_element != nil {
			tmp, tmp_indices, mapping, err := parseVertexDataVec3(layer_normal_element, "Normals", "NormalsIndex")
			if err != nil {
				return nil, err
			}
			if len(tmp) > 0 {
				geom.normals = splatVec3(mapping, tmp, tmp_indices, original_indices)
				remapVec3(&geom.normals, to_old_indices)
			}
		}
	}
	return geom, nil
}

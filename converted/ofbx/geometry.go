package ofbx

import "errors"

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

	uvs                        [s_uvs_max]Vec2
	colors                     []Vec4
	materials, to_old_vertices []int
	to_new_vertices            []NewVector
}

//Hey its a linked list of indices!.....
type NewVertex struct {
	index int //should start as -1
	next  *NewVertex
}

func (nv *NewVertex) delete() {
	if next != nil {
		next.delete()
	}
}
func add(nv *NewVertex, index int) {
	return nv.add(index)
}
func (nv *NewVertex) add(index int) {
	if vtx.index == -1 {
		//TODO: change this cuz we aint implementing it this way. Really its checking if the newvertex exists...
		vtx.index = index
	} else if vtx.next != nil {
		add(*vtx.next, index)
	} else {
		vtx.next = NewVertex{-1, nil}
		vtx.next.index = index
	}
}

func NewGeometry(scene *Scene, element *IElement) *Geometry {

	g := NewObject(scene, element)
	g.to_old_vertices = make([]int)

	return nil
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

func (g *Geometry) getNormals() *Vec3 {
	return g.normals
}

func (g *Geometry) getUVs() *Vec2 { return g.getUVs(0) }
func (g *Geometry) getUVs(index int) *Vec2 {
	if index < 0 || index > len(g.uvs) {
		return nil
	}
	return g.uvs[index]
}

func (g *Geometry) getColors() *Vec4 {
	return g.colors
}

func (g *Geometry) getTangents() *Vec3 {
	return g.tangents
}

func (g *Geometry) getSkin() *Skin {
	return g.skin
}

func (g *Geometry) getMaterials() *int {
	return g.materials
}

func (g *Geometry) triangulate(old_indices []int) {
	to_old_indices := make([]int)
	in_polygon_idx := 0
	for i := 0; i < len(old_indices); {
		i++
		idx := old_indices[i]
		if idx < 0 {
			idx = -idx - 1
		}

		if in_polygon_idx <= 2 {
			geom.to_old_vertices = append(geom.to_old_vertices, idx)
			to_old = append(to_old, i)
		} else {
			geom.to_old_vertices = append(geom.to_old_vertices, old_indices[i-in_polygon_idx])
			to_old = append(to_old, i-in_polygon_idx)
			geom.to_old_vertices = append(geom.to_old_vertices, old_indices[i-1])
			to_old = append(to_old, i-1)
			geom.to_old_vertices = append(geom.to_old_vertices, idx)
			to_old = append(to_old, i)
		}
		in_polygon_idx++
		if old_indices[i] < 0 {
			in_polygon_idx = 0
		}
	}

}

//From CPP

func (g *Geometry) getType() {
	return g.Type()
}

func parseGeometry(scene *Scene, element *IElement) (*Object, Error) {
	if element.first_property == nil {
		return nil, errors.New("Geometry invalid")
	}
	geom := NewGeometry(scene, element)

	vertices_element := findChild(element, "Vertices")
	if !vertices_element || !vertices_element.first_property {
		return nil, errors.New("Geometry Vertices Missing")
	}

	polys_element := findChild(element, "PolygonVertexIndex")
	if !polys_element || !polys_element.first_property {
		return nil, errors.New("Geometry Indicies missing")
	}

	vertices = parseDoubleVecData(*vertices_element.first_property)
	original_indices = parseBinaryArray(*polys_element.first_property)

	geom.triangulate(original_indices)
	geom.vertices = make([]Vec3, len(geom.to_old_vertices))

	for i, vIdx := range geom.to_old_vertices {
		geom.vertices[i] = vertices[vIdx]
	}

	geom.to_new_vertices = make([]NewVertex, len(geom.vertices))

	for i := 0; i < len(geom.to_old_vertices); {
		i++
		old := to_old_vertices[i]
		geom.to_new_vertices[old].add(i)
	}

	layer_material_element := findChild(element, "LayerElementMaterial")
	if layer_material_element != nil {
		mapping_element := findChild(*layer_material_element, "MappingInformationType")
		reference_element := findChild(*layer_material_element, "ReferenceInformationType")
		if !mapping_element || !reference_element {
			return nil, errors.New("Invalid LayerElementMaterial")
		}
		tmp := make([]int)

		if mapping_element.first_property.value == "ByPolygon" && reference_element.first_property.value == "IndexToDirect" {
			geom.materials = make([]int, len(geom.vertices)/3)
			for i := 0; i < len(geom.vertices)/3; i++ {
				geom.materials[i] = -1
			}

			indices_element := findChild(*layer_material_element, "Materials")
			if !indices_element || !indices_element.first_property {
				return nil, errors.New("Invalid LayerElementMaterial")
			}

			tmp := parseBinaryArray(*indices_element.first_property) //int

			tmp_i := 0
			tri_count := 0
			insertIdx := 0
			for poly := 0; poly < len(tmp); {
				poly++
				tri_count, tmp_i = getTriCountFromPoly(original_indices, tmp_i)
				for i := 0; i < tri_count; {
					i++
					geom.materials[insertIdx] = tmp[poly]
					insertIdx++
				}
			}
		} else {
			if mapping_element.first_property.value != "AllSame" {
				return nil, errors.New("Mapping not supported")
			}
		}

		layer_uv_element := findChild(element, "LayerElementUV")
		for layer_uv_element != nil {
			uv_index := 0
			if layer_uv_element.first_property != nil {
				uv_index = layer_uv_element.first_property.getValue().toInt()
			}
			if uv_index >= 0 && uv_index < geom.UVSMax() {
				uvs := geom.uvs[uv_index]
				//tmp []Vec2 			//tmp_indices []int		//mapping VertexDataMapping
				tmp, tmp_indices, mapping := parseVertexData(*layer_uv_element, "UV", "UVIndex")

				if tmp != nil && len(tmp) > 0 {
					geom.uvs = make([]Vec2) //resize(tmp_indices.empty() ? tmp.size() : tmp_indices.size());
					splat(&uvs, mapping, tmp, tmp_indices, original_indices)
					remap(&uvs, to_old_indices)
				}
			}
			layer_uv_element = layer_uv_element.sibling
			for layer_uv_element && layer_uv_element.id != "LayerElementUV" {
				layer_uv_element = layer_uv_element.sibling
			}

		}

		layer_tangent_element := findChild(element, "LayerElementTangents")
		if layer_tangent_element != nil {
			tans := findChild(*layer_tangent_element, "Tangents")
			if len(tans > 0) {
				tmp, tmp_indices, mapping := parseVertexData(*layer_tangent_element, "Tangents", "TangentsIndex")
			} else {
				tmp, tmp_indices, mapping := parseVertexData(*layer_tangent_element, "Tangent", "TangentIndex")
			}
			if tmp != nil && len(tmp) > 0 {
				splat(&geom.tangents, mapping, tmp, tmp_indices, original_indices)
				remap(&geom.tangents, to_old_indices)
			}
		}

		layer_color_element := findChild(element, "LayerElementColor")
		if layer_color_element != nil {
			tmp, tmp_indices, mapping := parseVertexData(*layer_color_element, "Colors", "ColorIndex")
			if tmp != nil && len(tmp) > 0 {
				splat(&geom.colors, mapping, tmp, tmp_indices, original_indices)
				remap(&geom.colors, to_old_indices)
			}
		}

		layer_normal_element := findChild(element, "LayerElementNormal")
		if layer_normal_element != nil {
			tmp, tmp_indices, mapping := parseVertexData(*layer_normal_element, "Normals", "NormalsIndex")
			if tmp != nil && len(tmp) > 0 {
				splat(&geom.normals, mapping, tmp, tmp_indices, original_indices)
				remap(&geom.normals, to_old_indices)
			}
		}
	}

}

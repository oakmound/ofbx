package ofbx





type VertexDataMapping int 

const (
		BY_POLYGON_VERTEX = iota,
		BY_POLYGON = iota,
		BY_VERTEX = iota
)

const s_uvs_max = 4


type Geometry struct {
	Object

	skin *Skin

	vertices, normals, tangents []Vec3
	
	uvs [s_uvs_max]Vec2
	 colors []Vec4
	materials, to_old_vertices []int
	to_new_vertices []NewVector

	
}


//Hey its a linked list of indices!.....
type  NewVertex struct {
	index int //should start as -1
	next *NewVertex
}
func (nv *NewVertex) ~NewVertex() {
	if(next!=nil){
		next.~NewVertex()
	}
}
func add(nv *NewVertex, index int){
	return nv.add(index)
}
func (nv *NewVertex){
	if (vtx.index == -1){
		 //TODO: change this cuz we aint implementing it this way. Really its checking if the newvertex exists...
		vtx.index = index;
	}
	else if (vtx.next){
		add(*vtx.next, index);
	}else{
		vtx.next = new GeometryImpl::NewVertex;
		vtx.next.index = index;
	}
}




func NewGeometry(scene *Scene, element *IElement) *Geometry {
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
func (g *Geometry) getVertexCount() int{
	return len(g.vertices)
}

func (g *Geometry) getNormals() *Vec3 {
	return g.normals
}

func (g *Geometry) getUVs() *Vec2{return g.getUVs(0)}
func (g *Geometry) getUVs(index int) *Vec2 {
	if(index < 0 || index > len(g.uvs)){
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




//From CPP

func (g *Geometry) getType(){
	return g.Type()
}


func parseGeometry(scene *Scene, element *IElement) *Object, Error{
	if(element.first_property == nil){
		return nil, errors.New("Geometry invalid");
	}
	geom := NewGeometry(scene, element)

	vertices_element := findChild(element, "Vertices")
	if (!vertices_element || !vertices_element.first_property){
		return nil, errors.New("Geometry Vertices Missing")
	}

	polys_element := findChild(element, "PolygonVertexIndex")
	if (!polys_element || !polys_element.first_property){
		return nil, errors.New("Geometry Indicies missing")
	}

	geom.vertices = parseDoubleVecData(*vertices_element.first_property)
	geom.polys_element = parseBinaryArray(*polys_element.first_property)
	


}




static OptionalError<Object*> parseGeometry(const Scene& scene, const Element& element){
	

	//here down
	std::vector<Vec3> vertices;
	if (!parseDoubleVecData(*vertices_element.first_property, &vertices)) return Error("Failed to parse vertices");
	std::vector<int> original_indices;
	if (!parseBinaryArray(*polys_element.first_property, &original_indices)) return Error("Failed to parse indices");

	std::vector<int> to_old_indices;
	geom.triangulate(original_indices, &geom.to_old_vertices, &to_old_indices);
	geom.vertices.resize(geom.to_old_vertices.size());

	for (int i = 0, c = (int)geom.to_old_vertices.size(); i < c; ++i)
	{
		geom.vertices[i] = vertices[geom.to_old_vertices[i]];
	}

	geom.to_new_vertices.resize(vertices.size()); // some vertices can be unused, so this isn't necessarily the same size as to_old_vertices.
	const int* to_old_vertices = geom.to_old_vertices.empty() ? nullptr : &geom.to_old_vertices[0];
	for (int i = 0, c = (int)geom.to_old_vertices.size(); i < c; ++i)
	{
		int old = to_old_vertices[i];
		add(geom.to_new_vertices[old], i);
	}

	const Element* layer_material_element = findChild(element, "LayerElementMaterial");
	if (layer_material_element)
	{
		const Element* mapping_element = findChild(*layer_material_element, "MappingInformationType");
		const Element* reference_element = findChild(*layer_material_element, "ReferenceInformationType");

		std::vector<int> tmp;

		if (!mapping_element || !reference_element) return Error("Invalid LayerElementMaterial");

		if (mapping_element.first_property.value == "ByPolygon" &&
			reference_element.first_property.value == "IndexToDirect")
		{
			geom.materials.reserve(geom.vertices.size() / 3);
			for (int& i : geom.materials) i = -1;

			const Element* indices_element = findChild(*layer_material_element, "Materials");
			if (!indices_element || !indices_element.first_property) return Error("Invalid LayerElementMaterial");

			if (!parseBinaryArray(*indices_element.first_property, &tmp)) return Error("Failed to parse material indices");

			int tmp_i = 0;
			for (int poly = 0, c = (int)tmp.size(); poly < c; ++poly)
			{
				int tri_count = getTriCountFromPoly(original_indices, &tmp_i);
				for (int i = 0; i < tri_count; ++i)
				{
					geom.materials.push_back(tmp[poly]);
				}
			}
		}
		else
		{
			if (mapping_element.first_property.value != "AllSame") return Error("Mapping not supported");
		}
	}

	const Element* layer_uv_element = findChild(element, "LayerElementUV");
    while (layer_uv_element)
    {
        const int uv_index = layer_uv_element.first_property ? layer_uv_element.first_property.getValue().toInt() : 0;
        if (uv_index >= 0 && uv_index < Geometry::s_uvs_max)
        {
            std::vector<Vec2>& uvs = geom.uvs[uv_index];

            std::vector<Vec2> tmp;
            std::vector<int> tmp_indices;
            GeometryImpl::VertexDataMapping mapping;
            if (!parseVertexData(*layer_uv_element, "UV", "UVIndex", &tmp, &tmp_indices, &mapping)) return Error("Invalid UVs");
            if (!tmp.empty())
            {
                uvs.resize(tmp_indices.empty() ? tmp.size() : tmp_indices.size());
                splat(&uvs, mapping, tmp, tmp_indices, original_indices);
                remap(&uvs, to_old_indices);
            }
        }

        do
        {
            layer_uv_element = layer_uv_element.sibling;
        } while (layer_uv_element && layer_uv_element.id != "LayerElementUV");
    }

	const Element* layer_tangent_element = findChild(element, "LayerElementTangents");
	if (layer_tangent_element)
	{
		std::vector<Vec3> tmp;
		std::vector<int> tmp_indices;
		GeometryImpl::VertexDataMapping mapping;
		if (findChild(*layer_tangent_element, "Tangents"))
		{
			if (!parseVertexData(*layer_tangent_element, "Tangents", "TangentsIndex", &tmp, &tmp_indices, &mapping)) return Error("Invalid tangets");
		}
		else
		{
			if (!parseVertexData(*layer_tangent_element, "Tangent", "TangentIndex", &tmp, &tmp_indices, &mapping))  return Error("Invalid tangets");
		}
		if (!tmp.empty())
		{
			splat(&geom.tangents, mapping, tmp, tmp_indices, original_indices);
			remap(&geom.tangents, to_old_indices);
		}
	}

	const Element* layer_color_element = findChild(element, "LayerElementColor");
	if (layer_color_element)
	{
		std::vector<Vec4> tmp;
		std::vector<int> tmp_indices;
		GeometryImpl::VertexDataMapping mapping;
		if (!parseVertexData(*layer_color_element, "Colors", "ColorIndex", &tmp, &tmp_indices, &mapping)) return Error("Invalid colors");
		if (!tmp.empty())
		{
			splat(&geom.colors, mapping, tmp, tmp_indices, original_indices);
			remap(&geom.colors, to_old_indices);
		}
	}

	const Element* layer_normal_element = findChild(element, "LayerElementNormal");
	if (layer_normal_element)
	{
		std::vector<Vec3> tmp;
		std::vector<int> tmp_indices;
		GeometryImpl::VertexDataMapping mapping;
		if (!parseVertexData(*layer_normal_element, "Normals", "NormalsIndex", &tmp, &tmp_indices, &mapping)) return Error("Invalid normals");
		if (!tmp.empty())
		{
			splat(&geom.normals, mapping, tmp, tmp_indices, original_indices);
			remap(&geom.normals, to_old_indices);
		}
	}

	return geom.release();
}



struct GeometryImpl : Geometry
{
	
	void triangulate(const std::vector<int>& old_indices, std::vector<int>* indices, std::vector<int>* to_old)	{
		assert(indices);
		assert(to_old);

		auto getIdx = [&old_indices](int i) . int {
			int idx = old_indices[i];
			return idx < 0 ? -idx - 1 : idx;
		};

		int in_polygon_idx = 0;
		for (int i = 0; i < old_indices.size(); ++i)
		{
			int idx = getIdx(i);
			if (in_polygon_idx <= 2)
			{
				indices.push_back(idx);
				to_old.push_back(i);
			}
			else
			{
				indices.push_back(old_indices[i - in_polygon_idx]);
				to_old.push_back(i - in_polygon_idx);
				indices.push_back(old_indices[i - 1]);
				to_old.push_back(i - 1);
				indices.push_back(idx);
				to_old.push_back(i);
			}
			++in_polygon_idx;
			if (old_indices[i] < 0)
			{
				in_polygon_idx = 0;
			}
		}
	}
};


package ofbx

type Geometry struct {
	Object
}

func NewGeometry(scene *Scene, element *IElement) *Geometry {
	return nil
}

func (g *Geometry) Type() Type {
	return GEOMETRY
}

func (g *Geometry) UVSMax() int {
	return 4
}

func (g *Geometry) getVertices() []Vec3 {
	return nil
}

func (g *Geometry) getNormals() *Vec3 {
	return nil
}

func (g *Geometry) getUVs(index int) *Vec2 {
	return nil
}

func (g *Geometry) getColors() *Vec4 {
	return nil
}

func (g *Geometry) getTangents() *Vec3 {
	return nil
}

func (g *Geometry) getSkin() *Skin {
	return nil
}

func (g *Geometry) getMaterials() *int {
	return nil
}


//From CPP

static void add(GeometryImpl::NewVertex& vtx, int index)
{
	if (vtx.index == -1)
	{
		vtx.index = index;
	}
	else if (vtx.next)
	{
		add(*vtx.next, index);
	}
	else
	{
		vtx.next = new GeometryImpl::NewVertex;
		vtx.next.index = index;
	}
}


static OptionalError<Object*> parseGeometry(const Scene& scene, const Element& element)
{
	assert(element.first_property);

	const Element* vertices_element = findChild(element, "Vertices");
	if (!vertices_element || !vertices_element.first_property) return Error("Vertices missing");

	const Element* polys_element = findChild(element, "PolygonVertexIndex");
	if (!polys_element || !polys_element.first_property) return Error("Indices missing");

	std::unique_ptr<GeometryImpl> geom = std::make_unique<GeometryImpl>(scene, element);

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

Geometry::Geometry(const Scene& _scene, const IElement& _element)
	: Object(_scene, _element)
{
}


struct GeometryImpl : Geometry
{
	enum VertexDataMapping
	{
		BY_POLYGON_VERTEX,
		BY_POLYGON,
		BY_VERTEX
	};

	struct NewVertex
	{
		~NewVertex() { delete next; }

		int index = -1;
		NewVertex* next = nullptr;
	};

	std::vector<Vec3> vertices;
	std::vector<Vec3> normals;
	std::vector<Vec2> uvs[s_uvs_max];
	std::vector<Vec4> colors;
	std::vector<Vec3> tangents;
	std::vector<int> materials;

	const Skin* skin = nullptr;

	std::vector<int> to_old_vertices;
	std::vector<NewVertex> to_new_vertices;

	GeometryImpl(const Scene& _scene, const IElement& _element)
		: Geometry(_scene, _element)
	{
	}


	Type getType() const override { return Type::GEOMETRY; }
	int getVertexCount() const override { return (int)vertices.size(); }
	const Vec3* getVertices() const override { return &vertices[0]; }
	const Vec3* getNormals() const override { return normals.empty() ? nullptr : &normals[0]; }
	const Vec2* getUVs(int index = 0) const override { return index < 0 || index >= s_uvs_max || uvs[index].empty() ? nullptr : &uvs[index][0]; }
	const Vec4* getColors() const override { return colors.empty() ? nullptr : &colors[0]; }
	const Vec3* getTangents() const override { return tangents.empty() ? nullptr : &tangents[0]; }
	const Skin* getSkin() const override { return skin; }
	const int* getMaterials() const override { return materials.empty() ? nullptr : &materials[0]; }


	void triangulate(const std::vector<int>& old_indices, std::vector<int>* indices, std::vector<int>* to_old)
	{
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


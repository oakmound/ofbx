package ofbx

type Cluster struct {
	Object
}

func NewCluster(scene *Scene, element *IElement) *Cluster {
	return nil
}

func (c *Cluster) Type() Type {
	return CLUSTER
}

func (c *Cluster) getIndices() []int {
	return nil
}

func (c *Cluster) getWeights() *float64 {
	return nil
}

func (c *Cluster) getTransformMatrix() Matrix {
	return Matrix{}
}

func (c *Cluster) getTransformLinkMatrix() Matrix {
	return Matrix{}
}

func (c *Cluster) getLink() *Object {
	return nil
}


//from cpp

struct ClusterImpl : Cluster
{
	ClusterImpl(const Scene& _scene, const IElement& _element)
		: Cluster(_scene, _element)
	{
	}

	const int* getIndices() const override { return &indices[0]; }
	int getIndicesCount() const override { return (int)indices.size(); }
	const double* getWeights() const override { return &weights[0]; }
	int getWeightsCount() const override { return (int)weights.size(); }
	Matrix getTransformMatrix() const override { return transform_matrix; }
	Matrix getTransformLinkMatrix() const override { return transform_link_matrix; }
	Object* getLink() const override { return link; }


	bool postprocess()
	{
		assert(skin);

		GeometryImpl* geom = (GeometryImpl*)skin.resolveObjectLinkReverse(Object::Type::GEOMETRY);
		if (!geom) return false;

		std::vector<int> old_indices;
		const Element* indexes = findChild((const Element&)element, "Indexes");
		if (indexes && indexes.first_property)
		{
			if (!parseBinaryArray(*indexes.first_property, &old_indices)) return false;
		}

		std::vector<double> old_weights;
		const Element* weights_el = findChild((const Element&)element, "Weights");
		if (weights_el && weights_el.first_property)
		{
			if (!parseBinaryArray(*weights_el.first_property, &old_weights)) return false;
		}

		if (old_indices.size() != old_weights.size()) return false;

		indices.reserve(old_indices.size());
		weights.reserve(old_indices.size());
		int* ir = old_indices.empty() ? nullptr : &old_indices[0];
		double* wr = old_weights.empty() ? nullptr : &old_weights[0];
		for (int i = 0, c = (int)old_indices.size(); i < c; ++i)
		{
			int old_idx = ir[i];
			double w = wr[i];
			GeometryImpl::NewVertex* n = &geom.to_new_vertices[old_idx];
			if (n.index == -1) continue; // skip vertices which aren't indexed.
			while (n)
			{
				indices.push_back(n.index);
				weights.push_back(w);
				n = n.next;
			}
		}

		return true;
	}


	Object* link = nullptr;
	Skin* skin = nullptr;
	std::vector<int> indices;
	std::vector<double> weights;
	Matrix transform_matrix;
	Matrix transform_link_matrix;
	Type getType() const override { return Type::CLUSTER; }
};

Cluster::Cluster(const Scene& _scene, const IElement& _element)
	: Object(_scene, _element)
{
}


static OptionalError<Object*> parseCluster(const Scene& scene, const Element& element)
{
	std::unique_ptr<ClusterImpl> obj = std::make_unique<ClusterImpl>(scene, element);

	const Element* transform_link = findChild(element, "TransformLink");
	if (transform_link && transform_link.first_property)
	{
		if (!parseArrayRaw(
				*transform_link.first_property, &obj.transform_link_matrix, sizeof(obj.transform_link_matrix)))
		{
			return Error("Failed to parse TransformLink");
		}
	}
	const Element* transform = findChild(element, "Transform");
	if (transform && transform.first_property)
	{
		if (!parseArrayRaw(*transform.first_property, &obj.transform_matrix, sizeof(obj.transform_matrix)))
		{
			return Error("Failed to parse Transform");

		}
	}

	return obj.release();
}
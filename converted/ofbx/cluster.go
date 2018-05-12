package ofbx

type Cluster struct {
	Object

	 link  *Object
	 skin *Skin
	 indices []int
	 weights []float64
	transform_matrix Matrix
	transform_link_matrix Matrix
}

func NewCluster(scene *Scene, element *IElement) *Cluster {




	c := NewObject(scene, element)
	c.clusterPostProcess(scene, element)
	c.link = &Object{}
	c.skin := &Skin{}
	c.transform_link_matrix = NewMatrix(scene, element) 
	c.transform_matrix= NewMatrix(scene, element)

	return c
}

func (c *Cluster) Type() Type {
	return CLUSTER
}

//TODO: should this really just be a link to the first index? as  pointer?
//const int* getIndices() const override { return &indices[0]; }
func (c *Cluster) getIndices() []int {
	return nil
	
}
func (c *Cluster) getIndicesCount() int{
	return len(getIndices)
}

func (c *Cluster) getWeights() *float64 {
	return c.weights
}
func (c *Cluster) getWeightsCount() int{
	return len(c.weights)
}

func (c *Cluster) getTransformMatrix() Matrix {
	return c.transform_matrix
}

func (c *Cluster) getTransformLinkMatrix() Matrix {
	return c.transform_link_matrix
}

func (c *Cluster) getLink() *Object {
	return c.link
}


// clusterPostProcess adds the additional fields that clusters have over just object fields.
// In this case its setting up indicies and weights
func (c *Cluster) clusterPostProcess(scene, element){
	 geom := (*GeometryImpl).skin.resolveObjectLinkReverse(GEOMETRY)
	 if (geom==nil){
		return false
  	 }
	   old_indices := []int{}
	   indexes := findChild(element, "Indexes");
	   if (indexes != nil && indexes.first_property)   {
		   if (!parseBinaryArray(*indexes.first_property, &old_indices)){ 
			   return false
			}
	   }
	   old_weights := []float64
	   weights_el := findChild(element, "Weights");
   		if (weights_el != nil && weights_el.first_property)	{
			if (!parseBinaryArray(*weights_el.first_property, &old_weights)){
				 return false
			}
		}

		iLen := len(old_indices)
		if (iLen != len(old_weights)){
			return false
		}

		c.weights := make([]float64)
		c.indices := make([]int)


		for(i := 0; i < iLen; i++){
			n := &geom.to_new_vertices[old_idx] //was a geometryimpl NewVertex
			if (n.index == -1) continue; // skip vertices which aren't indexed.
			for (n != nil)
			{
				c.indices[i]
				c.indices = c.indices.append(n.index)
				c.weights = c.weights.append(w);
				n = n.next;
			}
		}
		return true

}


func parseCluster(scene *Scene, element *IElement) *Object, Error{
	cluster := NewCLuster(scene, element)
	cluster.transform_link = findChild(element, "TransformLink");
	if (transform_link !=nil && transform_link.first_property)
	{
		if (!parseArrayRaw(
				*transform_link.first_property, &obj.transform_link_matrix, sizeof(obj.transform_link_matrix)))
		{
			return nil, errors.New("Failed to parse TransformLink");
		}
	}
	cluster.transform = findChild(element, "Transform");
	if (transform!=nil && transform.first_property)
	{
		if (!parseArrayRaw(*transform.first_property, &obj.transform_matrix, sizeof(obj.transform_matrix)))
		{
			return nil, errors.New("Failed to parse Transform");

		}
	}	
	return cluster, nil
}


//from cpp

func (c *Cluster) getType() Type{
	return c.Type()
}





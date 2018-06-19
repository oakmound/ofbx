package ofbx

import "github.com/pkg/errors"

type Cluster struct {
	Object

	link                  Obj
	skin                  *Skin
	indices               []int
	weights               []float64
	transform_matrix      Matrix
	transform_link_matrix Matrix
}

func NewCluster(scene *Scene, element *Element) *Cluster {
	c := Cluster{}
	c.Object = *NewObject(scene, element)
	return &c
}

func (c *Cluster) Type() Type {
	return CLUSTER
}

func (c *Cluster) getIndices() []int {
	return c.indices

}
func (c *Cluster) getIndicesCount() int {
	return len(c.indices)
}

func (c *Cluster) getWeights() []float64 {
	return c.weights
}
func (c *Cluster) getWeightsCount() int {
	return len(c.weights)
}

func (c *Cluster) getTransformMatrix() Matrix {
	return c.transform_matrix
}

func (c *Cluster) getTransformLinkMatrix() Matrix {
	return c.transform_link_matrix
}

func (c *Cluster) getLink() Obj {
	return c.link
}

// postProcess adds the additional fields that clusters have over just object fields.
// In this case its setting up indicies and weights
func (c *Cluster) postProcess() bool {
	element := c.Element()
	geom, ok := resolveObjectLinkReverse(c.skin, GEOMETRY).(*Geometry)
	if !ok {
		return false
	}
	var old_indices []int
	var err error
	indexes := findChild(element, "Indexes")
	if indexes != nil && indexes.first_property != nil {
		if old_indices, err = parseBinaryArrayInt(indexes.first_property); err != nil {
			return false
		}
	}
	var old_weights []float64
	weights_el := findChild(element, "Weights")
	if weights_el != nil && weights_el.first_property != nil {
		if old_weights, err = parseBinaryArrayFloat64(weights_el.first_property); err != nil {
			return false
		}
	}

	iLen := len(old_indices)
	if iLen != len(old_weights) {
		return false
	}

	c.weights = make([]float64, 0, iLen)
	c.indices = make([]int, 0, iLen)

	for i := 0; i < iLen; i++ {
		n := &geom.to_new_vertices[old_indices[i]] //was a geometryimpl NewVertex
		if n.index == -1 {
			continue // skip vertices which aren't indexed.
		}
		for n != nil {
			c.indices = append(c.indices, n.index)
			c.weights = append(c.weights, old_weights[i])
			n = n.next
		}
	}
	return true

}

func parseCluster(scene *Scene, element *Element) (*Cluster, error) {
	obj := NewCluster(scene, element)
	transform_link := findChild(element, "TransformLink")
	if transform_link != nil && transform_link.first_property != nil {
		mx, err := parseArrayRawFloat64(transform_link.first_property)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse TransformLink")
		}
		obj.transform_link_matrix, err = matrixFromSlice(mx)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse TransformLink")
		}
	}
	transform := findChild(element, "Transform")
	if transform != nil && transform.first_property != nil {
		mx, err := parseArrayRawFloat64(transform.first_property)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse TransformLink")
		}
		obj.transform_matrix, err = matrixFromSlice(mx)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse TransformLink")
		}
	}
	return obj, nil
}

func (c *Cluster) getType() Type {
	return c.Type()
}

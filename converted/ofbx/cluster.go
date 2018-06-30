package ofbx

import (
	"fmt"

	"github.com/pkg/errors"
)

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
	prop := findChildProperty(element, "Indexes")
	if prop != nil {
		if old_indices, err = parseBinaryArrayInt(prop[0]); err != nil {
			return false
		}
	}
	var old_weights []float64
	prop = findChildProperty(element, "Weights")
	if prop != nil {
		if old_weights, err = parseBinaryArrayFloat64(prop[0]); err != nil {
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

	prop := findChildProperty(element, "TransformLink")
	if prop != nil {
		mx, err := parseArrayRawFloat64(prop[0])
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse TransformLink")
		}
		obj.transform_link_matrix, err = matrixFromSlice(mx)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse TransformLink")
		}
	}
	prop = findChildProperty(element, "Transform")
	if prop != nil {
		mx, err := parseArrayRawFloat64(prop[0])
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

func (c *Cluster) String() string {
	s := "Cluster: " + c.Object.String()
	s += " link=" + c.link.String()
	s += " skin=" + c.skin.String()
	s += "indicies= " + fmt.Sprintf("%d", c.indices)
	s += " weights=" + fmt.Sprintf("%d", c.weights)
	s += " transform_matrix=" + fmt.Sprint(c.transform_matrix)
	s += " transform_link_matrix=" + fmt.Sprint(c.transform_link_matrix)
	return s
}

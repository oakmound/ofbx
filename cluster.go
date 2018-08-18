package ofbx

import (
	"fmt"

	"github.com/pkg/errors"
)

// Cluster is an entity which acts on a subset of a geometry's control points
// For each control point that the cluster acts on, the intensity of the cluster's action is modulated by a weight. The link mode (ELinkMode) specifies how the weights are taken into account.
type Cluster struct {
	Object

	Link          Obj
	Skin          *Skin
	Indices       []int
	Weights       []float64
	Transform     Matrix
	TransformLink Matrix
}

// NewCluster creates a new empty Cluster object
func NewCluster(scene *Scene, element *Element) *Cluster {
	c := Cluster{}
	c.Object = *NewObject(scene, element)
	return &c
}

// String pretty formats the Cluster for printing
func (c *Cluster) String() string {
	return c.stringPrefix("")
}

// stringPrefix pretty formats the Cluster while respecting the given prefix indentation
func (c *Cluster) stringPrefix(prefix string) string {
	s := prefix + "Cluster:" + "\n"
	s += c.Object.stringPrefix(prefix + "\t")
	s += prefix + "link:" + "\n" + c.Link.stringPrefix(prefix+"\t")
	s += prefix + "indices:" + fmt.Sprintf("%v", c.Indices) + "," + "\n"
	s += prefix + "weights:" + fmt.Sprintf("%v", c.Weights) + "," + "\n"
	s += prefix + "transform_matrix:" + fmt.Sprintf("%v", c.Transform) + "," + "\n"
	s += prefix + "transform_link_matrix:" + fmt.Sprintf("%v", c.TransformLink) + "," + "\n"
	return s
}

// Type returns CLUSTER
func (c *Cluster) Type() Type {
	return CLUSTER
}

// postProcess adds the additional fields that clusters have over just object fields.
// In this case its setting up indicies and weights
func (c *Cluster) postProcess() bool {
	element := c.Element()
	geom, ok := resolveObjectLinkReverse(c.Skin, GEOMETRY).(*Geometry)
	if !ok {
		return false
	}
	var oldIndices []int
	var err error
	prop := findChildProperty(element, "Indexes")
	if prop != nil {
		if oldIndices, err = parseBinaryArrayInt(prop[0]); err != nil {
			return false
		}
	}
	var oldWeights []float64
	prop = findChildProperty(element, "Weights")
	if prop != nil {
		if oldWeights, err = parseBinaryArrayFloat64(prop[0]); err != nil {
			return false
		}
	}

	iLen := len(oldIndices)
	if iLen != len(oldWeights) {
		return false
	}

	c.Weights = make([]float64, 0, iLen)
	c.Indices = make([]int, 0, iLen)

	for i := 0; i < iLen; i++ {
		n := &geom.newVerts[oldIndices[i]] //was a geometryimpl NewVertex
		if n.index == -1 {
			continue // skip vertices which aren't indexed.
		}
		for n != nil {
			c.Indices = append(c.Indices, n.index)
			c.Weights = append(c.Weights, oldWeights[i])
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
		obj.TransformLink, err = matrixFromSlice(mx)
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
		obj.Transform, err = matrixFromSlice(mx)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse TransformLink")
		}
	}
	return obj, nil
}

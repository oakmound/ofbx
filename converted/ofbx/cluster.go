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

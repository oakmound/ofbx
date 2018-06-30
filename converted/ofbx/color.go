package ofbx

import "fmt"

type Color struct {
	r, g, b float32
}

func (c *Color) String() string {
	return "Color: " + "r" + fmt.Sprintf("%e", c.r) +
		"g" + fmt.Sprintf("%e", c.g) +
		"b" + fmt.Sprintf("%e", c.b)
}

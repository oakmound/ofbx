package ofbx

import "fmt"

type Color struct {
	r, g, b float32
}

func (c *Color) String() string {
	return "Color: " + "r" + fmt.Sprintf("%d", c.r) +
		"g" + fmt.Sprintf("%d", c.g) +
		"b" + fmt.Sprintf("%d", c.b)
}

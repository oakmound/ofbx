package ofbx

import "fmt"

type Color struct {
	r, g, b float32
}

func (c *Color) String() string {
	return "R=" + fmt.Sprintf("%f", c.r) +
		" G=" + fmt.Sprintf("%f", c.g) +
		" B=" + fmt.Sprintf("%f", c.b)
}

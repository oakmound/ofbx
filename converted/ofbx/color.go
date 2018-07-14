package ofbx

import "fmt"

type Color struct {
	R, G, B float32
}

func (c *Color) String() string {
	return "R=" + fmt.Sprintf("%f", c.R) +
		" G=" + fmt.Sprintf("%f", c.G) +
		" B=" + fmt.Sprintf("%f", c.B)
}

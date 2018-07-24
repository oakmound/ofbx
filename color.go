package ofbx

import "fmt"

// Color is a set of floats RGB
type Color struct {
	R, G, B float32
}

func (c *Color) String() string {
	return "R=" + fmt.Sprintf("%f", c.R) +
		" G=" + fmt.Sprintf("%f", c.G) +
		" B=" + fmt.Sprintf("%f", c.B)
}

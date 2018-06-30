package ofbx

import (
	"fmt"
)

type TakeInfo struct {
	name                *DataView
	filename            *DataView
	local_time_from     float64
	local_time_to       float64
	reference_time_from float64
	reference_time_to   float64
}

func (t *TakeInfo) String() string {
	s := "TakeInfo: " + t.name.String()
	s += "," + t.filename.String()
	s += "times=" + fmt.Sprintf("%f,%f,%f,%f",
		t.local_time_from,
		t.local_time_to,
		t.reference_time_from,
		t.reference_time_to)
	return s + "\n"
}

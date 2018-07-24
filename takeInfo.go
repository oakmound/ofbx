package ofbx

import (
	"fmt"
)

// TakeInfo is a set of data for a time data set
type TakeInfo struct {
	name          *DataView
	filename      *DataView
	localTimeFrom float64
	localTimeTo   float64
	refTimeFrom   float64
	refTimeTo     float64
}

func (t *TakeInfo) String() string {
	s := "TakeInfo: " + t.name.String()
	s += "," + t.filename.String()
	s += ", times=" + fmt.Sprintf("%f,%f,%f,%f",
		t.localTimeFrom,
		t.localTimeTo,
		t.refTimeFrom,
		t.refTimeTo)
	return s + "\n"
}

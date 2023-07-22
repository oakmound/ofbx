package ofbx

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"time"
)

func (s *Scene) FBXString() string {
	w := &bytes.Buffer{}
	s.WriteText(w)
	return w.String()
}

func (s *Scene) WriteText(w io.Writer) {
	bf := bufio.NewWriter(w)
	bf.WriteString("; FBX 6.1.0 project file\n")
	bf.WriteString("; Created by oakmound/ofbx\n")
	bf.WriteString("; --------------------------\n")
	bf.WriteString("\n")
	writeFBXHeader(bf)
	bf.WriteString("\n")

}

const fbxTimeFormat = "2006-01-02 15:04:05:000"

func writeFBXHeader(bf *bufio.Writer) {
	now := time.Now()

	bf.WriteString("FBXHeaderExtension:  {\n")
	bf.WriteString("\tFBXHeaderVersion: 1003\n")
	bf.WriteString("\tFBXVersion: 6100\n")
	bf.WriteString("\tCreationTimeStamp:  {\n")
	bf.WriteString("\t\tVersion: 1000\n")
	bf.WriteString("\t\tYear: " + strconv.Itoa(now.Year()) + "\n")
	bf.WriteString("\t\tMonth: " + strconv.Itoa(int(now.Month())) + "\n")
	bf.WriteString("\t\tDay: " + strconv.Itoa(now.Day()) + "\n")
	bf.WriteString("\t\tHour: " + strconv.Itoa(now.Hour()) + "\n")
	bf.WriteString("\t\tMinute: " + strconv.Itoa(now.Minute()) + "\n")
	bf.WriteString("\t\tSecond: " + strconv.Itoa(now.Second()) + "\n")
	bf.WriteString("\t\tMillisecond: " + strconv.Itoa(now.Nanosecond()/1000) + "\n")
	bf.WriteString("\t}\n")
	bf.WriteString("\tCreator: \"oakmound/ofbx version 0.2.0\"\n")
	bf.WriteString("\tOtherFlags:  {\n")
	bf.WriteString("\t\tFlagPLE: 0\n")
	bf.WriteString("\t}\n")
	bf.WriteString("}\n")
	bf.WriteString("CreationTime:" + now.Format(fbxTimeFormat) + "\n")
	bf.WriteString("Creator: \"oakmound/ofbx version 0.2.0\"\n")

}

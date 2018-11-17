package ofbx

import (
	"bytes"
	"io"
)

func IsBinary(r io.Reader) bool {
	magic := append([]byte("Kaydara FBX Binary  "), 0)
	header := make([]byte, len(magic))
	n, err := r.Read(header)
	if n != len(header) {
		return false
	}
	if err != nil {
		return false
	}
	return bytes.Equal(magic, header)
}

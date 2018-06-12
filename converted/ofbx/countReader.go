package ofbx

import "io"

type CountReader struct {
	io.Reader
	ReadSoFar int
}

func (c *CountReader) Read(p []byte) (n int, err error) {
	n, err = c.Reader.Read(p)
	c.ReadSoFar += n
	return n, err
}

func NewCountReader(r io.Reader) *CountReader {
	return &CountReader{
		r, 0,
	}
}

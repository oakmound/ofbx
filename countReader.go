package ofbx

import "io"

// CountReader an io reader that knows how many bytes it has currently read
type CountReader struct {
	io.Reader
	ReadSoFar int
}

// Read reads in some number of bytes and returns the result from the underlying io.Reader
func (c *CountReader) Read(p []byte) (n int, err error) {
	n, err = c.Reader.Read(p)
	c.ReadSoFar += n
	return n, err
}

// NewCountReader creates a new wrapper around an io.Reader with a count of 0
func NewCountReader(r io.Reader) *CountReader {
	return &CountReader{
		r, 0,
	}
}

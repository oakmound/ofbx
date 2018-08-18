package ofbx

import "fmt"

type stringPrefixer interface {
	fmt.Stringer
	stringPrefix(string) string
}

package ofbx

import (
	"os"
	"testing"
)

func TestIsBinary(t *testing.T) {
	// Todo: we don't commit these test files, we should
	// have committable test files
	type testCase struct {
		file   string
		binary bool
	}
	testCases := []testCase{
		{"flex.fbx", true},
	}
	for _, tc := range testCases {
		r, err := os.Open(tc.file)
		if err != nil {
			t.Fatal("got not nil error")
		}
		if tc.binary != IsBinary(r) {
			t.Fatal("isBinary mismatch")
		}
	}
}

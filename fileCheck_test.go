package ofbx

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
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
		require.Nil(t, err)
		require.Equal(t, tc.binary, IsBinary(r))
	}
}

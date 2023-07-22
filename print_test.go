package ofbx_test

import (
	"log"
	"os"
	"testing"

	"github.com/oakmound/ofbx"
	"github.com/oakmound/ofbx/threefbx"
)

func TestPrintScene(t *testing.T) {
	f, err := os.Open("flex.FBX")
	if err != nil {
		log.Fatal(err)
	}
	_, err = ofbx.Load(f)
	if err != nil {
		log.Fatal(err)
	}
}

func TestThreePrint(t *testing.T) {
	f, err := os.Open("character3.fbx")
	if err != nil {
		log.Fatal(err)
	}
	_, err = threefbx.Load(f, "")
	if err != nil {
		log.Fatal(err)
	}
}

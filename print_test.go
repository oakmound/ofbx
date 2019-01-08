package ofbx_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/oakmound/ofbx"
	"github.com/oakmound/ofbx/threefbx"
)

func TestPrintScene(t *testing.T) {
	f, err := os.Open("flex.FBX")
	if err != nil {
		log.Fatal(err)
	}
	scene, err := ofbx.Load(f)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(scene)
	//fmt.Println(scene.Geometries())
	log.Fatal("failing for print")
}

func TestThreePrint(t *testing.T) {
	f, err := os.Open("character3.fbx")
	if err != nil {
		log.Fatal(err)
	}
	scene, err := threefbx.Load(f, "")
	if err != nil {
		log.Fatal(err)
	}
	spew.Dump(scene)
	log.Fatal("failing for print")
}

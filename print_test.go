package ofbx_test

import (
	"fmt"
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
	scene, err := ofbx.Load(f)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(scene)
	//fmt.Println(scene.Geometries())
	log.Fatal("failing for print")
}

func TestThreePrint(t *testing.T) {
	f, err := os.Open("flex.FBX")
	if err != nil {
		log.Fatal(err)
	}
	scene, err := threefbx.Load(f, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(scene)
}

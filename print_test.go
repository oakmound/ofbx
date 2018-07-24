package ofbx

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestPrintScene(t *testing.T) {
	f, err := os.Open("cuberotateblend.FBX")
	if err != nil {
		log.Fatal(err)
	}
	scene, err := Load(f)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(scene)
	//fmt.Println(scene.Geometries())
}

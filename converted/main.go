package main

import (
	"fmt"
	"log"
	"os"

	"github.com/oakmound/openfbx/converted/ofbx"
)

func main() {
	f, err := os.Open("cuberotateblend.FBX")
	if err != nil {
		log.Fatal(err)
	}
	scene, err := ofbx.Load(f)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(scene)
}

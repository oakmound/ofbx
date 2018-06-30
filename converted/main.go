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
	fmt.Print("\n\n\n~~~\n\n\n")
	fmt.Println(scene)
}

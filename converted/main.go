package main

import (
	"fmt"
	"log"
	"os"

	"github.com/oakmound/openfbx/converted/ofbx"
)

func main() {
	fmt.Println("Starting main")
	f, err := os.Open("character3.FBX")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("File opened")
	scene, err := ofbx.Load(f)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(scene)
}

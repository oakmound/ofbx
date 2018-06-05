package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/oakmound/openfbx/converted/ofbx"
)

func main() {
	fmt.Println("Starting main")
	f, err := os.Open("a.FBX")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("File opened")
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("File read")
	scene, err := ofbx.Load(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(scene)
}

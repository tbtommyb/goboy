package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/tbtommyb/goboy/pkg/cpu"
)

func main() {
	romPtr := flag.String("file", "input.rom", "ROM path to read from")
	flag.Parse()
	data, err := ioutil.ReadFile(*romPtr)
	if err != nil {
		log.Fatalf("File reading error %#v", err)
	}

	cpu := cpu.Init()
	cpu.LoadROM(data)
	cpu.Run()
}

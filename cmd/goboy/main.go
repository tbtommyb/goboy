package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hajimehoshi/ebiten"
	"github.com/tbtommyb/goboy/pkg/cpu"
)

var CYCLES_PER_FRAME = 70224

func main() {
	romPtr := flag.String("file", "input.rom", "ROM path to read from")
	flag.Parse()
	data, err := ioutil.ReadFile(*romPtr)
	if err != nil {
		log.Fatalf("File reading error %#v", err)
	}

	cpu := cpu.Init()
	cpu.LoadROM(data)

	f := func(screen *ebiten.Image) error {
		for i := 0; i < CYCLES_PER_FRAME; i++ {
			cpu.Step()
		}
		screen.ReplacePixels(cpu.Display.Pixels())
		ebiten.SetWindowTitle("Goboy")
		return nil
	}

	ebiten.SetRunnableInBackground(true)
	err = ebiten.Run(f, int(160), int(144), 2, "Goboy")
	if err != nil {
		fmt.Sprintf("Exited main() with error: %s", err)
	}
	return
}

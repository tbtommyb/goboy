package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hajimehoshi/ebiten"
	"github.com/tbtommyb/goboy/pkg/constants"
	"github.com/tbtommyb/goboy/pkg/cpu"
)

var CYCLES_PER_FRAME = 70224

func main() {
	biosPtr := flag.String("bios", "bios.rom", "BIOS path to read from")
	romPtr := flag.String("file", "input.rom", "ROM path to read from")
	flag.Parse()
	data, err := ioutil.ReadFile(*romPtr)
	if err != nil {
		log.Fatalf("File reading error %#v", err)
	}
	bios, err := ioutil.ReadFile(*biosPtr)
	if err != nil {
		log.Fatalf("File reading error %#v", err)
	}

	cpu := cpu.Init()
	cpu.LoadBIOS(data)
	cpu.LoadBIOS(bios)

	f := func(screen *ebiten.Image) error {
		for i := 0; i < CYCLES_PER_FRAME; i++ {
			cycles := cpu.Step()
			cpu.Display.Update(cycles)
		}
		screen.ReplacePixels(cpu.Display.Pixels())
		return nil
	}

	ebiten.SetWindowTitle("Goboy")
	ebiten.SetRunnableInBackground(true)
	err = ebiten.Run(f, constants.ScreenWidth, constants.ScreenHeight, constants.ScreenScaling, "Goboy")
	if err != nil {
		fmt.Sprintf("Exited main() with error: %s", err)
	}
	return
}

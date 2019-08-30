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

var GameboyClockSpeed = 4194300
var EbitenFPS = 60
var CyclesPerFrame = GameboyClockSpeed / EbitenFPS

func main() {
	var loadBIOS, testMode bool
	var bios, test, rom []byte
	var err error

	biosPtr := flag.String("bios", "", "BIOS path to read from")
	romPtr := flag.String("rom", "", "ROM path to read from")
	testPtr := flag.String("test", "", "test file to run")
	flag.Parse()

	if *testPtr != "" {
		test, err = ioutil.ReadFile(*testPtr)
		if err != nil {
			log.Fatalf("Error reading test ROM %#v", err)
		}
		testMode = true
	}
	if *biosPtr != "" {
		bios, err = ioutil.ReadFile(*biosPtr)
		if err != nil {
			log.Fatalf("Error reading BIOS ROM %#v", err)
		}
		loadBIOS = true
	}
	if *romPtr != "" {
		rom, err = ioutil.ReadFile(*romPtr)
		if err != nil {
			log.Fatalf("Error reading ROM %#v", err)
		}
	}

	cpu := cpu.Init(loadBIOS)
	if testMode {
		cpu.LoadROM(test)
	} else {
		cpu.LoadROM(rom)
		if loadBIOS {
			cpu.LoadBIOS(bios)
		}
	}

	f := func(screen *ebiten.Image) error {
		for i := 0; i < CyclesPerFrame; i++ {
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

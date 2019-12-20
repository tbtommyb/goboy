package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/tbtommyb/goboy/pkg/constants"
	"github.com/tbtommyb/goboy/pkg/cpu"
	"github.com/tbtommyb/goboy/pkg/display"
)

var EbitenFPS = 60
var CyclesPerFrame = cpu.GameboyClockSpeed / EbitenFPS

var keyMap = map[ebiten.Key]cpu.Button{
	ebiten.KeyZ:         cpu.ButtonA,
	ebiten.KeyX:         cpu.ButtonB,
	ebiten.KeyBackspace: cpu.ButtonSelect,
	ebiten.KeyEnter:     cpu.ButtonStart,
	ebiten.KeyRight:     cpu.ButtonRight,
	ebiten.KeyLeft:      cpu.ButtonLeft,
	ebiten.KeyUp:        cpu.ButtonUp,
	ebiten.KeyDown:      cpu.ButtonDown,
}

func main() {
	var loadBIOS bool
	var bios, rom []byte
	var err error

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	biosPtr := flag.String("bios", "", "BIOS path to read from")
	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Fatalf("ROM path not provided")
	}

	if *biosPtr != "" {
		bios, err = ioutil.ReadFile(filepath.Join(filepath.Dir(ex), *biosPtr))
		if err != nil {
			log.Fatalf("Error reading BIOS ROM %s", err.Error())
		}
		loadBIOS = true
	}

	rom, err = ioutil.ReadFile(filepath.Join(filepath.Dir(ex), flag.Args()[0]))
	if err != nil {
		log.Fatalf("Error reading ROM %s", err.Error())
	}

	fmt.Printf("%x\n", rom[:1000])
	gameboy := cpu.Init(loadBIOS)
	gameboy.LoadROM(rom)
	if loadBIOS {
		gameboy.LoadBIOS(bios)
	}
	display := display.Init()
	gameboy.AttachDisplay(display)

	f := func(screen *ebiten.Image) error {
		for i := 0; i < CyclesPerFrame; i++ {
			gameboy.HandleInterrupts()
			cycles := gameboy.Step()
			gameboy.RunFor(cycles)
		}

		for key, button := range keyMap {
			if inpututil.IsKeyJustPressed(key) {
				gameboy.PressButton(button)
			}
			if inpututil.IsKeyJustReleased(key) {
				gameboy.ReleaseButton(button)
			}
		}
		screen.ReplacePixels(display.Pixels())
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

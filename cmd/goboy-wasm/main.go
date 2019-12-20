package main

import (
	"fmt"

	"syscall/js"

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

type jsRom struct {
	name string
	data []byte
}

func runGame(rom []byte) {
	var err error
	gameboy := cpu.Init(false)
	gameboy.LoadROM(rom)
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
}

func main() {
	romChannel := make(chan jsRom)

	js.Global().Set("loadROM", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var rom []byte
		for _, el := range args[1].String() {
			rom = append(rom, byte(el))
		}
		romChannel <- jsRom{
			name: args[0].String(),
			data: rom,
		}
		return nil
	}))

	select {
	case rom := <-romChannel:
		runGame(rom.data)
	}

	return
}

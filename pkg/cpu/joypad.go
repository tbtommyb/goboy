package cpu

import (
	"github.com/tbtommyb/goboy/pkg/utils"
)

type Button byte

const (
	ButtonA      Button = 0
	ButtonB             = 1
	ButtonSelect        = 2
	ButtonStart         = 3
	ButtonRight         = 4
	ButtonLeft          = 5
	ButtonUp            = 6
	ButtonDown          = 7
)

func (cpu *CPU) PressButton(button Button) {
	cpu.joypad = utils.SetBit(byte(button), cpu.joypad, 0)
	// Request the joypad interrupt
}

func (cpu *CPU) ReleaseButton(button Button) {
	cpu.joypad = utils.SetBit(byte(button), cpu.joypad, 1)
	// Request the joypad interrupt
}

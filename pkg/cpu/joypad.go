package cpu

import (
	"github.com/tbtommyb/goboy/pkg/utils"
)

type Button byte

const (
	ButtonRight  Button = 0
	ButtonLeft          = 1
	ButtonUp            = 2
	ButtonDown          = 3
	ButtonA             = 4
	ButtonB             = 5
	ButtonSelect        = 6
	ButtonStart         = 7
)

func (cpu *CPU) PressButton(button Button) {
	previouslyUnset := utils.IsSet(byte(button), cpu.joypad)

	cpu.joypad = utils.SetBit(byte(button), cpu.joypad, 0)

	notDirectional := true

	if button > ButtonDown {
		notDirectional = false
	}

	buttonReq := cpu.memory.ioram[0] // janky
	requestInterrupt := false

	if notDirectional && !utils.IsSet(5, buttonReq) {
		requestInterrupt = true
	} else if !notDirectional && !utils.IsSet(4, buttonReq) {
		requestInterrupt = true
	}

	if requestInterrupt && !previouslyUnset {
		cpu.requestInterrupt(Input)
	}
}

func (cpu *CPU) ReleaseButton(button Button) {
	cpu.joypad = utils.SetBit(byte(button), cpu.joypad, 1)
}

func (cpu *CPU) getJoypadState() byte {
	res := cpu.memory.ioram[0]
	res ^= 0xFF

	if !utils.IsSet(4, res) {
		topJoypad := cpu.joypad >> 4
		topJoypad |= 0xF0
		res &= topJoypad
	} else if !utils.IsSet(5, res) {
		bottomJoypad := cpu.joypad & 0xF
		bottomJoypad |= 0xF0
		res &= bottomJoypad
	}
	return res
}

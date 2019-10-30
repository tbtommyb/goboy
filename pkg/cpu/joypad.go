package cpu

import (
	"github.com/tbtommyb/goboy/pkg/utils"
)

type Joypad struct {
	buttons byte
	selection
}

const joypadSelectionMask = 0x30

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

type selection byte

const (
	directional    selection = 0x20
	nonDirectional           = 0x10
)

func (cpu *CPU) PressButton(button Button) {
	initialButtons := cpu.joypadInternalState.buttons
	updatedButtons := utils.SetBit(byte(button), initialButtons, 1)

	cpu.joypadInternalState.buttons = updatedButtons

	if initialButtons == 0 && updatedButtons > initialButtons {
		cpu.requestInterrupt(Input)
	}
}

func (cpu *CPU) ReleaseButton(button Button) {
	initialButtons := cpu.joypadInternalState.buttons
	cpu.joypadInternalState.buttons = utils.SetBit(byte(button), initialButtons, 0)
}

func (cpu *CPU) getJoypadState() byte {
	return cpu.joypadInternalState.toRegisterFormat()
}

func (cpu *CPU) setJoypadSelection(value byte) {
	cpu.joypadInternalState.selection = selection(value & joypadSelectionMask)
}

func (joypad *Joypad) toRegisterFormat() byte {
	switch joypad.selection {
	case directional:
		return (^(joypad.buttons & 0xf) & 0xf) | byte(joypad.selection)
	case nonDirectional:
		return (^(joypad.buttons >> 4) & 0xf) | byte(joypad.selection)
	default:
		return byte(joypad.selection) | 0xf
	}
}

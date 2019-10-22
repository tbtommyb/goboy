package cpu

import (
	"github.com/tbtommyb/goboy/pkg/utils"
)

type Interrupt byte

const (
	VBlank        Interrupt = 0
	LCDCStatus              = 1
	TimerOverflow           = 3
	Input                   = 4
)

const (
	VBlankInterruptHandlerAddress        uint16 = 0x40
	LCDCStatusInterruptHandlerAddress           = 0x48
	TimerOverflowInterruptHandlerAddress        = 0x50
	InputInterruptHandlerAddress                = 0x60
)

const (
	InterruptFlagAddress   uint16 = 0xFF0F
	InterruptEnableAddress        = 0xFFFF
)

var Interrupts = []Interrupt{VBlank, LCDCStatus, TimerOverflow, Input}

func (cpu *CPU) requestInterrupt(interrupt Interrupt) {
	requested := cpu.memory.get(InterruptFlagAddress)
	requested = utils.SetBit(byte(interrupt), requested, 1)
	cpu.memory.set(InterruptFlagAddress, requested)
}

func (cpu *CPU) clearInterrupt(interrupt Interrupt) {
	requested := cpu.memory.get(InterruptFlagAddress)
	requested = utils.SetBit(byte(interrupt), requested, 0)
	cpu.memory.set(InterruptFlagAddress, requested)
}

func (cpu *CPU) CheckInterrupts() {
	if !cpu.interruptsEnabled() {
		return
	}

	requested := cpu.memory.get(InterruptFlagAddress)
	if requested == 0 {
		return
	}

	enabled := cpu.memory.get(InterruptEnableAddress)
	for i, interrupt := range Interrupts {
		if utils.IsSet(byte(i), requested) && utils.IsSet(byte(i), enabled) {
			cpu.serviceInterrupt(interrupt)
		}
	}
}

func (cpu *CPU) serviceInterrupt(interrupt Interrupt) {
	cpu.disableInterrupts()

	cpu.clearInterrupt(interrupt)

	high, low := utils.SplitPair(cpu.GetPC())
	cpu.pushStack(high)
	cpu.pushStack(low)

	switch interrupt {
	case VBlank:
		cpu.setPC(VBlankInterruptHandlerAddress)
	case LCDCStatus:
		cpu.setPC(LCDCStatusInterruptHandlerAddress)
	case TimerOverflow:
		cpu.setPC(TimerOverflowInterruptHandlerAddress)
	case Input:
		cpu.setPC(InputInterruptHandlerAddress)
	}
}

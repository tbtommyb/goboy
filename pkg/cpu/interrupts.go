package cpu

import (
	"github.com/tbtommyb/goboy/pkg/utils"
)

type Interrupt byte

const (
	VBlank        Interrupt = 0
	LCDCStatus              = 1
	TimerOverflow           = 2
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

func (cpu *CPU) HandleInterrupts() {
	var postHaltAddress uint16
	if cpu.halt {
		cpu.halt = false
		postHaltAddress = cpu.GetPC() + 1
	}

	if !cpu.interruptsEnabled() {
		if postHaltAddress != 0 {
			cpu.setPC(postHaltAddress)
		}
		return
	}
	requested := cpu.memory.get(InterruptFlagAddress)
	if requested == 0 {
		return
	}
	enabled := cpu.memory.get(InterruptEnableAddress)
	if enabled == 0 {
		return
	}

	for _, interrupt := range Interrupts {
		if utils.IsSet(byte(interrupt), requested) && utils.IsSet(byte(interrupt), enabled) {
			cpu.serviceInterrupt(interrupt)
		}
	}
}

func (cpu *CPU) requestInterrupt(interrupt Interrupt) {
	cpu.memory.setBitAt(InterruptFlagAddress, byte(interrupt), 1)
}

func (cpu *CPU) clearInterrupt(interrupt Interrupt) {
	cpu.memory.setBitAt(InterruptFlagAddress, byte(interrupt), 0)
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

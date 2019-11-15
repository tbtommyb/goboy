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
	requested := cpu.memory.get(InterruptFlagAddress)
	if requested == 0 {
		return
	}
	enabled := cpu.memory.get(InterruptEnableAddress)
	if enabled == 0 {
		return
	}

	cpu.halt = false
	if !cpu.interruptsEnabled() {
		return
	}

	for _, interrupt := range Interrupts {
		if utils.IsSet(byte(interrupt), requested) && utils.IsSet(byte(interrupt), enabled) {
			returnAddress := cpu.GetPC()
			if cpu.halt {
				returnAddress += 1
			}
			cpu.serviceInterrupt(interrupt, returnAddress)
		}
	}

}

func (cpu *CPU) requestInterrupt(interrupt Interrupt) {
	cpu.memory.setBitAt(InterruptFlagAddress, byte(interrupt), 1)
}

func (cpu *CPU) clearInterrupt(interrupt Interrupt) {
	cpu.memory.setBitAt(InterruptFlagAddress, byte(interrupt), 0)
}

func (cpu *CPU) serviceInterrupt(interrupt Interrupt, returnAddress uint16) {
	cpu.disableInterrupts()

	cpu.clearInterrupt(interrupt)

	high, low := utils.SplitPair(returnAddress)
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

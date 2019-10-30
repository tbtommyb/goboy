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
	VBlankInterruptHandlerAddress        Address = 0x40
	LCDCStatusInterruptHandlerAddress            = 0x48
	TimerOverflowInterruptHandlerAddress         = 0x50
	InputInterruptHandlerAddress                 = 0x60
)

const (
	InterruptFlagAddress   Address = 0xFF0F
	InterruptEnableAddress         = 0xFFFF
)

var Interrupts = []Interrupt{VBlank, LCDCStatus, TimerOverflow, Input}

func (cpu *CPU) HandleInterrupts(interrupts <-chan Interrupt, complete chan<- bool) {
	for interrupt := range interrupts {
		if !cpu.interruptsEnabled() {
			complete <- true
			continue
		}
		enabled := cpu.memory.get(InterruptEnableAddress)
		if utils.IsSet(byte(interrupt), enabled) {
			cpu.serviceInterrupt(interrupt)
		}
		complete <- true
	}
}

func (cpu *CPU) requestInterrupt(interrupt Interrupt) {
	cpu.memory.setBitAt(InterruptFlagAddress, byte(interrupt), 1)
	cpu.interrupts <- interrupt
	<-cpu.complete
}

func (cpu *CPU) clearInterrupt(interrupt Interrupt) {
	cpu.memory.setBitAt(InterruptFlagAddress, byte(interrupt), 0)
}

func (cpu *CPU) serviceInterrupt(interrupt Interrupt) {
	cpu.disableInterrupts()

	cpu.clearInterrupt(interrupt)

	high, low := utils.SplitPair(uint16(cpu.GetPC()))
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

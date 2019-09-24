package cpu

import (
	"github.com/tbtommyb/goboy/pkg/utils"
)

const InterruptFlagAddress = 0xFF0F
const InterruptEnableAddress = 0xFFFF
const InterruptCount = 4

type Interrupt byte

const (
	VBlank        Interrupt = 0
	LCDCStatus              = 1
	TimerOverflow           = 3
	Input                   = 4
)

func (cpu *CPU) requestInterrupt(interrupt Interrupt) {
	request := cpu.memory.get(InterruptFlagAddress)
	request = utils.SetBit(byte(interrupt), request, 1)
	cpu.memory.set(InterruptFlagAddress, request)
}

func (cpu *CPU) clearInterrupt(interrupt Interrupt) {
	request := cpu.memory.get(InterruptFlagAddress)
	request = utils.SetBit(byte(interrupt), request, 0)
	cpu.memory.set(InterruptFlagAddress, request)
}

func (cpu *CPU) CheckInterrupts() {
	if cpu.interruptsEnabled() {
		request := cpu.memory.get(InterruptFlagAddress)
		enabled := cpu.memory.get(InterruptEnableAddress)
		if request != 0 {
			for i := byte(0); i <= InterruptCount; i++ {
				if utils.IsSet(i, request) {
					if utils.IsSet(i, enabled) {
						cpu.serviceInterrupt(Interrupt(i))
					}
				}
			}
		}
	}
}

func (cpu *CPU) serviceInterrupt(interrupt Interrupt) {
	cpu.disableInterrupts()

	request := cpu.memory.get(InterruptFlagAddress)
	request = utils.SetBit(byte(interrupt), request, 0)
	cpu.memory.set(InterruptFlagAddress, request)

	high, low := utils.SplitPair(cpu.GetPC())
	cpu.pushStack(high)
	cpu.pushStack(low)

	switch interrupt {
	case VBlank:
		cpu.setPC(0x40)
	case LCDCStatus:
		cpu.setPC(0x48)
	case TimerOverflow:
		cpu.setPC(0x50)
	case Input:
		cpu.setPC(0x60)
	}
}

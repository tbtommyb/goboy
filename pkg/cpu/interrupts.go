package cpu

import (
	"github.com/tbtommyb/goboy/pkg/utils"
)

type Interrupt byte

const (
	VBlank     Interrupt = 0
	LCDCStatus           = 1
)

func (cpu *CPU) requestInterrupt(id byte) {
	request := cpu.memory.get(0xFF0F)
	request = utils.SetBit(id, request, 1)
	cpu.memory.set(0xFF0F, request)
}

func (cpu *CPU) clearInterrupt(id byte) {
	request := cpu.memory.get(0xFF0F)
	request = utils.SetBit(id, request, 0)
	cpu.memory.set(0xFF0F, request)
}

func (cpu *CPU) CheckInterrupts() {
	if cpu.interruptsEnabled() {
		request := cpu.memory.get(0xFF0F)
		enabled := cpu.memory.get(0xFFFF)
		if request > 0 {
			for i := byte(0); i < 5; i++ {
				if utils.IsSet(i, request) {
					if utils.IsSet(i, enabled) {
						cpu.serviceInterrupt(i)
					}
				}
			}
		}
	}
}

func (cpu *CPU) serviceInterrupt(interrupt byte) {
	cpu.disableInterrupts()

	request := cpu.memory.get(0xFF0F)
	request = utils.SetBit(interrupt, request, 0)
	cpu.memory.set(0xFF0F, request)

	high, low := utils.SplitPair(cpu.GetPC())
	cpu.pushStack(high)
	cpu.pushStack(low)

	switch interrupt {
	case 0:
		cpu.setPC(0x40)
	case 1:
		cpu.setPC(0x48)
	case 2:
		cpu.setPC(0x50)
	case 4:
		cpu.setPC(0x60)
	}
}

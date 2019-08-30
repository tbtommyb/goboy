package cpu

import "github.com/tbtommyb/goboy/pkg/utils"

func (cpu *CPU) UpdateTimers(cycles uint) {
	cpu.doDividerRegister(cycles)

	if cpu.isClockEnabled() {
		cpu.timerCounter -= int(cycles)

		if cpu.timerCounter <= 0 {
			cpu.setClockFreq()

			if cpu.getTIMA() == 255 {
				cpu.setTIMA(cpu.getTMA())
				cpu.requestInterrupt(2)
			} else {
				cpu.setTIMA(cpu.getTIMA() + 1)
			}
		}
	}
}

func (cpu *CPU) isClockEnabled() bool {
	return utils.IsSet(2, cpu.getTMC())
}

func (cpu *CPU) getClockFreq() byte {
	return cpu.memory.get(TMCAddress) & 0x3
}

func (cpu *CPU) setClockFreq() {
	freq := cpu.getClockFreq()
	switch freq {
	case 0:
		cpu.timerCounter = 1024
	case 1:
		cpu.timerCounter = 16
	case 2:
		cpu.timerCounter = 64
	case 3:
		cpu.timerCounter = 256
	}
}

func (cpu *CPU) doDividerRegister(cycles uint) {
	// TODO: use bytes and detect overflow?
	cpu.dividerRegister += cycles
	if cpu.dividerCounter >= 255 {
		cpu.dividerCounter = 0
		cpu.memory.set(0xFF04, cpu.memory.get(0xFF04)+1)
	}
}

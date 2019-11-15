package cpu

import (
	"github.com/tbtommyb/goboy/pkg/utils"
)

const (
	TimerControlBit      byte = 2
	InputClockSelectMask      = 3
)

// f/2^10, f/2^4, f/2^6, f/2^8
var inputClocks = []uint16{1024, 16, 64, 256}

func (cpu *CPU) UpdateTimers(cycles uint) {
	cpu.internalTimer += uint16(cycles)

	if !cpu.isTimerEnabled() {
		return
	}

	cpu.cyclesForCurrentTick -= int(cycles)
	if cpu.cyclesForCurrentTick > 0 {
		return
	}

	cpu.setTIMA(cpu.getTIMA() + 1)
	if cpu.getTIMA() == 0 {
		cpu.setTIMA(cpu.getTMA())
		cpu.requestInterrupt(TimerOverflow)
	}
	cpu.resetCyclesForCurrentTick()
}

func (cpu *CPU) isTimerEnabled() bool {
	return utils.IsSet(TimerControlBit, cpu.getTAC())
}

func (cpu *CPU) getClockFreq() uint16 {
	return inputClocks[cpu.getTAC()&InputClockSelectMask]
}

func (cpu *CPU) resetCyclesForCurrentTick() {
	cpu.cyclesForCurrentTick = int(cpu.getClockFreq())
}

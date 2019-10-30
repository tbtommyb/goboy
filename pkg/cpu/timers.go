package cpu

import (
	"github.com/tbtommyb/goboy/pkg/utils"
)

const (
	TimerControlBit      byte = 2
	InputClockSelectMask      = 3
)

const DIVAddress = 0xFF04
const TIMAAddress = 0xFF05
const TMAAddress = 0xFF06
const TACAddress = 0xFF07

// f/20^10, f/2^4, f/2^6, f/2^8
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

func (cpu *CPU) getDIV() byte {
	return cpu.memory.get(DIVAddress)
}

func (cpu *CPU) setDIV(value byte) {
	cpu.memory.set(DIVAddress, value)
}

func (cpu *CPU) getTIMA() byte {
	return cpu.memory.get(TIMAAddress)
}

func (cpu *CPU) setTIMA(value byte) {
	cpu.memory.set(TIMAAddress, value)
}

func (cpu *CPU) getTMA() byte {
	return cpu.memory.get(TMAAddress)
}

func (cpu *CPU) getTAC() byte {
	return cpu.memory.get(TACAddress)
}

func (cpu *CPU) setTAC(value byte) {
	cpu.memory.set(TACAddress, value)
}

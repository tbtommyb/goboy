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

func (cpu *CPU) UpdateTimers() {
	cpu.internalTimer++

	if !cpu.isTimerEnabled() {
		return
	}

	cpu.cyclesForCurrentTick--
	if cpu.cyclesForCurrentTick > 0 {
		return
	}

	cpu.WriteIO(TIMAAddress, cpu.ReadIO(TIMAAddress)+1)
	if cpu.ReadIO(TIMAAddress) == 0 {
		cpu.WriteIO(TIMAAddress, cpu.ReadIO(TMAAddress))
		cpu.requestInterrupt(TimerOverflow)
	}
	cpu.ResetCyclesForTimerTick()
}

func (cpu *CPU) GetInternalTimer() uint16 {
	return cpu.internalTimer
}

func (cpu *CPU) isTimerEnabled() bool {
	return utils.IsSet(TimerControlBit, cpu.ReadIO(TACAddress))
}

func (cpu *CPU) getClockFreq() uint16 {
	return inputClocks[cpu.ReadIO(TACAddress)&InputClockSelectMask]
}

func (cpu *CPU) ResetCyclesForTimerTick() {
	cpu.cyclesForCurrentTick = int(cpu.getClockFreq())
}

func (cpu *CPU) ResetInternalTimer() {
	cpu.internalTimer = 0
}

package cpu

import (
	c "github.com/tbtommyb/goboy/pkg/constants"
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

	cpu.WriteIO(c.TIMAAddress, cpu.ReadIO(c.TIMAAddress)+1)
	if cpu.ReadIO(c.TIMAAddress) == 0 {
		cpu.WriteIO(c.TIMAAddress, cpu.ReadIO(c.TMAAddress))
		cpu.requestInterrupt(TimerOverflow)
	}
	cpu.ResetCyclesForTimerTick()
}

func (cpu *CPU) GetInternalTimer() uint16 {
	return cpu.internalTimer
}

func (cpu *CPU) isTimerEnabled() bool {
	return utils.IsSet(TimerControlBit, cpu.ReadIO(c.TACAddress))
}

func (cpu *CPU) getClockFreq() uint16 {
	return inputClocks[cpu.ReadIO(c.TACAddress)&InputClockSelectMask]
}

func (cpu *CPU) ResetCyclesForTimerTick() {
	cpu.cyclesForCurrentTick = int(cpu.getClockFreq())
}

func (cpu *CPU) ResetInternalTimer() {
	cpu.internalTimer = 0
}

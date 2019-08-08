package cpu

import "github.com/tbtommyb/goboy/pkg/registers"

// TODO: create separate stack structure
func (cpu *CPU) GetSP() uint16 {
	return cpu.SP
}

func (cpu *CPU) setSP(value uint16) uint16 {
	cpu.incrementCycles()
	cpu.SP = value
	return value
}

func (cpu *CPU) incrementSP() {
	cpu.SP += 1
}

func (cpu *CPU) decrementSP() {
	cpu.SP -= 1
}

func (cpu *CPU) pushStack(val byte) byte {
	cpu.decrementSP()
	return cpu.SetMem(registers.SP, val)
}

func (cpu *CPU) popStack() byte {
	val := cpu.GetMem(registers.SP)
	cpu.incrementSP()
	return val
}

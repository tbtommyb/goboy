package cpu

// TODO: create separate stack structure
func (cpu *CPU) GetSP() uint16 {
	return cpu.SP
}

func (cpu *CPU) setSP(value uint16) uint16 {
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
	return cpu.SetMem(SP, val)
}

func (cpu *CPU) popStack() byte {
	val := cpu.GetMem(SP)
	cpu.incrementSP()
	return val
}

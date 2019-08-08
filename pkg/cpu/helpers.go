package cpu

func (cpu *CPU) computeOffset(offset uint16) uint16 {
	return 0xFF00 + offset
}

func (cpu *CPU) carryBit(withCarry bool, flag Flag) byte {
	var carry byte
	if withCarry && cpu.isSet(flag) {
		carry = 1
	}
	return carry
}

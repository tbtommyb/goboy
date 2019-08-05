package cpu

func (cpu *CPU) computeOffset(offset uint16) uint16 {
	return 0xFF00 + offset
}

func mergePair(high, low byte) uint16 {
	return uint16(high)<<8 | uint16(low)
}

func setBit(pos, value, flag byte) byte {
	value ^= (-flag ^ value) & (1 << pos)
	return value
}

func (cpu *CPU) carryBit(withCarry bool, flag Flag) byte {
	var carry byte
	if withCarry && cpu.isSet(flag) {
		carry = 1
	}
	return carry
}

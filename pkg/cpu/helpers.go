package cpu

func (cpu *CPU) computeOffset(offset uint16) uint16 {
	return 0xFF00 + offset
}

func mergePair(high, low byte) uint16 {
	return uint16(high)<<8 | uint16(low)
}

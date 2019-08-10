package utils

func MergePair(high, low byte) uint16 {
	return uint16(high)<<8 | uint16(low)
}

func ReverseMergePair(low, high byte) uint16 {
	return MergePair(high, low)
}

func SetBit(pos, input, bitValue byte) byte {
	input ^= (-bitValue ^ input) & (1 << pos)
	return input
}

func IsSet(pos, input byte) bool {
	return input&(1<<pos) > 0
}

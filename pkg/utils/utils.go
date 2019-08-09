package utils

func MergePair(high, low byte) uint16 {
	return uint16(high)<<8 | uint16(low)
}

func ReverseMergePair(low, high byte) uint16 {
	return MergePair(high, low)
}

func SetBit(pos, value, flag byte) byte {
	value ^= (-flag ^ value) & (1 << pos)
	return value
}

func IsSet(pos, value byte) bool {
	return value&(1<<pos) > 0
}

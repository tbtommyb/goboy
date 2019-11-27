package cpu

import (
	"math/bits"

	in "github.com/tbtommyb/goboy/pkg/instructions"
	"github.com/tbtommyb/goboy/pkg/registers"
)

func addOp(args ...byte) (byte, FlagSet) {
	a, b, carry := args[0], args[1], args[2]
	result := a + b + carry
	flagSet := FlagSet{
		Zero:      result == 0,
		Negative:  false,
		HalfCarry: isAddHalfCarry(a, b, carry),
		FullCarry: isAddFullCarry(a, b, carry),
	}
	return result, flagSet
}

func subOp(args ...byte) (byte, FlagSet) {
	a, b, carry := args[0], args[1], args[2]
	result := a - b - carry
	flagSet := FlagSet{
		Zero:      result == 0,
		Negative:  true,
		HalfCarry: isSubHalfCarry(a, b, carry),
		FullCarry: isSubFullCarry(a, b, carry),
	}
	return result, flagSet
}

func andOp(args ...byte) (byte, FlagSet) {
	a, b := args[0], args[1]
	result := a & b
	flagSet := FlagSet{
		Zero:      result == 0,
		Negative:  false,
		HalfCarry: true,
		FullCarry: false,
	}
	return result, flagSet
}

func orOp(args ...byte) (byte, FlagSet) {
	a, b := args[0], args[1]
	result := a | b
	flagSet := FlagSet{
		Zero:      result == 0,
		Negative:  false,
		HalfCarry: false,
		FullCarry: false,
	}
	return result, flagSet
}

func xorOp(args ...byte) (byte, FlagSet) {
	a, b := args[0], args[1]
	result := a ^ b
	flagSet := FlagSet{
		Zero:      result == 0,
		Negative:  false,
		HalfCarry: false,
		FullCarry: false,
	}
	return result, flagSet
}

func cmpOp(args ...byte) FlagSet {
	a, b := args[0], args[1]
	return FlagSet{
		Zero:      a == b,
		Negative:  true,
		HalfCarry: isSubHalfCarry(a, b, 0),
		FullCarry: isSubFullCarry(a, b, 0),
	}
}

func shiftOp(i in.Shift, value, flag byte) (byte, FlagSet) {
	var result byte
	var flags FlagSet
	switch i.GetDirection() {
	case in.Left:
		result = value << 1
		flags = FlagSet{
			FullCarry: bits.LeadingZeros8(value) == 0,
			Zero:      result == 0,
		}
	case in.Right:
		if i.IsWithCopy() {
			result = byte(int8(value) >> 1) // Sign-extend right shift
		} else {
			result = value >> 1
		}

		flags = FlagSet{
			FullCarry: bits.TrailingZeros8(value) == 0,
			Zero:      result == 0,
		}
	}
	return result, flags
}

func swapOp(value byte) (byte, FlagSet) {
	result := (value&0xf)<<4 | (value&0xf0)>>4
	flags := FlagSet{
		Zero: result == 0,
	}
	return result, flags
}

func rotateLeftOp(value byte, withCarry bool) (byte, FlagSet) {
	bit7 := value >> 7
	value = value << 1
	if withCarry {
		value = value | 0x1
	}

	var fc bool
	if bit7 != 0 {
		fc = true
	}
	return value, FlagSet{
		Negative:  false,
		HalfCarry: false,
		Zero:      value == 0,
		FullCarry: fc,
	}
}

func rotateLeftCarryOp(value byte) (byte, FlagSet) {
	result := (value << 1) | (value >> 7)
	fc := (value >> 7) > 0

	return result, FlagSet{
		Negative:  false,
		HalfCarry: false,
		Zero:      value == 0,
		FullCarry: fc,
	}
}

func rotateRightOp(value byte, withCarry bool) (byte, FlagSet) {
	var fc bool
	if value&0x1 == 0x1 {
		fc = true
	}

	value = value >> 1
	if withCarry {
		value = value | 0x80
	}
	return value, FlagSet{
		Zero:      value == 0,
		Negative:  false,
		HalfCarry: false,
		FullCarry: fc,
	}
}

func rotateRightCarryOp(value byte) (byte, FlagSet) {
	result := (value << 7) | (value >> 1)
	fc := value&0x1 > 0

	return result, FlagSet{
		Negative:  false,
		HalfCarry: false,
		Zero:      value == 0,
		FullCarry: fc,
	}
}

func (cpu *CPU) perform(f func(...byte) (byte, FlagSet), args ...byte) {
	result, flagSet := f(args...)
	cpu.Set(registers.A, result)
	cpu.setFlags(flagSet)
}

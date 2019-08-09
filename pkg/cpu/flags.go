package cpu

type Flag byte

type FlagSet struct {
	Zero, Negative, HalfCarry, FullCarry bool
}

const (
	Zero      Flag = 0x80
	Negative       = 0x40
	HalfCarry      = 0x20
	FullCarry      = 0x10
)

func isAddHalfCarry(a, b, carry byte) bool {
	return (a&0xf)+(b&0xf)+(carry&0xf) > 0xf
}

func isAddFullCarry(a, b, carry byte) bool {
	return uint16(a)+uint16(b)+uint16(carry) > 0xff
}

func isAddHalfCarry16(a, b uint16) bool {
	return (a&0xfff)+(b&0xfff) > 0xfff
}

func isAddFullCarry16(a, b uint16) bool {
	return uint32(a)+uint32(b) > 0xffff
}

func isSubHalfCarry(a, b, carry byte) bool {
	return int8(a&0xf)-int8(b&0xf)-int8(carry&0xf) < 0
}

func isSubFullCarry(a, b, carry byte) bool {
	return int16(a)-int16(b)-int16(carry) < 0
}

func (cpu *CPU) GetFlags() byte {
	return byte(cpu.flags)
}

func (cpu *CPU) setFlags(fs FlagSet) {
	cpu.setFlag(Zero, fs.Zero)
	cpu.setFlag(Negative, fs.Negative)
	cpu.setFlag(HalfCarry, fs.HalfCarry)
	cpu.setFlag(FullCarry, fs.FullCarry)
}

func (cpu *CPU) setFlag(flag Flag, value bool) {
	var bitValue int8
	if value {
		bitValue = 1
	}
	cpu.flags ^= byte((-bitValue ^ int8(cpu.flags)) & int8(flag))
}

func (cpu *CPU) isSet(flag Flag) bool {
	return (cpu.flags & byte(flag)) > 0
}

func (cpu *CPU) getFlag(flag Flag) byte {
	var value byte
	switch flag {
	case Zero:
		value = (cpu.flags & byte(Zero)) >> 7
	case Negative:
		value = (cpu.flags & byte(Negative)) >> 6
	case HalfCarry:
		value = (cpu.flags & byte(HalfCarry)) >> 5
	case FullCarry:
		value = (cpu.flags & byte(FullCarry)) >> 4
	}
	return value
}

func (cpu *CPU) enableInterrupts() {
	cpu.IME = true
}

func (cpu *CPU) disableInterrupts() {
	cpu.IME = false
}

func (cpu *CPU) interruptsEnabled() bool {
	return cpu.IME
}

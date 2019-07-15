package cpu

const LoadRegisterMask = 0xC0
const LoadRegisterPattern = 0x40
const LoadImmediateMask = 0xC7
const LoadImmediatePattern = 0x6
const LoadRegisterMemoryMask = 0xC7
const LoadRegisterMemoryPattern = 0x46
const StoreMemoryRegisterMask = 0xF8
const StoreMemoryRegisterPattern = 0x70

const DestRegisterMask = 0x38
const DestRegisterShift = 3
const SourceRegisterMask = 0x7

type Instruction interface {
	Opcode() []byte
}

type InvalidInstruction struct{ opcode byte }

func (i InvalidInstruction) Opcode() []byte { return []byte{i.opcode} }

type EmptyInstruction struct{}

func (i EmptyInstruction) Opcode() []byte { return []byte{0} }

type LoadRegister struct {
	source, dest Register
}

func (i LoadRegister) Opcode() []byte {
	return []byte{byte(LoadRegisterPattern | i.source | i.dest<<DestRegisterShift)}
}

type LoadImmediate struct {
	dest      Register
	immediate byte
}

func (i LoadImmediate) Opcode() []byte {
	return []byte{byte(LoadImmediatePattern | i.dest<<DestRegisterShift), i.immediate}
}

type LoadRegisterMemory struct {
	dest Register
}

func (i LoadRegisterMemory) Opcode() []byte {
	return []byte{byte(LoadRegisterMemoryPattern | i.dest<<DestRegisterShift)}
}

type StoreMemoryRegister struct {
	source Register
}

func (i StoreMemoryRegister) Opcode() []byte {
	return []byte{byte(StoreMemoryRegisterPattern | i.source)}
}

func Decode(op byte) Instruction {
	switch {
	case op&LoadRegisterMemoryMask == LoadRegisterMemoryPattern:
		// LD D, (HL), 0b01ddd110
		// This bit pattern conflicts with the value of L
		// so needs to come before LoadRegister check
		// OR take H, L out of const Register assignment?
		dest := Register(op & DestRegisterMask >> DestRegisterShift)

		return LoadRegisterMemory{dest}
	case op&StoreMemoryRegisterMask == StoreMemoryRegisterPattern:
		source := Register(op & SourceRegisterMask)
		return StoreMemoryRegister{source}
	case op&LoadRegisterMask == LoadRegisterPattern:
		// LD D, S. 0b01dddsss
		source := Register(op & SourceRegisterMask)
		dest := Register(op & DestRegisterMask >> DestRegisterShift)
		return LoadRegister{source, dest}
	case op&LoadImmediateMask == LoadImmediatePattern:
		// LD D, n. 0b00ddd110
		dest := Register(op & DestRegisterMask >> DestRegisterShift) // TODO extract this
		return LoadImmediate{dest: dest}
	case op == 0:
		return EmptyInstruction{}
	default:
		return InvalidInstruction{opcode: op}
	}
}

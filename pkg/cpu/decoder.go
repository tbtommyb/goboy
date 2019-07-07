package cpu

const LoadRegisterMask = 0xC0
const LoadRegisterPattern = 0x40
const LoadImmediateMask = 0xC7
const LoadImmediatePattern = 0x6

const DestRegisterMask = 0x38
const DestRegisterShift = 3
const SourceRegisterMask = 0x7

type Instruction interface{
	Opcode() []byte
}

type InvalidInstruction struct { opcode byte }
func (i InvalidInstruction) Opcode() []byte { return []byte{i.opcode} }

type LoadRegister struct {
	source, dest Register
}

func (i LoadRegister) Opcode() []byte {
	return []byte{byte((LoadRegisterPattern | i.source | i.dest << DestRegisterShift))}
}

type LoadImmediate struct {
	dest Register
	immediate byte
}

func (i LoadImmediate) Opcode() []byte { return []byte{0xFF} } // TODO implement

func (cpu *CPU) Decode(op byte) Instruction {
	switch {
	case op&LoadRegisterMask == LoadRegisterPattern: // LD D, S. 0b01dddsss
		source := Register(op & SourceRegisterMask)
		dest := Register(op & DestRegisterMask >> DestRegisterShift)
		return LoadRegister{source, dest}
		case op&LoadImmediateMask == LoadImmediatePattern: // LD D, n. 0b00ddd110
		dest := Register(op & DestRegisterMask >> DestRegisterShift) // TODO extract this
		immediate := cpu.fetchAndIncrement()
		return LoadImmediate{dest, immediate}
	default:
		return InvalidInstruction{opcode: op}
	}
}

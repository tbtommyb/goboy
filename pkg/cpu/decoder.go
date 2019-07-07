package cpu

const LoadRegisterMask = 0xC0
const LoadRegisterPattern = 0x40
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

func Decode(op byte) Instruction {
	switch {
	case op&LoadRegisterMask == LoadRegisterPattern:
		source := Register(op & SourceRegisterMask)
		dest := Register(op & DestRegisterMask >> DestRegisterShift)

		return LoadRegister{source, dest}
	default:
		return InvalidInstruction{opcode: op}
	}
}

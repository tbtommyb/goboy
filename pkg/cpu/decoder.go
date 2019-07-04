package cpu

type Opcode byte

type Instruction interface{} // TODO: specify interface
type LoadRegisterInstr struct {
	source, dest Register
}

const LoadRegisterMask = 0xC0
const LoadRegisterPattern = 0x40
const DestRegisterMask = 0x38
const DestRegisterShift = 3
const SourceRegisterMask = 0x7

func Decode(op Opcode) Instruction {
	switch {
	case op&LoadRegisterMask == LoadRegisterPattern:
		source := Register(op & SourceRegisterMask)
		dest := Register(op & DestRegisterMask >> DestRegisterShift)

		return LoadRegisterInstr{source, dest}
	default:
		return nil
	}
}

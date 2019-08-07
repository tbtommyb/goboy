package cpu

type InstructionIterator interface {
	next() byte
}

type InstructionList struct {
	instructions    []byte
	currentPosition int
}

func (il *InstructionList) Load(input []byte) {
	il.instructions = input
}

func (il *InstructionList) next() byte {
	if il.currentPosition >= len(il.instructions) {
		return 0
	}
	instruction := il.instructions[il.currentPosition]
	il.currentPosition += 1
	return instruction
}

// func Decoder(il InstructionList, handle func(Instruction)) {
func DecodeTest(il InstructionList, handle func(Instruction)) {
	for op := il.next(); op != 0; op = il.next() {
		switch {
		case op&MoveMask == MovePattern:
			// LD D, S. 0b01dd dsss
			handle(Move{source: source(op), dest: dest(op)})
		case op&MoveImmediateMask == MoveImmediatePattern:
			// LD D, n. 0b00dd d110
			immediate := il.next()
			handle(MoveImmediate{dest: dest(op), immediate: immediate})
		case op == 0:
			handle(EmptyInstruction{})
			return
		default:
			handle(InvalidInstruction{opcode: op})
			return
		}
	}
}

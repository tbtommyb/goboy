package cpu

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

package disassembler

type Disassembler struct {
	Instructions    []byte
	currentPosition int
}

func (il *Disassembler) Load(input []byte) {
	il.Instructions = input
	il.currentPosition = 0
}

func (il *Disassembler) Next() byte {
	if il.currentPosition >= len(il.Instructions) {
		return 0
	}
	instruction := il.Instructions[il.currentPosition]
	il.currentPosition += 1
	return instruction
}

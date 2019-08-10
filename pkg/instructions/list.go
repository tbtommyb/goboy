package instructions

type List struct {
	Instructions    []byte
	currentPosition int
}

func (il *List) Load(input []byte) {
	il.Instructions = input
	il.currentPosition = 0
}

func (il *List) Next() byte {
	if il.currentPosition >= len(il.Instructions) {
		return 0
	}
	instruction := il.Instructions[il.currentPosition]
	il.currentPosition += 1
	return instruction
}

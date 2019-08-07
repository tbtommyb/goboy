package main

import (
	"fmt"

	"github.com/tbtommyb/goboy/pkg/cpu"
)

func printOp(i cpu.Instruction) {
	fmt.Printf("%#v\n", i)
}

func main() {
	il := cpu.InstructionList{}
	il.Load([]byte{0x77, 0x45, 0x6, 0x1, 0x77})
	cpu.DecodeTest(il, printOp)
}

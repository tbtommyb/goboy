package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	in "github.com/tbtommyb/goboy/pkg/instructions"
)

func printOp(i in.Instruction) {
	fmt.Printf("%#v\n", i)
}

func main() {
	romPtr := flag.String("path", "input.rom", "ROM path to read from")
	flag.Parse()
	data, err := ioutil.ReadFile(*romPtr)
	if err != nil {
		fmt.Printf("File reading error %#v", err)
		return
	}
	disassembler := in.List{}
	disassembler.Load(data)

	// decoder.Decode(&disassembler, printOp)

}

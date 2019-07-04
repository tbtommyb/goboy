package main

import "github.com/tbtommyb/goboy/pkg/cpu"

func main() {
	gameboy := cpu.Init()
	gameboy.Set(cpu.A, 3)
	gameboy.LoadProgram([]byte{0x47, 0x48}) // LD B, A then LD C, B. Need better system for this
	gameboy.Run()
}

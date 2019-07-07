package cpu

import "testing"

func encode(instructions []Instruction) []byte {
	var opcodes []byte
	for _, instruction := range instructions {
		instrOpcodes := instruction.Opcode()
		for _, instrOpcode := range instrOpcodes {
			opcodes = append(opcodes, instrOpcode)
		}
	}
	return opcodes
}

func TestSetAndGetRegister(t *testing.T) {
	cpu := Init()

	cpu.Set(A, 3)
	cpu.LoadProgram(encode([]Instruction{
		LoadRegister{source: A, dest: B},
		LoadRegister{source: B, dest: C},
	}))

	cpu.Run()

	expectedOpcode := LoadRegister{source: B, dest: C}.Opcode()[0]
	if actual := cpu.memory[0x151]; actual != expectedOpcode {
		t.Fatalf("Expected 0x88, got %x", actual)
	}

	if regValue := cpu.Get(C); regValue != 3 {
		t.Fatalf("Expected 3, got %d", regValue)
	}
}

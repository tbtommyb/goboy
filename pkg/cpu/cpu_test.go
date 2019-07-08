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

func TestIncrementPC(t *testing.T) {
	testCases := []struct{instructions []Instruction; expected uint16}{
		{instructions: []Instruction{}, expected: 1},
		{instructions: []Instruction{LoadRegister{source: A, dest: B}}, expected: 2},
		{instructions: []Instruction{LoadRegister{source: A, dest: B}, LoadRegister{source: B, dest: C}}, expected: 3},
	}

	for _, test := range testCases {
		cpu := Init()
		initialPC := cpu.GetPC()
		cpu.LoadProgram(encode(test.instructions))
		cpu.Run()

		if currentPC := cpu.GetPC(); currentPC - initialPC != test.expected {
			t.Errorf("Incorrect PC value. Expected %d, got %d", test.expected, currentPC - initialPC)
		}
	}
}

func TestLoadProgram(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		LoadRegister{source: A, dest: B},
		LoadRegister{source: B, dest: C},
	}))
	cpu.Run()

	expectedOpcode := LoadRegister{source: B, dest: C}.Opcode()[0]
	if actual := cpu.memory[ProgramStartAddress + 1]; actual != expectedOpcode {
		t.Errorf("Expected 0x88, got %x", actual)
	}
}

func TestLoadImmediate(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: A, immediate: 0xFF},
	}))
	cpu.Run()

	if regValue := cpu.Get(A); regValue != 0xFF {
		t.Errorf("Expected 0xFF, got %x", regValue)
	}
}

func TestSetAndGetRegister(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: A, immediate: 3},
		LoadRegister{source: A, dest: B},
		LoadRegister{source: B, dest: C},
		LoadRegister{source: C, dest: D},
		LoadRegister{source: D, dest: E},
	}))
	cpu.Run()

	if regValue := cpu.Get(E); regValue != 3 {
		t.Errorf("Expected 3, got %d", regValue)
	}
}

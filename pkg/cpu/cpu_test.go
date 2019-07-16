package cpu

import (
	"testing"
)

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
	testCases := []struct {
		instructions []Instruction
		expected     uint16
	}{
		{instructions: []Instruction{}, expected: 1},
		{instructions: []Instruction{Load{source: A, dest: B}}, expected: 2},
		{instructions: []Instruction{Load{source: A, dest: B}, Load{source: B, dest: C}}, expected: 3},
	}

	for _, test := range testCases {
		cpu := Init()
		initialPC := cpu.GetPC()
		cpu.LoadProgram(encode(test.instructions))
		cpu.Run()

		if currentPC := cpu.GetPC(); currentPC-initialPC != test.expected {
			t.Errorf("Incorrect PC value. Expected %d, got %d", test.expected, currentPC-initialPC)
		}
	}
}

func TestSetGetHL(t *testing.T) {
	var expected uint16 = 0x1000
	cpu := Init()
	cpu.SetHL(expected)

	if actual := cpu.GetHL(); actual != expected {
		t.Errorf("Expected %x, got %x", expected, actual)
	}
}

func TestLoadProgram(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		Load{source: A, dest: B},
		Load{source: B, dest: C},
	}))
	cpu.Run()

	expectedOpcode := Load{source: B, dest: C}.Opcode()[0]
	if actual := cpu.memory[ProgramStartAddress+1]; actual != expectedOpcode {
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
		Load{source: A, dest: B},
		Load{source: B, dest: C},
		Load{source: C, dest: D},
		Load{source: D, dest: E},
	}))
	cpu.Run()

	if regValue := cpu.Get(E); regValue != 3 {
		t.Errorf("Expected 3, got %d", regValue)
	}
}

func TestLoadMemory(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF
	cpu.memory.Load(0x1234, []byte{expected})

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: H, immediate: 0x12},
		LoadImmediate{dest: L, immediate: 0x34},
		Load{dest: A, source: M},
	}))
	cpu.Run()

	if actual := cpu.Get(A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestStoreMemory(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: A, immediate: expected},
		LoadImmediate{dest: H, immediate: 0x12},
		LoadImmediate{dest: L, immediate: 0x34},
		Load{source: A, dest: M},
	}))
	cpu.Run()

	if actual := cpu.memory[0x1234]; actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestLoadPair(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF
	cpu.memory.Load(0x1234, []byte{expected})

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: B, immediate: 0x12},
		LoadImmediate{dest: C, immediate: 0x34},
		LoadPair{dest: A, source: BC},
	}))
	cpu.Run()

	if actual := cpu.Get(A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestStorePair(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: A, immediate: expected},
		LoadImmediate{dest: B, immediate: 0x12},
		LoadImmediate{dest: C, immediate: 0x34},
		LoadPair{source: A, dest: BC},
	}))
	cpu.Run()

	if actual := cpu.memory[0x1234]; actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestLoadRelativeC(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF
	cpu.memory.Set(0xFF03, expected)

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: C, immediate: 3},
		LoadRelative{addressType: RelativeC},
	}))
	cpu.Run()

	if actual := cpu.Get(A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestStoreRelativeC(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: C, immediate: 3},
		LoadImmediate{dest: A, immediate: expected},
		StoreRelative{addressType: RelativeC},
	}))
	cpu.Run()

	if actual := cpu.memory.Get(0xFF03); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

// TODO: reduce duplication in these tests
func TestLoadRelativeN(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF
	cpu.memory.Set(0xFF03, expected)

	cpu.LoadProgram(encode([]Instruction{
		LoadRelative{addressType: RelativeN, immediate: 3},
	}))
	cpu.Run()

	if actual := cpu.Get(A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestStoreRelativeN(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: A, immediate: expected},
		StoreRelative{addressType: RelativeN, immediate: 3},
	}))
	cpu.Run()

	if actual := cpu.memory.Get(0xFF03); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestLoadNN(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF
	cpu.memory.Set(0xFF03, expected)

	cpu.LoadProgram(encode([]Instruction{
		LoadNN{immediate: 0xFF03},
	}))
	cpu.Run()

	if actual := cpu.Get(A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestStoreNN(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: A, immediate: expected},
		StoreNN{immediate: 0xFF03},
	}))
	cpu.Run()

	if actual := cpu.memory.Get(0xFF03); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestLoadIncrement(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF
	cpu.memory.Load(0x1234, []byte{expected})

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: H, immediate: 0x12},
		LoadImmediate{dest: L, immediate: 0x34},
		LoadIncrement{increment: 1},
	}))
	cpu.Run()

	if actual := cpu.Get(A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
	if hl := cpu.GetHL(); hl != 0x1235 {
		t.Errorf("Expected %#X, got %#X", 0x1235, hl)
	}
}

func TestLoadDecrement(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF
	cpu.memory.Load(0x1234, []byte{expected})

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: H, immediate: 0x12},
		LoadImmediate{dest: L, immediate: 0x34},
		LoadIncrement{increment: -1},
	}))
	cpu.Run()

	if actual := cpu.Get(A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
	if hl := cpu.GetHL(); hl != 0x1233 {
		t.Errorf("Expected %#X, got %#X", 0x1233, hl)
	}
}

func TestStoreIncrement(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: A, immediate: expected},
		LoadImmediate{dest: H, immediate: 0x12},
		LoadImmediate{dest: L, immediate: 0x34},
		StoreIncrement{increment: 1},
	}))
	cpu.Run()

	if actual := cpu.memory[0x1234]; actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
	if hl := cpu.GetHL(); hl != 0x1235 {
		t.Errorf("Expected %#X, got %#X", 0x1235, hl)
	}
}

func TestStoreDecrement(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF

	cpu.LoadProgram(encode([]Instruction{
		LoadImmediate{dest: A, immediate: expected},
		LoadImmediate{dest: H, immediate: 0x12},
		LoadImmediate{dest: L, immediate: 0x34},
		StoreIncrement{increment: -1},
	}))
	cpu.Run()

	if actual := cpu.memory[0x1234]; actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
	if hl := cpu.GetHL(); hl != 0x1233 {
		t.Errorf("Expected %#X, got %#X", 0x1233, hl)
	}
}

func TestInstructionCycles(t *testing.T) {
	// one more than the instruction cycle count because fetching the empty
	// instruction that ends the Run() loop costs a cycle
	testCases := []struct {
		instructions []Instruction
		expected     uint
	}{
		{instructions: []Instruction{Load{source: A, dest: B}}, expected: 2},
		{instructions: []Instruction{LoadImmediate{dest: H, immediate: 0x12}}, expected: 3},
		{instructions: []Instruction{Load{dest: M, source: A}}, expected: 3},
		{instructions: []Instruction{LoadImmediate{dest: H, immediate: 0x12}, Load{source: A, dest: M}}, expected: 5},
		{instructions: []Instruction{LoadImmediate{immediate: 0x12, dest: M}}, expected: 4},
		{instructions: []Instruction{LoadPair{source: A, dest: BC}}, expected: 3},
		{instructions: []Instruction{LoadRelative{addressType: RelativeC}}, expected: 3},
		{instructions: []Instruction{StoreRelative{addressType: RelativeC}}, expected: 3},
		{instructions: []Instruction{LoadRelative{addressType: RelativeN}}, expected: 4},
		{instructions: []Instruction{StoreRelative{addressType: RelativeN}}, expected: 4},
		{instructions: []Instruction{StoreNN{}}, expected: 5},
		{instructions: []Instruction{LoadNN{}}, expected: 5},
		{instructions: []Instruction{LoadIncrement{increment: 1}}, expected: 3},
		{instructions: []Instruction{LoadIncrement{increment: -11}}, expected: 3},
		{instructions: []Instruction{StoreIncrement{increment: 1}}, expected: 3},
		{instructions: []Instruction{StoreIncrement{increment: -11}}, expected: 3},
	}

	for _, test := range testCases {
		cpu := Init()
		initialCycles := cpu.GetCycles()
		cpu.LoadProgram(encode(test.instructions))
		cpu.Run()

		if cycles := cpu.GetCycles(); cycles-initialCycles != test.expected {
			t.Errorf("Incorrect cycles value. Expected %d, got %d", test.expected, cycles-initialCycles)
		}
	}
}

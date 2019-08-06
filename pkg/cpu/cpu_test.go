package cpu

import (
	"fmt"
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
		{instructions: []Instruction{Move{source: A, dest: B}}, expected: 2},
		{instructions: []Instruction{Move{source: A, dest: B}, Move{source: B, dest: C}}, expected: 3},
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
		Move{source: A, dest: B},
		Move{source: B, dest: C},
	}))
	cpu.Run()

	expectedOpcode := Move{source: B, dest: C}.Opcode()[0]
	if actual := cpu.memory[ProgramStartAddress+1]; actual != expectedOpcode {
		t.Errorf("Expected 0x88, got %x", actual)
	}
}

func TestLoadImmediate(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: A, immediate: 0xFF},
	}))
	cpu.Run()

	if regValue := cpu.Get(A); regValue != 0xFF {
		t.Errorf("Expected 0xFF, got %x", regValue)
	}
}

func TestSetAndGetRegister(t *testing.T) {
	cpu := Init()

	var expected byte = 3
	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: A, immediate: expected},
		Move{source: A, dest: B},
		Move{source: B, dest: C},
		Move{source: C, dest: D},
		Move{source: D, dest: E},
	}))
	cpu.Run()

	if regValue := cpu.Get(E); regValue != 3 {
		t.Errorf("Expected %X, got %X", expected, regValue)
	}
}

func TestLoadMemory(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF
	cpu.memory.load(0x1234, []byte{expected})

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: H, immediate: 0x12},
		MoveImmediate{dest: L, immediate: 0x34},
		Move{dest: A, source: M},
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
		MoveImmediate{dest: A, immediate: expected},
		MoveImmediate{dest: H, immediate: 0x12},
		MoveImmediate{dest: L, immediate: 0x34},
		Move{source: A, dest: M},
	}))
	cpu.Run()

	if actual := cpu.memory[0x1234]; actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestLoadIndirect(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF
	cpu.memory.load(0x1234, []byte{expected})

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: B, immediate: 0x12},
		MoveImmediate{dest: C, immediate: 0x34},
		LoadIndirect{dest: A, source: BC},
	}))
	cpu.Run()

	if actual := cpu.Get(A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestStoreIndirect(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: A, immediate: expected},
		MoveImmediate{dest: B, immediate: 0x12},
		MoveImmediate{dest: C, immediate: 0x34},
		StoreIndirect{source: A, dest: BC},
	}))
	cpu.Run()

	if actual := cpu.memory[0x1234]; actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestLoadRelativeC(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF
	cpu.memory.set(0xFF03, expected)

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: C, immediate: 3},
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
		MoveImmediate{dest: C, immediate: 3},
		MoveImmediate{dest: A, immediate: expected},
		StoreRelative{addressType: RelativeC},
	}))
	cpu.Run()

	if actual := cpu.memory.get(0xFF03); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

// TODO: reduce duplication in these tests
func TestLoadRelativeN(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF
	cpu.memory.set(0xFF03, expected)

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
		MoveImmediate{dest: A, immediate: expected},
		StoreRelative{addressType: RelativeN, immediate: 3},
	}))
	cpu.Run()

	if actual := cpu.memory.get(0xFF03); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestLoadNN(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF
	cpu.memory.set(0xFF03, expected)

	cpu.LoadProgram(encode([]Instruction{
		LoadRelative{addressType: RelativeNN, immediate: 0xFF03},
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
		MoveImmediate{dest: A, immediate: expected},
		StoreRelative{addressType: RelativeNN, immediate: 0xFF03},
	}))
	cpu.Run()

	if actual := cpu.memory.get(0xFF03); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestLoadIncrement(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF
	cpu.memory.load(0x1234, []byte{expected})

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: H, immediate: 0x12},
		MoveImmediate{dest: L, immediate: 0x34},
		LoadIncrement{},
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
	cpu.memory.load(0x1234, []byte{expected})

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: H, immediate: 0x12},
		MoveImmediate{dest: L, immediate: 0x34},
		LoadDecrement{},
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
		MoveImmediate{dest: A, immediate: expected},
		MoveImmediate{dest: H, immediate: 0x12},
		MoveImmediate{dest: L, immediate: 0x34},
		StoreIncrement{},
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
		MoveImmediate{dest: A, immediate: expected},
		MoveImmediate{dest: H, immediate: 0x12},
		MoveImmediate{dest: L, immediate: 0x34},
		StoreDecrement{},
	}))
	cpu.Run()

	if actual := cpu.memory[0x1234]; actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
	if hl := cpu.GetHL(); hl != 0x1233 {
		t.Errorf("Expected %#X, got %#X", 0x1233, hl)
	}
}

func TestLoadRegisterPairImmediate(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		LoadRegisterPairImmediate{dest: BC, immediate: 0x1234},
		LoadRegisterPairImmediate{dest: DE, immediate: 0x1235},
		LoadRegisterPairImmediate{dest: HL, immediate: 0x1236},
		LoadRegisterPairImmediate{dest: SP, immediate: 0x1237},
	}))
	cpu.Run()

	if bc := cpu.GetBC(); bc != 0x1234 {
		t.Errorf("Expected %#X, got %#X", 0x1234, bc)
	}
	if de := cpu.GetDE(); de != 0x1235 {
		t.Errorf("Expected %#X, got %#X", 0x1235, de)
	}
	if hl := cpu.GetHL(); hl != 0x1236 {
		t.Errorf("Expected %#X, got %#X", 0x1236, hl)
	}
	if sp := cpu.GetSP(); sp != 0x1237 {
		t.Errorf("Expected %#X, got %#X", 0x1237, sp)
	}
}

func TestHLtoSP(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		LoadRegisterPairImmediate{dest: HL, immediate: 0x4321},
		HLtoSP{},
	}))
	cpu.Run()

	if sp := cpu.GetSP(); sp != 0x4321 {
		t.Errorf("Expected %#X, got %#X", 0x4321, sp)
	}
}

func TestPush(t *testing.T) {
	cpu := Init()

	startingSP := cpu.GetSP()
	cpu.LoadProgram(encode([]Instruction{
		LoadRegisterPairImmediate{dest: HL, immediate: 0x1236},
		Push{source: HL},
	}))
	cpu.Run()

	currentSP := cpu.GetSP()
	if currentSP != startingSP-2 {
		t.Errorf("SP incorrect: %#v\n", currentSP)
	}

	if actual := cpu.memory[currentSP : currentSP+2]; actual[0] != 0x36 || actual[1] != 0x12 {
		t.Errorf("Expected %#X, got %#X%X", 0x1236, actual[0], actual[1])
	}
}

func TestPop(t *testing.T) {
	cpu := Init()

	startingSP := cpu.GetSP()
	cpu.LoadProgram(encode([]Instruction{
		LoadRegisterPairImmediate{dest: HL, immediate: 0x1236},
		Push{source: HL},
		Pop{dest: BC},
	}))
	cpu.Run()

	currentSP := cpu.GetSP()
	if currentSP != startingSP {
		t.Errorf("SP incorrect: %#v\n", currentSP)
	}

	if cpu.GetBC() != cpu.GetHL() {
		t.Errorf("Expected %#X, got %#X", cpu.GetHL(), cpu.GetBC())
	}
}

func TestLoadHLSPPositive(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: H, immediate: 0xFF},
		MoveImmediate{dest: L, immediate: 0xF8},
		HLtoSP{},
		LoadHLSP{immediate: 2},
	}))
	cpu.Run()

	if actual := cpu.GetHL(); actual != 0xFFFA {
		t.Errorf("Expected %#X, got %#X\n", 0xFFFA, actual)
	}
	expectFlagSet(t, cpu, "load HL SP positive", FlagSet{})
}
func TestLoadHLSPNegative(t *testing.T) {
	cpu := Init()

	initialSP := cpu.GetSP()
	cpu.LoadProgram(encode([]Instruction{
		LoadHLSP{immediate: -10},
	}))
	cpu.Run()

	if actual := cpu.GetHL(); actual != initialSP-10 {
		t.Errorf("Expected %#X, got %#X\n", initialSP-10, actual)
	}
}

func TestStoreSP(t *testing.T) {
	cpu := Init()

	var initial uint16 = 0xFFCD
	cpu.setSP(initial)
	cpu.LoadProgram(encode([]Instruction{
		StoreSP{immediate: 0x1234},
	}))
	cpu.Run()

	first := cpu.memory.get(0x1234)
	second := cpu.memory.get(0x1235)
	if first != 0xCD {
		t.Errorf("Expected %#X, got %#X\n", 0xCD, first)
	}
	if second != 0xFF {
		t.Errorf("Expected %#X, got %#X\n", 0xFF, second)
	}
}

func TestArithmetic(t *testing.T) {
	testCases := []struct {
		name         string
		instructions []Instruction
		expected     byte
		flags        FlagSet
		withCarry    bool
	}{
		{
			name:     "add",
			expected: 0x0,
			flags:    FlagSet{Zero: true, FullCarry: true, HalfCarry: true},
			instructions: []Instruction{
				MoveImmediate{dest: B, immediate: 0xC6},
				MoveImmediate{dest: A, immediate: 0x3A},
				Add{source: B},
			},
		},
		{
			name:      "add with carry",
			withCarry: true,
			expected:  0xF1,
			flags:     FlagSet{HalfCarry: true},
			instructions: []Instruction{
				MoveImmediate{dest: E, immediate: 0x0F},
				MoveImmediate{dest: A, immediate: 0xE1},
				Add{source: E, withCarry: true},
			},
		},
		{
			name:     "add immediate",
			expected: 0x3B,
			flags:    FlagSet{FullCarry: true, HalfCarry: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x3C},
				AddImmediate{immediate: 0xFF},
			},
		},
		{
			name:      "add immediate with carry",
			withCarry: true,
			expected:  0x1D,
			flags:     FlagSet{FullCarry: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0xE1},
				AddImmediate{immediate: 0x3B, withCarry: true},
			},
		},
		{
			name:     "subtract",
			expected: 0x0,
			flags:    FlagSet{Negative: true, Zero: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x3E},
				MoveImmediate{dest: E, immediate: 0x3E},
				Subtract{source: E},
			},
		},
		{
			name:      "subtract with carry",
			withCarry: true,
			expected:  0x10,
			flags:     FlagSet{Negative: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x3B},
				MoveImmediate{dest: H, immediate: 0x2A},
				Subtract{source: H, withCarry: true},
			},
		},
		{
			name:     "subtract immediate",
			expected: 0x2F,
			flags:    FlagSet{Negative: true, HalfCarry: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x3E},
				SubtractImmediate{immediate: 0x0F},
			},
		},
		{
			name:      "subtract immediate with carry",
			withCarry: true,
			expected:  0x0,
			flags:     FlagSet{Negative: true, Zero: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x3B},
				SubtractImmediate{immediate: 0x3A, withCarry: true},
			},
		},
		{
			name:     "and",
			expected: 0x1A,
			flags:    FlagSet{HalfCarry: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x5A},
				MoveImmediate{dest: L, immediate: 0x3F},
				And{source: L},
			},
		},
		{
			name:     "and immediate",
			expected: 0x18,
			flags:    FlagSet{HalfCarry: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x5A},
				AndImmediate{immediate: 0x38},
			},
		},
		{
			name:     "or",
			expected: 0x5A,
			flags:    FlagSet{},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x5A},
				Or{source: A},
			},
		},
		{
			name:     "or immediate",
			expected: 0x5B,
			flags:    FlagSet{},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x5A},
				OrImmediate{immediate: 0x3},
			},
		},
		{
			name:     "xor",
			expected: 0x0,
			flags:    FlagSet{Zero: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0xFF},
				Xor{source: A},
			},
		},
		{
			name:     "xor immediate",
			expected: 0xF0,
			flags:    FlagSet{},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0xFF},
				XorImmediate{immediate: 0x0F},
			},
		},
		{
			name:     "inc",
			expected: 0x0,
			flags:    FlagSet{Zero: true, HalfCarry: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0xFF},
				Increment{dest: A},
			},
		},
		{
			name:     "dec",
			expected: 0x0,
			flags:    FlagSet{Zero: true, Negative: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x01},
				Decrement{dest: A},
			},
		},
	}
	for _, test := range testCases {
		cpu := Init()
		cpu.LoadProgram(encode(test.instructions))
		if test.withCarry {
			cpu.setFlag(FullCarry, true)
		}
		cpu.Run()

		if actual := cpu.Get(A); actual != test.expected {
			t.Errorf("%s: expected %#X, got %#X\n", test.name, test.expected, actual)
		}
		expectFlagSet(t, cpu, test.name, test.flags)
	}
}

func TestArithmeticMemory(t *testing.T) {
	testCases := []struct {
		name         string
		instructions []Instruction
		expected     byte
		flags        FlagSet
		withCarry    bool
		memory       byte
	}{
		{
			name:     "add memory",
			memory:   0x12,
			expected: 0x4E,
			flags:    FlagSet{},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x3C},
				Add{source: M},
			},
		},
		{
			name:      "add memory with carry",
			withCarry: true,
			memory:    0x1E,
			expected:  0x0,
			flags:     FlagSet{Zero: true, FullCarry: true, HalfCarry: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0xE1},
				Add{source: M, withCarry: true},
			},
		},
		{
			name:     "subtract memory",
			memory:   0x40,
			expected: 0xFE,
			flags:    FlagSet{Negative: true, FullCarry: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x3E},
				Subtract{source: M},
			},
		},
		{
			name:      "subtract memory with carry",
			memory:    0x4f,
			withCarry: true,
			expected:  0xEB,
			flags:     FlagSet{Negative: true, HalfCarry: true, FullCarry: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x3B},
				Subtract{source: M, withCarry: true},
			},
		},
		{
			name:     "and memory",
			memory:   0x0,
			expected: 0x0,
			flags:    FlagSet{HalfCarry: true, Zero: true},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x5A},
				And{source: M},
			},
		},
		{
			name:     "or memory",
			memory:   0xF,
			expected: 0x5F,
			flags:    FlagSet{},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0x5A},
				Or{source: M},
			},
		},
		{
			name:     "xor memory",
			memory:   0x8A,
			expected: 0x75,
			flags:    FlagSet{},
			instructions: []Instruction{
				MoveImmediate{dest: A, immediate: 0xFF},
				Xor{source: M},
			},
		},
	}

	for _, test := range testCases {
		cpu := Init()
		cpu.LoadProgram(encode(append([]Instruction{
			MoveImmediate{dest: H, immediate: 0x12},
			MoveImmediate{dest: L, immediate: 0x34},
		}, test.instructions...)))
		if test.withCarry {
			cpu.setFlag(FullCarry, true)
		}
		cpu.memory.load(0x1234, []byte{test.memory})
		cpu.Run()

		if actual := cpu.Get(A); actual != test.expected {
			t.Errorf("%s: expected %#X, got %#X\n", test.name, test.expected, actual)
		}
		expectFlagSet(t, cpu, test.name, test.flags)
	}
}

func TestIncrementMemory(t *testing.T) {
	cpu := Init()

	cpu.memory.load(0x1234, []byte{0x50})
	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: H, immediate: 0x12},
		MoveImmediate{dest: L, immediate: 0x34},
		Increment{M},
	}))
	cpu.Run()
	if actual := cpu.GetMem(HL); actual != 0x51 {
		t.Errorf("expected %#X, got %#X\n", 0x51, actual)
	}
	expectFlagSet(t, cpu, "inc memory", FlagSet{})
}

func TestDecrementMemory(t *testing.T) {
	cpu := Init()

	cpu.memory.load(0x1234, []byte{0x0})
	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: H, immediate: 0x12},
		MoveImmediate{dest: L, immediate: 0x34},
		Decrement{M},
	}))
	cpu.Run()
	if actual := cpu.GetMem(HL); actual != 0xFF {
		t.Errorf("expected %#X, got %#X\n", 0xFF, actual)
	}
	expectFlagSet(t, cpu, "inc memory", FlagSet{HalfCarry: true, Negative: true})
}

func TestCompare(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: A, immediate: 0x3C},
		MoveImmediate{dest: B, immediate: 0x2F},
		Cmp{source: B},
	}))
	cpu.Run()
	expectFlagSet(t, cpu, "cmp", FlagSet{Negative: true, HalfCarry: true})
}

func TestCompareImmediate(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: A, immediate: 0x3C},
		CmpImmediate{immediate: 0x3C},
	}))
	cpu.Run()
	expectFlagSet(t, cpu, "cmp immediate", FlagSet{Negative: true, Zero: true})
}

func TestCompareMemory(t *testing.T) {
	cpu := Init()

	cpu.memory.load(0x1234, []byte{0x40})
	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: H, immediate: 0x12},
		MoveImmediate{dest: L, immediate: 0x34},
		MoveImmediate{dest: A, immediate: 0x3C},
		Cmp{source: M},
	}))
	cpu.Run()
	expectFlagSet(t, cpu, "cmp memory", FlagSet{Negative: true, FullCarry: true})
}

func TestAddPair(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: H, immediate: 0x8A},
		MoveImmediate{dest: L, immediate: 0x23},
		MoveImmediate{dest: B, immediate: 0x06},
		MoveImmediate{dest: C, immediate: 0x05},
		AddPair{source: BC},
	}))
	cpu.Run()

	if actual := cpu.GetHL(); actual != 0x9028 {
		t.Errorf("expected %#X, got %#X\n", 0x9028, actual)
	}
	expectFlagSet(t, cpu, "cmp memory", FlagSet{HalfCarry: true})
}

// TODO: refactor into test cases
func TestAddPairSecond(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: H, immediate: 0x8A},
		MoveImmediate{dest: L, immediate: 0x23},
		MoveImmediate{dest: B, immediate: 0x06},
		MoveImmediate{dest: C, immediate: 0x05},
		AddPair{source: HL},
	}))
	cpu.Run()

	if actual := cpu.GetHL(); actual != 0x1446 {
		t.Errorf("expected %#X, got %#X\n", 0x1446, actual)
	}
	expectFlagSet(t, cpu, "cmp memory", FlagSet{HalfCarry: true, FullCarry: true})
}

func TestAddSP(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: H, immediate: 0xFF},
		MoveImmediate{dest: L, immediate: 0xF8},
		HLtoSP{},
		AddSP{immediate: 2},
	}))
	cpu.Run()

	if actual := cpu.GetSP(); actual != 0xFFFA {
		t.Errorf("expected %#X, got %#X\n", 0xFFFA, actual)
	}
	expectFlagSet(t, cpu, "cmp memory", FlagSet{})
}

func TestInstructionCycles(t *testing.T) {
	testCases := []struct {
		instructions []Instruction
		expected     uint
		message      string
	}{
		{instructions: []Instruction{Move{source: A, dest: B}}, expected: 1, message: "Move"},
		{instructions: []Instruction{MoveImmediate{dest: H, immediate: 0x12}}, expected: 2, message: "Move immediate"},
		{instructions: []Instruction{Move{dest: M, source: A}}, expected: 2, message: "Move memory"},
		{instructions: []Instruction{MoveImmediate{dest: H, immediate: 0x12}, Move{source: A, dest: M}}, expected: 4, message: "Move immediate and move"},
		{instructions: []Instruction{MoveImmediate{immediate: 0x12, dest: M}}, expected: 3, message: "Move immediate memory"},
		{instructions: []Instruction{LoadIndirect{dest: A, source: BC}}, expected: 2, message: "load Pair BC"},
		{instructions: []Instruction{StoreIndirect{dest: BC, source: A}}, expected: 2, message: "Store Pair BC"},
		{instructions: []Instruction{LoadRelative{addressType: RelativeC}}, expected: 2, message: "load Relative C"},
		{instructions: []Instruction{StoreRelative{addressType: RelativeC}}, expected: 2, message: "Store Relative C"},
		{instructions: []Instruction{LoadRelative{addressType: RelativeN}}, expected: 3, message: "load Relative N"},
		{instructions: []Instruction{StoreRelative{addressType: RelativeN}}, expected: 3, message: "Store Relative N"},
		{instructions: []Instruction{LoadRelative{addressType: RelativeNN}}, expected: 4, message: "load NN"},
		{instructions: []Instruction{StoreRelative{addressType: RelativeNN}}, expected: 4, message: "Store NN"},
		{instructions: []Instruction{LoadIncrement{}}, expected: 2, message: "load increment"},
		{instructions: []Instruction{LoadDecrement{}}, expected: 2, message: "load decrement"},
		{instructions: []Instruction{StoreIncrement{}}, expected: 2, message: "Store increment"},
		{instructions: []Instruction{StoreDecrement{}}, expected: 2, message: "Store decrement"},
		{instructions: []Instruction{LoadRegisterPairImmediate{dest: BC, immediate: 0x1234}}, expected: 3, message: "load register pair immediate"},
		{instructions: []Instruction{HLtoSP{}}, expected: 2, message: "HL to SP"},
		{instructions: []Instruction{Push{source: BC}}, expected: 4, message: "Push"},
		{instructions: []Instruction{Pop{dest: BC}}, expected: 3, message: "Pop"},
		{instructions: []Instruction{LoadHLSP{immediate: 20}}, expected: 3, message: "load HL SP"},
		{instructions: []Instruction{StoreSP{immediate: 0xDEAD}}, expected: 5, message: "Store SP"},
		{instructions: []Instruction{Add{source: B}}, expected: 1, message: "Add"},
		{instructions: []Instruction{Add{source: M}}, expected: 2, message: "Add from memory"},
		{instructions: []Instruction{AddImmediate{immediate: 0x12}}, expected: 2, message: "Add Immediate"},
		{instructions: []Instruction{Subtract{source: B}}, expected: 1, message: "Subtract"},
		{instructions: []Instruction{Subtract{source: M}}, expected: 2, message: "Subtract from memory"},
		{instructions: []Instruction{SubtractImmediate{immediate: 0x12}}, expected: 2, message: "Subtract Immediate"},
		{instructions: []Instruction{And{source: B}}, expected: 1, message: "And"},
		{instructions: []Instruction{And{source: M}}, expected: 2, message: "And from memory"},
		{instructions: []Instruction{AndImmediate{immediate: 0x12}}, expected: 2, message: "And Immediate"},
		{instructions: []Instruction{Or{source: B}}, expected: 1, message: "Or"},
		{instructions: []Instruction{Or{source: M}}, expected: 2, message: "Or from memory"},
		{instructions: []Instruction{OrImmediate{immediate: 0x12}}, expected: 2, message: "Or Immediate"},
		{instructions: []Instruction{Xor{source: B}}, expected: 1, message: "Xor"},
		{instructions: []Instruction{Xor{source: M}}, expected: 2, message: "Xor from memory"},
		{instructions: []Instruction{XorImmediate{immediate: 0x12}}, expected: 2, message: "Xor Immediate"},
		{instructions: []Instruction{Cmp{source: B}}, expected: 1, message: "Cmp"},
		{instructions: []Instruction{Cmp{source: M}}, expected: 2, message: "Cmp from memory"},
		{instructions: []Instruction{CmpImmediate{immediate: 0x12}}, expected: 2, message: "Cmp Immediate"},
		{instructions: []Instruction{Increment{dest: A}}, expected: 1, message: "Inc"},
		{instructions: []Instruction{Increment{dest: M}}, expected: 3, message: "Inc memory"},
		{instructions: []Instruction{Decrement{dest: A}}, expected: 1, message: "Dec"},
		{instructions: []Instruction{Decrement{dest: M}}, expected: 3, message: "Dec memory"},
		{instructions: []Instruction{AddPair{source: HL}}, expected: 2, message: "Add pair"},
		{instructions: []Instruction{AddSP{immediate: 3}}, expected: 4, message: "Add SP"},
	}

	for _, test := range testCases {
		cpu := Init()
		initialCycles := cpu.GetCycles()
		cpu.LoadProgram(encode(test.instructions))
		cpu.Run()

		// one more than the instruction cycle count because fetching the empty
		// instruction that ends the Run() loop costs a cycle
		if cycles := cpu.GetCycles(); cycles-initialCycles-1 != test.expected {
			t.Errorf("%s: Incorrect cycles value. Expected %d, got %d", test.message, test.expected, cycles-initialCycles-1)
		}
	}
}

func expectFlagSet(t *testing.T, cpu CPU, name string, fs FlagSet) {
	var errs []string
	if actual := cpu.isSet(Zero); actual != fs.Zero {
		errs = append(errs, fmt.Sprintf("expected Zero to be %t, got %t", fs.Zero, actual))
	}
	if actual := cpu.isSet(Negative); actual != fs.Negative {
		errs = append(errs, fmt.Sprintf("expected Negative to be %t, got %t", fs.Negative, actual))
	}
	if actual := cpu.isSet(HalfCarry); actual != fs.HalfCarry {
		errs = append(errs, fmt.Sprintf("expected HalfCarry to be %t, got %t", fs.HalfCarry, actual))
	}
	if actual := cpu.isSet(FullCarry); actual != fs.FullCarry {
		errs = append(errs, fmt.Sprintf("expected FullCarry to be %t, got %t", fs.FullCarry, actual))
	}
	for _, err := range errs {
		t.Errorf("%s: %s", name, err)
	}
}

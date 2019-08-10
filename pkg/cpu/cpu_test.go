package cpu

import (
	"fmt"
	"testing"

	"github.com/tbtommyb/goboy/pkg/conditions"
	in "github.com/tbtommyb/goboy/pkg/instructions"
	"github.com/tbtommyb/goboy/pkg/registers"
)

func encode(instructions []in.Instruction) []byte {
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
		instructions []in.Instruction
		expected     uint16
	}{
		{instructions: []in.Instruction{}, expected: 1},
		{instructions: []in.Instruction{in.Move{Source: registers.A, Dest: registers.B}}, expected: 2},
		{instructions: []in.Instruction{in.Move{Source: registers.A, Dest: registers.B}, in.Move{Source: registers.B, Dest: registers.C}}, expected: 3},
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

	cpu.LoadProgram(encode([]in.Instruction{
		in.Move{Source: registers.A, Dest: registers.B},
		in.Move{Source: registers.B, Dest: registers.C},
	}))
	cpu.Run()

	expectedOpcode := in.Move{Source: registers.B, Dest: registers.C}.Opcode()[0]
	if actual := cpu.memory[ProgramStartAddress+1]; actual != expectedOpcode {
		t.Errorf("Expected 0x88, got %x", actual)
	}
}

func TestLoadImmediate(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.A, Immediate: 0xFF},
	}))
	cpu.Run()

	if regValue := cpu.Get(registers.A); regValue != 0xFF {
		t.Errorf("Expected 0xFF, got %x", regValue)
	}
}

func TestSetAndGetRegister(t *testing.T) {
	cpu := Init()

	var expected byte = 3
	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.A, Immediate: expected},
		in.Move{Source: registers.A, Dest: registers.B},
		in.Move{Source: registers.B, Dest: registers.C},
		in.Move{Source: registers.C, Dest: registers.D},
		in.Move{Source: registers.D, Dest: registers.E},
	}))
	cpu.Run()

	if regValue := cpu.Get(registers.E); regValue != 3 {
		t.Errorf("Expected %X, got %X", expected, regValue)
	}
}

func TestLoadMemory(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF
	cpu.memory.load(0x1234, []byte{expected})

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
		in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		in.Move{Dest: registers.A, Source: registers.M},
	}))
	cpu.Run()

	if actual := cpu.Get(registers.A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestStoreMemory(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.A, Immediate: expected},
		in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
		in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		in.Move{Source: registers.A, Dest: registers.M},
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

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.B, Immediate: 0x12},
		in.MoveImmediate{Dest: registers.C, Immediate: 0x34},
		in.LoadIndirect{Dest: registers.A, Source: registers.BC},
	}))
	cpu.Run()

	if actual := cpu.Get(registers.A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestStoreIndirect(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.A, Immediate: expected},
		in.MoveImmediate{Dest: registers.B, Immediate: 0x12},
		in.MoveImmediate{Dest: registers.C, Immediate: 0x34},
		in.StoreIndirect{Source: registers.A, Dest: registers.BC},
	}))
	cpu.Run()

	if actual := cpu.memory[0x1234]; actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestLoadRelative(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF
	cpu.memory.set(0xFF03, expected)

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.C, Immediate: 3},
		in.LoadRelative{},
	}))
	cpu.Run()

	if actual := cpu.Get(registers.A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestStoreRelative(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.C, Immediate: 3},
		in.MoveImmediate{Dest: registers.A, Immediate: expected},
		in.StoreRelative{},
	}))
	cpu.Run()

	if actual := cpu.memory.get(0xFF03); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestLoadRelativeImmediateN(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF
	cpu.memory.set(0xFF03, expected)

	cpu.LoadProgram(encode([]in.Instruction{
		in.LoadRelativeImmediateN{Immediate: 3},
	}))
	cpu.Run()

	if actual := cpu.Get(registers.A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestStoreRelativeImmediateN(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.A, Immediate: expected},
		in.StoreRelativeImmediateN{Immediate: 3},
	}))
	cpu.Run()

	if actual := cpu.memory.get(0xFF03); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestLoadRelativeImmediateNN(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF
	cpu.memory.set(0xFF03, expected)

	cpu.LoadProgram(encode([]in.Instruction{
		in.LoadRelativeImmediateNN{Immediate: 0xFF03},
	}))
	cpu.Run()

	if actual := cpu.Get(registers.A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestStoreRelativeImmediateNN(t *testing.T) {
	cpu := Init()

	var expected byte = 0xFF

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.A, Immediate: expected},
		in.StoreRelativeImmediateNN{Immediate: 0xFF03},
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

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
		in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		in.LoadIncrement{},
	}))
	cpu.Run()

	if actual := cpu.Get(registers.A); actual != expected {
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

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
		in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		in.LoadDecrement{},
	}))
	cpu.Run()

	if actual := cpu.Get(registers.A); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
	if hl := cpu.GetHL(); hl != 0x1233 {
		t.Errorf("Expected %#X, got %#X", 0x1233, hl)
	}
}

func TestStoreIncrement(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.A, Immediate: expected},
		in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
		in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		in.StoreIncrement{},
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

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.A, Immediate: expected},
		in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
		in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		in.StoreDecrement{},
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

	cpu.LoadProgram(encode([]in.Instruction{
		in.LoadRegisterPairImmediate{Dest: registers.BC, Immediate: 0x1234},
		in.LoadRegisterPairImmediate{Dest: registers.DE, Immediate: 0x1235},
		in.LoadRegisterPairImmediate{Dest: registers.HL, Immediate: 0x1236},
		in.LoadRegisterPairImmediate{Dest: registers.SP, Immediate: 0x1237},
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

	cpu.LoadProgram(encode([]in.Instruction{
		in.LoadRegisterPairImmediate{Dest: registers.HL, Immediate: 0x4321},
		in.HLtoSP{},
	}))
	cpu.Run()

	if sp := cpu.GetSP(); sp != 0x4321 {
		t.Errorf("Expected %#X, got %#X", 0x4321, sp)
	}
}

func TestPush(t *testing.T) {
	cpu := Init()

	startingSP := cpu.GetSP()
	cpu.LoadProgram(encode([]in.Instruction{
		in.LoadRegisterPairImmediate{Dest: registers.HL, Immediate: 0x1236},
		in.Push{Source: registers.HL},
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
	cpu.LoadProgram(encode([]in.Instruction{
		in.LoadRegisterPairImmediate{Dest: registers.HL, Immediate: 0x1236},
		in.Push{Source: registers.HL},
		in.Pop{Dest: registers.BC},
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

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.H, Immediate: 0xFF},
		in.MoveImmediate{Dest: registers.L, Immediate: 0xF8},
		in.HLtoSP{},
		in.LoadHLSP{Immediate: 2},
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
	cpu.LoadProgram(encode([]in.Instruction{
		in.LoadHLSP{Immediate: -10},
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
	cpu.LoadProgram(encode([]in.Instruction{
		in.StoreSP{Immediate: 0x1234},
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
		instructions []in.Instruction
		expected     byte
		flags        FlagSet
		withCarry    bool
	}{
		{
			name:     "add",
			expected: 0x0,
			flags:    FlagSet{Zero: true, FullCarry: true, HalfCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.B, Immediate: 0xC6},
				in.MoveImmediate{Dest: registers.A, Immediate: 0x3A},
				in.Add{Source: registers.B},
			},
		},
		{
			name:      "add with carry",
			withCarry: true,
			expected:  0xF1,
			flags:     FlagSet{HalfCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.E, Immediate: 0x0F},
				in.MoveImmediate{Dest: registers.A, Immediate: 0xE1},
				in.Add{Source: registers.E, WithCarry: true},
			},
		},
		{
			name:     "add Immediate",
			expected: 0x3B,
			flags:    FlagSet{FullCarry: true, HalfCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x3C},
				in.AddImmediate{Immediate: 0xFF},
			},
		},
		{
			name:      "add Immediate with carry",
			withCarry: true,
			expected:  0x1D,
			flags:     FlagSet{FullCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0xE1},
				in.AddImmediate{Immediate: 0x3B, WithCarry: true},
			},
		},
		{
			name:     "subtract",
			expected: 0x0,
			flags:    FlagSet{Negative: true, Zero: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x3E},
				in.MoveImmediate{Dest: registers.E, Immediate: 0x3E},
				in.Subtract{Source: registers.E},
			},
		},
		{
			name:      "subtract with carry",
			withCarry: true,
			expected:  0x10,
			flags:     FlagSet{Negative: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x3B},
				in.MoveImmediate{Dest: registers.H, Immediate: 0x2A},
				in.Subtract{Source: registers.H, WithCarry: true},
			},
		},
		{
			name:     "subtract Immediate",
			expected: 0x2F,
			flags:    FlagSet{Negative: true, HalfCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x3E},
				in.SubtractImmediate{Immediate: 0x0F},
			},
		},
		{
			name:      "subtract Immediate with carry",
			withCarry: true,
			expected:  0x0,
			flags:     FlagSet{Negative: true, Zero: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x3B},
				in.SubtractImmediate{Immediate: 0x3A, WithCarry: true},
			},
		},
		{
			name:     "and",
			expected: 0x1A,
			flags:    FlagSet{HalfCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x5A},
				in.MoveImmediate{Dest: registers.L, Immediate: 0x3F},
				in.And{Source: registers.L},
			},
		},
		{
			name:     "and Immediate",
			expected: 0x18,
			flags:    FlagSet{HalfCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x5A},
				in.AndImmediate{Immediate: 0x38},
			},
		},
		{
			name:     "or",
			expected: 0x5A,
			flags:    FlagSet{},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x5A},
				in.Or{Source: registers.A},
			},
		},
		{
			name:     "or Immediate",
			expected: 0x5B,
			flags:    FlagSet{},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x5A},
				in.OrImmediate{Immediate: 0x3},
			},
		},
		{
			name:     "xor",
			expected: 0x0,
			flags:    FlagSet{Zero: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0xFF},
				in.Xor{Source: registers.A},
			},
		},
		{
			name:     "xor Immediate",
			expected: 0xF0,
			flags:    FlagSet{},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0xFF},
				in.XorImmediate{Immediate: 0x0F},
			},
		},
		{
			name:     "inc",
			expected: 0x0,
			flags:    FlagSet{Zero: true, HalfCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0xFF},
				in.Increment{Dest: registers.A},
			},
		},
		{
			name:     "dec",
			expected: 0x0,
			flags:    FlagSet{Zero: true, Negative: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x01},
				in.Decrement{Dest: registers.A},
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

		if actual := cpu.Get(registers.A); actual != test.expected {
			t.Errorf("%s: expected %#X, got %#X\n", test.name, test.expected, actual)
		}
		expectFlagSet(t, cpu, test.name, test.flags)
	}
}

func TestArithmeticMemory(t *testing.T) {
	testCases := []struct {
		name         string
		instructions []in.Instruction
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
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x3C},
				in.Add{Source: registers.M},
			},
		},
		{
			name:      "add memory with carry",
			withCarry: true,
			memory:    0x1E,
			expected:  0x0,
			flags:     FlagSet{Zero: true, FullCarry: true, HalfCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0xE1},
				in.Add{Source: registers.M, WithCarry: true},
			},
		},
		{
			name:     "subtract memory",
			memory:   0x40,
			expected: 0xFE,
			flags:    FlagSet{Negative: true, FullCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x3E},
				in.Subtract{Source: registers.M},
			},
		},
		{
			name:      "subtract memory with carry",
			memory:    0x4f,
			withCarry: true,
			expected:  0xEB,
			flags:     FlagSet{Negative: true, HalfCarry: true, FullCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x3B},
				in.Subtract{Source: registers.M, WithCarry: true},
			},
		},
		{
			name:     "and memory",
			memory:   0x0,
			expected: 0x0,
			flags:    FlagSet{HalfCarry: true, Zero: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x5A},
				in.And{Source: registers.M},
			},
		},
		{
			name:     "or memory",
			memory:   0xF,
			expected: 0x5F,
			flags:    FlagSet{},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x5A},
				in.Or{Source: registers.M},
			},
		},
		{
			name:     "xor memory",
			memory:   0x8A,
			expected: 0x75,
			flags:    FlagSet{},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0xFF},
				in.Xor{Source: registers.M},
			},
		},
	}

	for _, test := range testCases {
		cpu := Init()
		cpu.LoadProgram(encode(append([]in.Instruction{
			in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
			in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		}, test.instructions...)))
		if test.withCarry {
			cpu.setFlag(FullCarry, true)
		}
		cpu.memory.load(0x1234, []byte{test.memory})
		cpu.Run()

		if actual := cpu.Get(registers.A); actual != test.expected {
			t.Errorf("%s: expected %#X, got %#X\n", test.name, test.expected, actual)
		}
		expectFlagSet(t, cpu, test.name, test.flags)
	}
}

func TestIncrementMemory(t *testing.T) {
	cpu := Init()

	cpu.memory.load(0x1234, []byte{0x50})
	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
		in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		in.Increment{registers.M},
	}))
	cpu.Run()
	if actual := cpu.GetMem(registers.HL); actual != 0x51 {
		t.Errorf("expected %#X, got %#X\n", 0x51, actual)
	}
	expectFlagSet(t, cpu, "inc memory", FlagSet{})
}

func TestDecrementMemory(t *testing.T) {
	cpu := Init()

	cpu.memory.load(0x1234, []byte{0x0})
	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
		in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		in.Decrement{registers.M},
	}))
	cpu.Run()
	if actual := cpu.GetMem(registers.HL); actual != 0xFF {
		t.Errorf("expected %#X, got %#X\n", 0xFF, actual)
	}
	expectFlagSet(t, cpu, "inc memory", FlagSet{HalfCarry: true, Negative: true})
}

func TestCompare(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.A, Immediate: 0x3C},
		in.MoveImmediate{Dest: registers.B, Immediate: 0x2F},
		in.Cmp{Source: registers.B},
	}))
	cpu.Run()
	expectFlagSet(t, cpu, "cmp", FlagSet{Negative: true, HalfCarry: true})
}

func TestCompareImmediate(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.A, Immediate: 0x3C},
		in.CmpImmediate{Immediate: 0x3C},
	}))
	cpu.Run()
	expectFlagSet(t, cpu, "cmp Immediate", FlagSet{Negative: true, Zero: true})
}

func TestCompareMemory(t *testing.T) {
	cpu := Init()

	cpu.memory.load(0x1234, []byte{0x40})
	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
		in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		in.MoveImmediate{Dest: registers.A, Immediate: 0x3C},
		in.Cmp{Source: registers.M},
	}))
	cpu.Run()
	expectFlagSet(t, cpu, "cmp memory", FlagSet{Negative: true, FullCarry: true})
}

func TestAddPair(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.H, Immediate: 0x8A},
		in.MoveImmediate{Dest: registers.L, Immediate: 0x23},
		in.MoveImmediate{Dest: registers.B, Immediate: 0x06},
		in.MoveImmediate{Dest: registers.C, Immediate: 0x05},
		in.AddPair{Source: registers.BC},
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

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.H, Immediate: 0x8A},
		in.MoveImmediate{Dest: registers.L, Immediate: 0x23},
		in.MoveImmediate{Dest: registers.B, Immediate: 0x06},
		in.MoveImmediate{Dest: registers.C, Immediate: 0x05},
		in.AddPair{Source: registers.HL},
	}))
	cpu.Run()

	if actual := cpu.GetHL(); actual != 0x1446 {
		t.Errorf("expected %#X, got %#X\n", 0x1446, actual)
	}
	expectFlagSet(t, cpu, "cmp memory", FlagSet{HalfCarry: true, FullCarry: true})
}

func TestAddSP(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.H, Immediate: 0xFF},
		in.MoveImmediate{Dest: registers.L, Immediate: 0xF8},
		in.HLtoSP{},
		in.AddSP{Immediate: 2},
	}))
	cpu.Run()

	if actual := cpu.GetSP(); actual != 0xFFFA {
		t.Errorf("expected %#X, got %#X\n", 0xFFFA, actual)
	}
	expectFlagSet(t, cpu, "cmp memory", FlagSet{})
}

func TestIncrementPair(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.D, Immediate: 0x23},
		in.MoveImmediate{Dest: registers.E, Immediate: 0x5F},
		in.IncrementPair{Dest: registers.DE},
	}))
	cpu.Run()

	if actual := cpu.GetDE(); actual != 0x2360 {
		t.Errorf("expected %#X, got %#X\n", 0x2360, actual)
	}
}

func TestDecrementPair(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]in.Instruction{
		in.MoveImmediate{Dest: registers.D, Immediate: 0x23},
		in.MoveImmediate{Dest: registers.E, Immediate: 0x5F},
		in.DecrementPair{Dest: registers.DE},
	}))
	cpu.Run()

	if actual := cpu.GetDE(); actual != 0x235E {
		t.Errorf("expected %#X, got %#X\n", 0x235E, actual)
	}
}

func TestRotateA(t *testing.T) {
	testCases := []struct {
		name          string
		instructions  []in.Instruction
		inputFlags    FlagSet
		expected      byte
		expectedFlags FlagSet
		withCarry     bool
	}{
		{
			name:          "RLCA",
			expected:      0x0B,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{FullCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x85},
				in.RotateA{WithCopy: true, Direction: in.Left},
			},
		},
		{
			name:          "RLA",
			expected:      0x2B,
			inputFlags:    FlagSet{FullCarry: true},
			expectedFlags: FlagSet{FullCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x95},
				in.RotateA{Direction: in.Left},
			},
		},
		{
			name:          "RRCA",
			expected:      0x9D,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{FullCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x3B},
				in.RotateA{WithCopy: true, Direction: in.Right},
			},
		},
		{
			name:          "RRA",
			expected:      0x40,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{FullCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x81},
				in.RotateA{Direction: in.Right},
			},
		},
	}

	for _, test := range testCases {
		cpu := Init()
		cpu.LoadProgram(encode(test.instructions))
		cpu.setFlags(test.inputFlags)
		cpu.Run()

		if actual := cpu.Get(registers.A); actual != test.expected {
			t.Errorf("%s: expected %#X, got %#X\n", test.name, test.expected, actual)
		}
		expectFlagSet(t, cpu, test.name, test.expectedFlags)
	}
}

func TestRotateOperand(t *testing.T) {
	testCases := []struct {
		name          string
		instructions  []in.Instruction
		inputFlags    FlagSet
		expected      byte
		expectedFlags FlagSet
		withCarry     bool
	}{
		{
			name:          "RLC",
			expected:      0x0B,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{FullCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x85},
				in.RotateOperand{Source: registers.A, WithCopy: true, Direction: in.Left},
			},
		},
		{
			name:          "RL",
			expected:      0x00,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{FullCarry: true, Zero: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x80},
				in.RotateOperand{Source: registers.A, Direction: in.Left},
			},
		},
		{
			name:          "RRC",
			expected:      0x80,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{FullCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x1},
				in.RotateOperand{Source: registers.A, Direction: in.Right, WithCopy: true},
			},
		},
		{
			name:          "RR",
			expected:      0x0,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{FullCarry: true, Zero: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x1},
				in.RotateOperand{Source: registers.A, Direction: in.Right},
			},
		},
		{
			name:          "SLA",
			expected:      0x0,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{FullCarry: true, Zero: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x80},
				in.Shift{Source: registers.A, Direction: in.Left},
			},
		},
		{
			name:          "SRA",
			expected:      0xC5,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x8A},
				in.Shift{Source: registers.A, Direction: in.Right, WithCopy: true},
			},
		},
		{
			name:          "SRL",
			expected:      0x0,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{FullCarry: true, Zero: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x1},
				in.Shift{Source: registers.A, Direction: in.Right},
			},
		},
		{
			name:          "SWAP",
			expected:      0x0,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{Zero: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x0},
				in.Swap{Source: registers.A},
			},
		},
	}

	for _, test := range testCases {
		cpu := Init()
		cpu.LoadProgram(encode(test.instructions))
		cpu.setFlags(test.inputFlags)
		cpu.Run()

		if actual := cpu.Get(registers.A); actual != test.expected {
			t.Errorf("%s: expected %#X, got %#X\n", test.name, test.expected, actual)
		}
		expectFlagSet(t, cpu, test.name, test.expectedFlags)
	}
}

func TestRotateOperandWithMemory(t *testing.T) {
	testCases := []struct {
		name          string
		instructions  []in.Instruction
		inputFlags    FlagSet
		expected      byte
		expectedFlags FlagSet
		withCarry     bool
		memory        byte
	}{
		{
			name:          "RLC",
			memory:        0x0,
			expected:      0x0,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{Zero: true},
			instructions: []in.Instruction{
				in.RotateOperand{Source: registers.M, WithCopy: true, Direction: in.Left},
			},
		},
		{
			name:          "RL",
			memory:        0x11,
			expected:      0x22,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{},
			instructions: []in.Instruction{
				in.RotateOperand{Source: registers.M, Direction: in.Left},
			},
		},
		{
			name:          "RRC",
			memory:        0x0,
			expected:      0x0,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{Zero: true},
			instructions: []in.Instruction{
				in.RotateOperand{Source: registers.M, Direction: in.Right, WithCopy: true},
			},
		},
		{
			name:          "RR",
			memory:        0x8A,
			expected:      0x45,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{},
			instructions: []in.Instruction{
				in.RotateOperand{Source: registers.M, Direction: in.Right},
			},
		},
		{
			name:          "SLA",
			memory:        0xFF,
			expected:      0xFE,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{FullCarry: true},
			instructions: []in.Instruction{
				in.Shift{Source: registers.M, Direction: in.Left},
			},
		},
		{
			name:          "SRA",
			memory:        0x01,
			expected:      0x00,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{Zero: true, FullCarry: true},
			instructions: []in.Instruction{
				in.Shift{Source: registers.M, Direction: in.Right, WithCopy: true},
			},
		},
		{
			name:          "SRL",
			memory:        0xFF,
			expected:      0x7F,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{FullCarry: true},
			instructions: []in.Instruction{
				in.Shift{Source: registers.M, Direction: in.Right},
			},
		},
		{
			name:          "SWAP",
			memory:        0xF0,
			expected:      0x0F,
			inputFlags:    FlagSet{},
			expectedFlags: FlagSet{},
			instructions: []in.Instruction{
				in.Swap{Source: registers.M},
			},
		},
	}

	for _, test := range testCases {
		cpu := Init()
		cpu.LoadProgram(encode(append([]in.Instruction{
			in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
			in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		}, test.instructions...)))
		cpu.memory.load(0x1234, []byte{test.memory})
		cpu.setFlags(test.inputFlags)
		cpu.Run()

		if actual := cpu.GetMem(registers.HL); actual != test.expected {
			t.Errorf("%s: expected %#X, got %#X\n", test.name, test.expected, actual)
		}
		expectFlagSet(t, cpu, test.name, test.expectedFlags)
	}
}

func TestBit(t *testing.T) {
	testCases := []struct {
		name          string
		instructions  []in.Instruction
		expectedFlags FlagSet
		memory        byte
	}{
		{
			name:          "BIT",
			expectedFlags: FlagSet{HalfCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x80},
				in.Bit{BitNumber: 7, Source: registers.A},
			},
		},
		{
			name:          "BIT",
			expectedFlags: FlagSet{Zero: true, HalfCarry: true},
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0xEF},
				in.Bit{BitNumber: 4, Source: registers.A},
			},
		},
	}
	for _, test := range testCases {
		cpu := Init()
		cpu.LoadProgram(encode(test.instructions))
		cpu.Run()
		expectFlagSet(t, cpu, test.name, test.expectedFlags)
	}
}

func TestBitMemory(t *testing.T) {
	testCases := []struct {
		name          string
		instructions  []in.Instruction
		expectedFlags FlagSet
		memory        byte
	}{
		{
			name:          "BIT",
			memory:        0xFE,
			expectedFlags: FlagSet{Zero: true, HalfCarry: true},
			instructions: []in.Instruction{
				in.Bit{BitNumber: 0, Source: registers.M},
			},
		},
		{
			name:          "BIT",
			memory:        0xFE,
			expectedFlags: FlagSet{HalfCarry: true},
			instructions: []in.Instruction{
				in.Bit{BitNumber: 1, Source: registers.M},
			},
		},
	}
	for _, test := range testCases {
		cpu := Init()
		cpu.LoadProgram(encode(append([]in.Instruction{
			in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
			in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		}, test.instructions...)))
		cpu.memory.load(0x1234, []byte{test.memory})
		cpu.Run()

		expectFlagSet(t, cpu, test.name, test.expectedFlags)
	}
}

func TestSet(t *testing.T) {
	testCases := []struct {
		name         string
		instructions []in.Instruction
		expected     byte
	}{
		{
			name:     "SET",
			expected: 0x84,
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x80},
				in.Set{BitNumber: 2, Source: registers.A},
			},
		},
		{
			name:     "SET",
			expected: 0xBB,
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x3B},
				in.Set{BitNumber: 7, Source: registers.A},
			},
		},
		{
			name:     "RESET",
			expected: 0x0,
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x80},
				in.Reset{BitNumber: 7, Source: registers.A},
			},
		},
		{
			name:     "RESET",
			expected: 0x39,
			instructions: []in.Instruction{
				in.MoveImmediate{Dest: registers.A, Immediate: 0x3B},
				in.Reset{BitNumber: 1, Source: registers.A},
			},
		},
	}
	for _, test := range testCases {
		cpu := Init()
		cpu.LoadProgram(encode(test.instructions))
		cpu.Run()

		if actual := cpu.Get(registers.A); actual != test.expected {
			t.Errorf("Expected %x, got %x", test.expected, actual)
		}
	}
}

func TestSetMemory(t *testing.T) {
	testCases := []struct {
		name         string
		instructions []in.Instruction
		expected     byte
		memory       byte
	}{
		{
			name:     "SET",
			memory:   0x00,
			expected: 0x8,
			instructions: []in.Instruction{
				in.Set{BitNumber: 3, Source: registers.M},
			},
		},
		{
			name:     "RESET",
			memory:   0xFF,
			expected: 0xF7,
			instructions: []in.Instruction{
				in.Reset{BitNumber: 3, Source: registers.M},
			},
		},
	}
	for _, test := range testCases {
		cpu := Init()
		cpu.LoadProgram(encode(append([]in.Instruction{
			in.MoveImmediate{Dest: registers.H, Immediate: 0x12},
			in.MoveImmediate{Dest: registers.L, Immediate: 0x34},
		}, test.instructions...)))
		cpu.memory.load(0x1234, []byte{test.memory})
		cpu.Run()

		if actual := cpu.Get(registers.M); actual != test.expected {
			t.Errorf("Expected %x, got %x", test.expected, actual)
		}
	}
}

func TestJump(t *testing.T) {
	testCases := []struct {
		name         string
		instructions []in.Instruction
		expected     uint16
	}{
		{
			name:     "JP immediate",
			expected: 0x8000,
			instructions: []in.Instruction{
				in.JumpImmediate{Immediate: 0x8000},
			},
		},
		{
			name:     "JR",
			expected: ProgramStartAddress + 10,
			instructions: []in.Instruction{
				in.JumpRelative{Immediate: 10},
			},
		},
	}
	for _, test := range testCases {
		cpu := Init()
		cpu.LoadProgram(encode(test.instructions))
		cpu.Run()

		// -1 becuase the Run loop goes to next instruction before failing
		if actual := cpu.GetPC(); actual-1 != test.expected {
			t.Errorf("Expected %x, got %x", test.expected, actual-1)
		}
	}
}

func TestJumpConditional(t *testing.T) {
	testCases := []struct {
		name         string
		instructions []in.Instruction
		flags        FlagSet
		expected     uint16
	}{
		{
			name:     "JP conditional ?NZ",
			expected: ProgramStartAddress + 3,
			flags:    FlagSet{Zero: true},
			instructions: []in.Instruction{
				in.JumpImmediateConditional{Condition: conditions.NZ, Immediate: 0x8000},
			},
		},
		{
			name:     "JP conditional Z",
			expected: 0x8000,
			flags:    FlagSet{Zero: true},
			instructions: []in.Instruction{
				in.JumpImmediateConditional{Condition: conditions.Z, Immediate: 0x8000},
			},
		},
		{
			name:     "JP conditional C",
			expected: ProgramStartAddress + 3,
			flags:    FlagSet{Zero: true},
			instructions: []in.Instruction{
				in.JumpImmediateConditional{Condition: conditions.C, Immediate: 0x8000},
			},
		},
		{
			name:     "JP conditional NC",
			expected: 0x8000,
			flags:    FlagSet{Zero: true},
			instructions: []in.Instruction{
				in.JumpImmediateConditional{Condition: conditions.NC, Immediate: 0x8000},
			},
		},
		{
			name:     "JR conditional NC",
			expected: ProgramStartAddress + 12,
			flags:    FlagSet{Zero: true},
			instructions: []in.Instruction{
				in.JumpRelativeConditional{Condition: conditions.Z, Immediate: 12},
			},
		},
	}
	for _, test := range testCases {
		cpu := Init()
		cpu.LoadProgram(encode(test.instructions))
		cpu.setFlags(test.flags)
		cpu.Run()

		// -1 becuase the Run loop goes to next instruction before failing
		if actual := cpu.GetPC(); actual-1 != test.expected {
			t.Errorf("Expected %x, got %x", test.expected, actual-1)
		}
	}
}

func TestInstructionCycles(t *testing.T) {
	testCases := []struct {
		instructions []in.Instruction
		expected     uint
		message      string
	}{
		{instructions: []in.Instruction{in.Move{Source: registers.A, Dest: registers.B}}, expected: 1, message: "in.Move"},
		{instructions: []in.Instruction{in.MoveImmediate{Dest: registers.H, Immediate: 0x12}}, expected: 2, message: "in.Move Immediate"},
		{instructions: []in.Instruction{in.Move{Dest: registers.M, Source: registers.A}}, expected: 2, message: "Move memory"},
		{instructions: []in.Instruction{in.MoveImmediate{Dest: registers.H, Immediate: 0x12}, in.Move{Source: registers.A, Dest: registers.M}}, expected: 4, message: "Move Immediate and move"},
		{instructions: []in.Instruction{in.MoveImmediate{Immediate: 0x12, Dest: registers.M}}, expected: 3, message: "Move Immediate memory"},
		{instructions: []in.Instruction{in.LoadIndirect{Dest: registers.A, Source: registers.BC}}, expected: 2, message: "load Pair BC"},
		{instructions: []in.Instruction{in.StoreIndirect{Dest: registers.BC, Source: registers.A}}, expected: 2, message: "Store Pair BC"},
		{instructions: []in.Instruction{in.LoadRelative{}}, expected: 2, message: "load Relative"},
		{instructions: []in.Instruction{in.StoreRelative{}}, expected: 2, message: "Store Relative"},
		{instructions: []in.Instruction{in.LoadRelativeImmediateN{Immediate: 0x1}}, expected: 3, message: "load Relative N"},
		{instructions: []in.Instruction{in.StoreRelativeImmediateN{Immediate: 0x1}}, expected: 3, message: "Store Relative N"},
		{instructions: []in.Instruction{in.LoadRelativeImmediateNN{Immediate: 0x2}}, expected: 4, message: "load NN"},
		{instructions: []in.Instruction{in.StoreRelativeImmediateNN{}}, expected: 4, message: "Store NN"},
		{instructions: []in.Instruction{in.LoadIncrement{}}, expected: 2, message: "load increment"},
		{instructions: []in.Instruction{in.LoadDecrement{}}, expected: 2, message: "load decrement"},
		{instructions: []in.Instruction{in.StoreIncrement{}}, expected: 2, message: "Store increment"},
		{instructions: []in.Instruction{in.StoreDecrement{}}, expected: 2, message: "Store decrement"},
		{instructions: []in.Instruction{in.LoadRegisterPairImmediate{Dest: registers.BC, Immediate: 0x1234}}, expected: 3, message: "load register pair Immediate"},
		{instructions: []in.Instruction{in.HLtoSP{}}, expected: 2, message: "HL to SP"},
		{instructions: []in.Instruction{in.Push{Source: registers.BC}}, expected: 4, message: "Push"},
		{instructions: []in.Instruction{in.Pop{Dest: registers.BC}}, expected: 3, message: "Pop"},
		{instructions: []in.Instruction{in.LoadHLSP{Immediate: 20}}, expected: 3, message: "load HL SP"},
		{instructions: []in.Instruction{in.StoreSP{Immediate: 0xDEAD}}, expected: 5, message: "Store SP"},
		{instructions: []in.Instruction{in.Add{Source: registers.B}}, expected: 1, message: "Add"},
		{instructions: []in.Instruction{in.Add{Source: registers.M}}, expected: 2, message: "Add from memory"},
		{instructions: []in.Instruction{in.AddImmediate{Immediate: 0x12}}, expected: 2, message: "Add Immediate"},
		{instructions: []in.Instruction{in.Subtract{Source: registers.B}}, expected: 1, message: "Subtract"},
		{instructions: []in.Instruction{in.Subtract{Source: registers.M}}, expected: 2, message: "Subtract from memory"},
		{instructions: []in.Instruction{in.SubtractImmediate{Immediate: 0x12}}, expected: 2, message: "Subtract Immediate"},
		{instructions: []in.Instruction{in.And{Source: registers.B}}, expected: 1, message: "And"},
		{instructions: []in.Instruction{in.And{Source: registers.M}}, expected: 2, message: "And from memory"},
		{instructions: []in.Instruction{in.AndImmediate{Immediate: 0x12}}, expected: 2, message: "And Immediate"},
		{instructions: []in.Instruction{in.Or{Source: registers.B}}, expected: 1, message: "Or"},
		{instructions: []in.Instruction{in.Or{Source: registers.M}}, expected: 2, message: "Or from memory"},
		{instructions: []in.Instruction{in.OrImmediate{Immediate: 0x12}}, expected: 2, message: "Or Immediate"},
		{instructions: []in.Instruction{in.Xor{Source: registers.B}}, expected: 1, message: "Xor"},
		{instructions: []in.Instruction{in.Xor{Source: registers.M}}, expected: 2, message: "Xor from memory"},
		{instructions: []in.Instruction{in.XorImmediate{Immediate: 0x12}}, expected: 2, message: "Xor Immediate"},
		{instructions: []in.Instruction{in.Cmp{Source: registers.B}}, expected: 1, message: "Cmp"},
		{instructions: []in.Instruction{in.Cmp{Source: registers.M}}, expected: 2, message: "Cmp from memory"},
		{instructions: []in.Instruction{in.CmpImmediate{Immediate: 0x12}}, expected: 2, message: "Cmp Immediate"},
		{instructions: []in.Instruction{in.Increment{Dest: registers.A}}, expected: 1, message: "Inc"},
		{instructions: []in.Instruction{in.Increment{Dest: registers.M}}, expected: 3, message: "Inc memory"},
		{instructions: []in.Instruction{in.Decrement{Dest: registers.A}}, expected: 1, message: "Dec"},
		{instructions: []in.Instruction{in.Decrement{Dest: registers.M}}, expected: 3, message: "Dec memory"},
		{instructions: []in.Instruction{in.AddPair{Source: registers.HL}}, expected: 2, message: "Add pair"},
		{instructions: []in.Instruction{in.AddSP{Immediate: 3}}, expected: 4, message: "Add SP"},
		{instructions: []in.Instruction{in.IncrementPair{Dest: registers.DE}}, expected: 2, message: "Increment pair"},
		{instructions: []in.Instruction{in.DecrementPair{Dest: registers.DE}}, expected: 2, message: "Decrement pair"},
		{instructions: []in.Instruction{in.RotateA{WithCopy: true, Direction: in.Left}}, expected: 1, message: "RLCA"},
		{instructions: []in.Instruction{in.RotateA{Direction: in.Right}}, expected: 1, message: "RRA"},
		{instructions: []in.Instruction{in.RotateOperand{Direction: in.Left, WithCopy: true, Source: registers.A}}, expected: 2, message: "RLC"},
		{instructions: []in.Instruction{in.RotateOperand{Direction: in.Left, WithCopy: true, Source: registers.M}}, expected: 4, message: "RLC"},
		{instructions: []in.Instruction{in.Shift{Direction: in.Right, Source: registers.A}}, expected: 2, message: "SRL"},
		{instructions: []in.Instruction{in.Shift{Direction: in.Right, Source: registers.M}}, expected: 4, message: "SRL"},
		{instructions: []in.Instruction{in.Swap{Source: registers.B}}, expected: 2, message: "Swap register"},
		{instructions: []in.Instruction{in.Swap{Source: registers.M}}, expected: 4, message: "Swap memory"},
		{instructions: []in.Instruction{in.Bit{BitNumber: 2, Source: registers.C}}, expected: 2, message: "Bit"},
		{instructions: []in.Instruction{in.Bit{BitNumber: 2, Source: registers.M}}, expected: 3, message: "Bit memory"},
		{instructions: []in.Instruction{in.Set{BitNumber: 2, Source: registers.M}}, expected: 4, message: "Set memory"},
		{instructions: []in.Instruction{in.Set{BitNumber: 0, Source: registers.A}}, expected: 2, message: "Set"},
		{instructions: []in.Instruction{in.Reset{BitNumber: 0, Source: registers.A}}, expected: 2, message: "Reset"},
		{instructions: []in.Instruction{in.Reset{BitNumber: 0, Source: registers.M}}, expected: 4, message: "Reset memory"},
		{instructions: []in.Instruction{in.JumpImmediate{Immediate: 0x1234}}, expected: 4, message: "Jump immediate"},
		{instructions: []in.Instruction{in.JumpImmediateConditional{Condition: conditions.NC, Immediate: 0x1234}}, expected: 4, message: "Jump conditional met"},
		{instructions: []in.Instruction{
			in.Add{Source: registers.A},
			in.JumpImmediateConditional{Condition: conditions.NZ, Immediate: 0x1234},
		}, expected: 4, message: "Jump conditional not met"},
		{instructions: []in.Instruction{in.JumpRelative{Immediate: 2}}, expected: 3, message: "Jump relative"},
		{instructions: []in.Instruction{in.JumpRelativeConditional{Condition: conditions.NC, Immediate: 2}}, expected: 3, message: "JR conditional met"},
		{instructions: []in.Instruction{
			in.Add{Source: registers.A},
			in.JumpRelativeConditional{Condition: conditions.NZ, Immediate: 2},
		}, expected: 3, message: "JR conditional not met"},
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

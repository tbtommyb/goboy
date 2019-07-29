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

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: A, immediate: 3},
		Move{source: A, dest: B},
		Move{source: B, dest: C},
		Move{source: C, dest: D},
		Move{source: D, dest: E},
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
	cpu.memory.Load(0x1234, []byte{expected})

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
	cpu.memory.Set(0xFF03, expected)

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
		MoveImmediate{dest: A, immediate: expected},
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

	if actual := cpu.memory.Get(0xFF03); actual != expected {
		t.Errorf("Expected %#X, got %#X", expected, actual)
	}
}

func TestLoadIncrement(t *testing.T) {
	cpu := Init()
	var expected byte = 0xFF
	cpu.memory.Load(0x1234, []byte{expected})

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
	cpu.memory.Load(0x1234, []byte{expected})

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

	initialSP := cpu.GetSP()
	cpu.LoadProgram(encode([]Instruction{
		LoadHLSP{immediate: 5},
	}))
	cpu.Run()

	if actual := cpu.GetHL(); actual != initialSP+5 {
		t.Errorf("Expected %#X, got %#X\n", initialSP+5, actual)
	}
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
	cpu.SetSP(initial)
	cpu.LoadProgram(encode([]Instruction{
		StoreSP{immediate: 0x1234},
	}))
	cpu.Run()

	first := cpu.memory.Get(0x1234)
	second := cpu.memory.Get(0x1235)
	if first != 0xCD {
		t.Errorf("Expected %#X, got %#X\n", 0xCD, first)
	}
	if second != 0xFF {
		t.Errorf("Expected %#X, got %#X\n", 0xFF, second)
	}
}

func TestAdd(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: B, immediate: 0x10},
		MoveImmediate{dest: A, immediate: 0x1},
		Add{source: B},
	}))
	cpu.Run()

	if actual := cpu.Get(A); actual != 0x11 {
		t.Errorf("Expected 0x11, got %#X\n", actual)
	}
}

func TestAddImmediate(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: A, immediate: 0x1},
		AddImmediate{immediate: 0x10},
	}))
	cpu.Run()

	if actual := cpu.Get(A); actual != 0x11 {
		t.Errorf("Expected 0x11, got %#X\n", actual)
	}
}

func TestAddFlags(t *testing.T) {
	cpu := Init()

	cpu.LoadProgram(encode([]Instruction{
		MoveImmediate{dest: B, immediate: 0xFF},
		MoveImmediate{dest: A, immediate: 0x1},
		Add{source: B},
	}))
	cpu.Run()

	if actual := cpu.Get(A); actual != 0x0 {
		t.Errorf("Expected 0x0, got %#X\n", actual)
	}
	if flags := cpu.flags; flags != 0xB0 {
		t.Errorf("Expected 0xB0, got %#X\n", flags)
	}
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
		{instructions: []Instruction{LoadIndirect{dest: A, source: BC}}, expected: 2, message: "Load Pair BC"},
		{instructions: []Instruction{StoreIndirect{dest: BC, source: A}}, expected: 2, message: "Store Pair BC"},
		{instructions: []Instruction{LoadRelative{addressType: RelativeC}}, expected: 2, message: "Load Relative C"},
		{instructions: []Instruction{StoreRelative{addressType: RelativeC}}, expected: 2, message: "Store Relative C"},
		{instructions: []Instruction{LoadRelative{addressType: RelativeN}}, expected: 3, message: "Load Relative N"},
		{instructions: []Instruction{StoreRelative{addressType: RelativeN}}, expected: 3, message: "Store Relative N"},
		{instructions: []Instruction{LoadRelative{addressType: RelativeNN}}, expected: 4, message: "Load NN"},
		{instructions: []Instruction{StoreRelative{addressType: RelativeNN}}, expected: 4, message: "Store NN"},
		{instructions: []Instruction{LoadIncrement{}}, expected: 2, message: "Load increment"},
		{instructions: []Instruction{LoadDecrement{}}, expected: 2, message: "Load decrement"},
		{instructions: []Instruction{StoreIncrement{}}, expected: 2, message: "Store increment"},
		{instructions: []Instruction{StoreDecrement{}}, expected: 2, message: "Store decrement"},
		{instructions: []Instruction{LoadRegisterPairImmediate{dest: BC, immediate: 0x1234}}, expected: 3, message: "Load register pair immediate"},
		{instructions: []Instruction{HLtoSP{}}, expected: 2, message: "HL to SP"},
		{instructions: []Instruction{Push{source: BC}}, expected: 4, message: "Push"},
		{instructions: []Instruction{Pop{dest: BC}}, expected: 3, message: "Pop"},
		{instructions: []Instruction{LoadHLSP{immediate: 20}}, expected: 3, message: "Load HL SP"},
		{instructions: []Instruction{StoreSP{immediate: 0xDEAD}}, expected: 5, message: "Store SP"},
		{instructions: []Instruction{Add{source: B}}, expected: 1, message: "Add"},
		{instructions: []Instruction{AddImmediate{immediate: 0x12}}, expected: 2, message: "Add Immediate"},
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

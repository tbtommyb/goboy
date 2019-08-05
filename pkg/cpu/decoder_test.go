package cpu

import "testing"

var testCases = map[byte]Instruction{
	0xFF: InvalidInstruction{opcode: 0xFF},
	0x77: Move{source: A, dest: M},
	0x46: Move{source: M, dest: B},
	0x00: EmptyInstruction{},
	0x47: Move{source: A, dest: B},
	0x6:  MoveImmediate{dest: B},
	0x36: MoveImmediate{dest: M},
	0xA:  LoadIndirect{dest: A, source: BC},
	0x1A: LoadIndirect{dest: A, source: DE},
	0x2:  StoreIndirect{dest: BC, source: A},
	0x12: StoreIndirect{dest: DE, source: A},
	0xF2: LoadRelative{addressType: RelativeC},
	0xF0: LoadRelative{addressType: RelativeN},
	0xFA: LoadRelative{addressType: RelativeNN},
	0xE2: StoreRelative{addressType: RelativeC},
	0xE0: StoreRelative{addressType: RelativeN},
	0xEA: StoreRelative{addressType: RelativeNN},
	0x2A: LoadIncrement{},
	0x3A: LoadDecrement{},
	0x22: StoreIncrement{},
	0x32: StoreDecrement{},
	0x1:  LoadRegisterPairImmediate{dest: BC},
	0x11: LoadRegisterPairImmediate{dest: DE},
	0x21: LoadRegisterPairImmediate{dest: HL},
	0x31: LoadRegisterPairImmediate{dest: SP},
	0xF9: HLtoSP{},
	0xC5: Push{source: BC},
	0xD5: Push{source: DE},
	0xE5: Push{source: HL},
	0xF5: Push{source: AF},
	0xC1: Pop{dest: BC},
	0xD1: Pop{dest: DE},
	0xE1: Pop{dest: HL},
	0xF1: Pop{dest: AF},
	0xF8: LoadHLSP{},
	0x8:  StoreSP{},
	0x81: Add{source: C},
	0x8F: Add{source: A, withCarry: true},
	0x8E: Add{source: M, withCarry: true},
	0xC6: AddImmediate{},
	0xCE: AddImmediate{withCarry: true},
	0x91: Subtract{source: C},
	0x96: Subtract{source: M},
	0xD6: SubtractImmediate{},
	0x99: Subtract{source: C, withCarry: true},
	0x9E: Subtract{source: M, withCarry: true},
	0xDE: SubtractImmediate{withCarry: true},
}

func TestSimpleDecodes(t *testing.T) {
	for instruction, expected := range testCases {
		actual := Decode(instruction)
		if actual != expected {
			t.Errorf("Expected %#v, got %#v\n", expected, actual)
		}
	}
}

func TestSimpleOpcodes(t *testing.T) {
	for instruction, _ := range testCases {
		actual := Decode(instruction).Opcode()
		if actual[0] != instruction {
			t.Errorf("Expected %#v, got %#v\n", instruction, actual)
		}
	}
}

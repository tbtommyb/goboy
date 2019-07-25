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
	0xA:  MoveIndirect{dest: A, source: BC},
	0x1A: MoveIndirect{dest: A, source: DE},
	0x2:  MoveIndirect{dest: BC, source: A},
	0x12: MoveIndirect{dest: DE, source: A},
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
	0xD5: Push{source: PushDE},
	0xC5: Push{source: PushBC},
	0xE5: Push{source: PushHL},
	0xF5: Push{source: PushAF},
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

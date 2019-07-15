package cpu

import "testing"

func TestSimpleDecodes(t *testing.T) {
	testCases := map[byte]Instruction{
		0xFF: InvalidInstruction{opcode: 0xFF},
		0x77: Load{source: A, dest: M},
		0x46: Load{source: M, dest: B},
		0x00: EmptyInstruction{},
		0x47: Load{source: A, dest: B},
		0x6:  LoadImmediate{dest: B},
		0x36: LoadImmediate{dest: M},
		0xA:  LoadPair{dest: A, source: BC},
		0x1A: LoadPair{dest: A, source: DE},
		0x2:  LoadPair{dest: BC, source: A},
		0x12: LoadPair{dest: DE, source: A},
	}

	for instruction, expected := range testCases {
		actual := Decode(instruction)
		if actual != expected {
			t.Errorf("Expected %#v, got %#v\n", expected, actual)
		}
	}
}

// Not sure how useful these tests are
func TestSimpleOpcodes(t *testing.T) {
	testCases := []byte{0xFF, 0x77, 0x46, 0x80, 0x47, 0x6, 0x36, 0xA, 0x1A, 0x2, 0x12}

	for _, instruction := range testCases {
		actual := Decode(instruction).Opcode()
		if actual[0] != instruction {
			t.Errorf("Expected %#v, got %#v\n", instruction, actual)
		}
	}
}

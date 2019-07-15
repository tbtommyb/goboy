package cpu

import "testing"

func TestSimpleDecodes(t *testing.T) {
	testCases := map[byte]Instruction{
		0xFF: InvalidInstruction{opcode: 0xFF},
		0x77: StoreMemoryRegister{source: A},
		0x46: LoadRegisterMemory{dest: B},
		0x00: EmptyInstruction{},
		0x47: LoadRegister{source: A, dest: B},
		0x6:  LoadImmediate{dest: B},
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
	testCases := []byte{0xFF, 0x77, 0x46, 0x80, 0x47, 0x6}

	for _, instruction := range testCases {
		actual := Decode(instruction).Opcode()
		if actual[0] != instruction {
			t.Errorf("Expected %#v, got %#v\n", instruction, actual)
		}
	}
}

package cpu

import "testing"

// TODO: Decode requires access to memory so test ends up the same as cpu_test
func TestLoadImmediateOpcode(t *testing.T) {}
func TestLoadImmediateDecode(t *testing.T) {}

func TestSimpleDecodes(t *testing.T) {
	cpu := Init()

	testCases := map[byte]Instruction{
		0xFF: InvalidInstruction{opcode: 0xFF},
		0x77: StoreMemoryRegister{source: A},
		0x46: LoadRegisterMemory{dest: B},
		0x00: EmptyInstruction{opcode: 0},
		0x47: LoadRegister{source: A, dest: B},
	}

	for instruction, expected := range testCases {
		actual := cpu.Decode(instruction)
		if actual != expected {
			t.Errorf("Expected %#v, got %#v\n", expected, actual)
		}
	}
}

// Not sure how useful these tests are
func TestSimpleOpcodes(t *testing.T) {
	cpu := Init()

	testCases := []byte{0x47}

	for _, instruction := range testCases {
		actual := cpu.Decode(instruction).Opcode()
		if actual[0] != instruction {
			t.Errorf("Expected %#v, got %#v\n", instruction, actual)
		}
	}
}

package cpu

import "testing"

func TestLoadRegisterOpcode(t *testing.T) {
	cpu := Init()
	instr := cpu.Decode(0x47)
	if actual := instr.Opcode(); actual[0] != byte(0x47) {
		t.Errorf("Expected opcode 0x47, got %x", actual[0])
	}
}

func TestLoadRegisterDecode(t *testing.T) {
	cpu := Init()
	actual := cpu.Decode(0x47)
	expected := LoadRegister{source: A, dest: B}
	if actual != expected {
		t.Errorf("Expected LD B, A, got %#v", actual)
	}
}

// TODO: Decode requires access to memory so test ends up the same as cpu_test
func TestLoadImmediateOpcode(t *testing.T) {}
func TestLoadImmediateDecode(t *testing.T) {}

func TestInvalidOpcode(t *testing.T) {
	cpu := Init()
	actual := cpu.Decode(0xFF)
	expected := InvalidInstruction{opcode: 0xFF}
	if actual != expected {
		t.Errorf("Expected %#v, got %#v\n", expected, actual)
	}
}

func TestEmptyOpcode(t *testing.T) {
	cpu := Init()
	actual := cpu.Decode(0x00)
	expected := EmptyInstruction{opcode: 0}
	if actual != expected {
		t.Errorf("Expected %#v, got %#v\n", expected, actual)
	}
}

// TODO: deduplicate these tests
func TestLoadRegisterMemoryDecode(t *testing.T) {
	cpu := Init()
	actual := cpu.Decode(0x46)
	expected := LoadRegisterMemory{dest: B}
	if actual != expected {
		t.Errorf("Expected %#v, got %#v\n", expected, actual)
	}
}

func TestStoreMemoryRegisterDecode(t *testing.T) {
	cpu := Init()
	actual := cpu.Decode(0x77)
	expected := StoreMemoryRegister{source: A}
	if actual != expected {
		t.Errorf("Expected %#v, got %#v\n", expected, actual)
	}
}

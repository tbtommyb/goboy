package cpu

import "testing"

func TestSetAndGetRegister(t *testing.T) {
	cpu := Init()

	cpu.Set(A, 3)
	cpu.LoadProgram([]byte{0x47, 0x48}) // LD B, A then LD C, B. Need better system for this

	cpu.Run()
	if actual := cpu.memory[0x151]; actual != 0x48 {
		t.Fatalf("Expected 0x88, got %x", actual)
	}

	if regValue := cpu.Get(C); regValue != 3 {
		t.Fatalf("Expected 3, got %d", regValue)
	}
}

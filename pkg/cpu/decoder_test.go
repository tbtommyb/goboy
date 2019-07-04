package cpu

import "testing"

func TestLoadRegisterDecode(t *testing.T) {
	instr := Decode(0x47)
	switch actual := instr.(type) {
	case LoadRegisterInstr:
	default:
		t.Errorf("Expected LoadRegisterInstr, got %#v", actual)
	}
}

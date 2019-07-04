package cpu

import "testing"

func TestGetAndSetRegister(t *testing.T) {
	cpu := Init()

	cpu.Set(A, 3)

	cpu.Run(0x47) // LD B, A

	if actual := cpu.Get(B); actual != 3 {
		t.Fatalf("Exected B: 3, got B: %d", actual)
	}
}

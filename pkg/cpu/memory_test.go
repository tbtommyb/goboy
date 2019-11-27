package cpu

import "testing"

func TestRegisterWriteMasks(t *testing.T) {
	testCases := []struct {
		address         uint16
		expected, input byte
	}{
		{address: InterruptFlagAddress, input: 0x0, expected: 0xE0},
		{address: InterruptEnableAddress, input: 0xAA, expected: 0xAA},
		{address: TMAAddress, input: 0xFF, expected: 0xFF},
		{address: TMAAddress, input: 0xAA, expected: 0xAA},
		{address: TIMAAddress, input: 0xFF, expected: 0xFF},
		{address: TIMAAddress, input: 0xAA, expected: 0xAA},
		{address: TACAddress, input: 0x3, expected: 0xFB},
		{address: LCDCAddress, input: 0xAA, expected: 0xAA},
		{address: LYCAddress, input: 0xAA, expected: 0xAA},
		{address: STATAddress, input: 0x7F, expected: 0xFA},
		{address: ScrollXAddress, input: 0xAA, expected: 0xAA},
		{address: ScrollYAddress, input: 0xAA, expected: 0xAA},
		{address: WindowXAddress, input: 0xAA, expected: 0xAA},
		{address: WindowYAddress, input: 0xAA, expected: 0xAA},
	}

	for _, test := range testCases {
		cpu := createTestCPU()
		cpu.memory.set(test.address, test.input)

		if actual := cpu.memory.get(test.address); actual != test.expected {
			t.Errorf("Expected write to %x to return %x, got %x\n", test.address, test.expected, actual)
		}
	}
}

func TestDividerWrite(t *testing.T) {
	cpu := createTestCPU()
	cpu.internalTimer = 0xABCD

	cpu.memory.set(DIVAddress, 0x12)

	if actual := cpu.memory.get(DIVAddress); actual != 0 {
		t.Errorf("Expected write to DIV to reset DIV to 0, got %x\n", actual)
	}
	if actual := cpu.internalTimer; actual != 0 {
		t.Errorf("Expected write to DIV to reset internal clock to 0, got %x\n", actual)
	}
}

func TestLYReset(t *testing.T) {
	var initial byte = 0x12
	cpu := createTestCPU()
	cpu.setLY(initial)
	cpu.memory.set(LYAddress, 0x55)

	if actual := cpu.memory.get(LYAddress); actual != 0 {
		t.Errorf("Write to LY should reset, got %x\n", actual)
	}
}

func createTestCPU() *CPU {
	cpu := Init(false)
	cpu.setPC(0)
	return cpu
}

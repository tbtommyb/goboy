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

func TestMBC1(t *testing.T) {
	cpu := createTestCPU()
	cpu.memory.bankingEnabled = true
	cpu.memory.bankingMode = ROMBanking
	cpu.memory.rom = make([]byte, 0x200000)
	cpu.memory.mbc = MBC1

	ramEnableTestCases := []struct {
		address        uint16
		input          byte
		enableExpected bool
	}{
		{address: 0x0000, input: 0xAA, enableExpected: true},
		{address: 0x1000, input: 0x0A, enableExpected: true},
		{address: 0x1FFF, input: 0x12, enableExpected: false},
		{address: 0x0FFF, input: 0xA0, enableExpected: false},
	}

	for _, test := range ramEnableTestCases {
		cpu.memory.set(test.address, test.input)
		if actual := cpu.memory.get(test.address); actual != 0x0 {
			t.Errorf("%x should be read only. Got %x\n", test.address, actual)
		}
		if actual := cpu.memory.enableRam; actual != test.enableExpected {
			t.Errorf("RAM enable state is %t, expected %t\n", actual, test.enableExpected)
		}
	}

	romBankTestCases := []struct {
		lowerAddress, upperAddress uint16
		lowerVal, upperVal         byte
		expectedBank               uint
	}{
		{lowerAddress: 0x2000, lowerVal: 0x00, upperAddress: 0x4000, upperVal: 0x0, expectedBank: 0x1},
		{lowerAddress: 0x2400, lowerVal: 0x01, upperAddress: 0x4800, upperVal: 0x1, expectedBank: 0x21},
		{lowerAddress: 0x3000, lowerVal: 0x10, upperAddress: 0x5000, upperVal: 0x2, expectedBank: 0x50},
		{lowerAddress: 0x3600, lowerVal: 0xf, upperAddress: 0x5FFF, upperVal: 0x3, expectedBank: 0x6F},
	}

	for _, test := range romBankTestCases {
		cpu.memory.set(test.lowerAddress, test.lowerVal)
		cpu.memory.set(test.upperAddress, test.upperVal)
		if actual := cpu.memory.get(test.lowerAddress); actual != 0x0 {
			t.Errorf("%x should be read only. Got %x\n", test.lowerAddress, actual)
		}
		if actual := cpu.memory.get(test.upperAddress); actual != 0x0 {
			t.Errorf("%x should be read only. Got %x\n", test.upperAddress, actual)
		}
		if actual := cpu.currentROMBank; actual != test.expectedBank {
			t.Errorf("Expected ROM bank %x, got %x\n", test.expectedBank, actual)
		}
	}

	romRamSelectTestCases := []struct {
		address      uint16
		value        byte
		expectedMode BankingMode
	}{
		{address: 0x6000, value: 0x0, expectedMode: ROMBanking},
		{address: 0x6F00, value: 0x1, expectedMode: RAMBanking},
		{address: 0x7000, value: 0x20, expectedMode: ROMBanking},
		{address: 0x7FFF, value: 0x21, expectedMode: RAMBanking},
	}

	for _, test := range romRamSelectTestCases {
		cpu.memory.set(test.address, test.value)
		if actual := cpu.memory.get(test.address); actual != 0x0 {
			t.Errorf("%x should be read only. Got %x\n", test.address, actual)
		}
		if actual := cpu.memory.bankingMode; actual != test.expectedMode {
			t.Errorf("Expected mode %x, got %x\n", test.expectedMode, actual)
		}
	}
}

func TestRomModeRamBankReset(t *testing.T) {
	cpu := createTestCPU()
	cpu.memory.bankingEnabled = true
	cpu.memory.bankingMode = ROMBanking
	cpu.memory.rom = make([]byte, 0x200000)
	cpu.memory.mbc = MBC1

	cpu.memory.set(0x3600, 0x1F)
	cpu.memory.set(0x5FFF, 0x3)

	if actual := cpu.currentROMBank; actual != 0x7F {
		t.Errorf("Expected ROM bank %x, got %x\n", 0x7F, actual)
	}
	if actual := cpu.memory.currentRAMBank; actual != 0 {
		t.Errorf("Expected RAM bank %x, got %x\n", 0, actual)
	}

	cpu.memory.set(0x6FFF, 0x3)

	if actual := cpu.memory.bankingMode; actual != RAMBanking {
		t.Errorf("Expected RAM banking, got %x\n", actual)
	}
	if actual := cpu.memory.currentRAMBank; actual != 3 {
		t.Errorf("Expected RAM bank %x, got %x\n", 3, actual)
	}
	if actual := cpu.currentROMBank; actual != 0x1F {
		t.Errorf("Expected ROM bank %x, got %x\n", 0x1F, actual)
	}

	cpu.memory.set(0x6FFF, 0x40)

	if actual := cpu.memory.bankingMode; actual != ROMBanking {
		t.Errorf("Expected ROM banking, got %x\n", actual)
	}
	if actual := cpu.memory.currentRAMBank; actual != 0 {
		t.Errorf("Expected RAM bank %x, got %x\n", 0, actual)
	}
	if actual := cpu.currentROMBank; actual != 0x7F {
		t.Errorf("Expected ROM bank %x, got %x\n", 0x7F, actual)
	}

	cpu.memory.set(0x6FFF, 0x3)

	if actual := cpu.memory.bankingMode; actual != RAMBanking {
		t.Errorf("Expected RAM banking, got %x\n", actual)
	}
	if actual := cpu.memory.currentRAMBank; actual != 3 {
		t.Errorf("Expected RAM bank %x, got %x\n", 3, actual)
	}
	if actual := cpu.currentROMBank; actual != 0x1F {
		t.Errorf("Expected ROM bank %x, got %x\n", 0x1F, actual)
	}

}
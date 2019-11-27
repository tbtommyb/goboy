package memory

type BankingMode byte

const (
	ROMBanking BankingMode = 0x0
	RAMBanking             = 0x1
)

type MBC byte

const (
	MBC1 MBC = 1
	MBC2     = 2
)

func (m *Memory) handleBanking(address uint16, value byte) {
	switch {
	case address < RAMEnableLimit:
		m.ramBankEnable(address, value)
	case address >= RAMEnableLimit && address < ROMBankNumberLimit:
		m.setLowerROMBankBits(address, value)
	case address >= ROMBankNumberLimit && address < RAMBankNumberLimit:
		switch m.mbc {
		case MBC1:
			valueBits := uint(value & 0x3)
			if m.bankingMode == ROMBanking {
				m.setUpperROMBankBits(valueBits)
			} else {
				m.currentRAMBank = uint(valueBits)
			}
		}
	case address >= RAMBankNumberLimit && address < ROMRAMModeSelectLimit:
		switch m.mbc {
		case MBC1:
			newMode := BankingMode(value & 0x1)
			if newMode == RAMBanking && m.bankingMode != RAMBanking {
				m.currentRAMBank = (m.currentROMBank & 0x60) >> 5
				m.currentROMBank = (m.currentROMBank & 0x1F)
			} else {
				m.currentROMBank = (m.currentRAMBank << 5) | m.currentROMBank
				m.currentRAMBank = 0
			}
			m.bankingMode = newMode
		}
	}
}

func (m *Memory) ramBankEnable(address uint16, value byte) {
	if m.mbc == MBC2 && address&0x100 > 0 {
		return
	}
	m.enableRam = (value & 0xF) == 0xA
}

func (m *Memory) setLowerROMBankBits(address uint16, value byte) {
	switch m.mbc {
	case MBC1:
		bankNum := uint(value & 0x1F)
		if bankNum == 0 {
			bankNum = 1
		}
		bankNum = (m.currentROMBank &^ 0x1F) | bankNum
		m.currentROMBank = bankNum
	case MBC2:
		if address&0x100 == 0 {
			return
		}
		bankNum := uint(value & 0x1F)
		if bankNum == 0 {
			bankNum = 1
		}
		m.currentROMBank = bankNum
	}
}

func (m *Memory) setUpperROMBankBits(bottomBits uint) {
	remainingBits := m.currentROMBank & 0x1F

	m.currentROMBank = (bottomBits << 5) | remainingBits
}

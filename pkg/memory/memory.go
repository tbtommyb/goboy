package memory

import (
	"fmt"

	c "github.com/tbtommyb/goboy/pkg/constants"
)

type CPUInterface interface {
	WriteVRAM(address uint16, value byte)
	ReadVRAM(address uint16) byte
	WriteOAM(address uint16, value byte)
	ReadOAM(address uint16) byte
	WriteJoypad(value byte)
	ReadJoypad() byte
	WriteIO(address uint16, value byte)
	ReadIO(address uint16) byte
	ResetInternalTimer()
	GetInternalTimer() uint16
	ResetCyclesForTimerTick()
	BIOSLoaded() bool
	RunFor(cycles uint)
}

type Memory struct {
	rom             []byte
	bios            [0x100]byte
	eram            [0x8000]byte
	wram            [0x2000]byte
	hram            [0x7F]byte
	interruptEnable byte
	statMode        byte
	cpu             CPUInterface
	currentRAMBank  uint
	currentROMBank  uint
	enableRam       bool
	ramAvailable    bool
	ramBankSize     uint
	bankingMode     BankingMode
	bankingEnabled  bool
	mbc             MBC
}

const CartridgeTypeAddress = 0x147
const ROMSizeAddress = 0x148
const RAMSizeAddress = 0x149

const RAMEnableLimit = 0x2000
const ROMBankNumberLimit = 0x4000
const RAMBankNumberLimit = 0x6000
const ROMRAMModeSelectLimit = 0x8000
const ProgramStartAddress = 0x100
const StackStartAddress = 0xFFFE
const OAMStart = 0xFE00
const TACMask = 0x7
const DMAAddress = 0xFF46
const ROMBank00Limit = 0x4000
const ROMBankLimit = 0x8000
const ROMBankSize = 0x4000
const CartRAMStart = 0xA000
const CartRAMEnd = 0xBFFF

func Init(cpu CPUInterface) *Memory {
	return &Memory{
		bios:           [0x100]byte{},
		eram:           [0x8000]byte{},
		wram:           [0x2000]byte{},
		hram:           [0x7F]byte{},
		currentRAMBank: 0,
		currentROMBank: 1,
		cpu:            cpu,
	}
}

func (m *Memory) load(start uint, data []byte) {
	for i := 0; i < len(data); i++ {
		m.rom[start+uint(i)] = data[i]
	}
}

func (m *Memory) LoadBIOS(program []byte) {
	for i := 0; i < len(program); i++ {
		m.bios[uint(i)] = program[i]
	}
}

func (m *Memory) Set(address uint16, value byte) {
	switch {
	case address < ROMBankLimit:
		if m.bankingEnabled {
			m.handleBanking(address, value)
		}
	case address >= 0x8000 && address <= 0x9FFF:
		// video ram
		m.cpu.WriteVRAM(address, value)
	case address >= CartRAMStart && address <= CartRAMEnd:
		offset := uint(address - CartRAMStart)
		// blargg oam_bug test output
		if offset == 0x00 && m.eram[1] == 0xDE && m.eram[2] == 0xB0 && m.eram[3] == 0x61 {
			if value != 0x80 {
				for i := 4; m.eram[i] != 0x0; i++ {
					fmt.Printf("%s", string(m.eram[i]))
				}
			}
		}
		if !m.enableRam {
			// TODO: check if this should exist outside of TEST mode
			m.eram[offset] = value
			return
		}
		switch m.mbc {
		case MBC1:
			m.eram[offset+(m.currentRAMBank*m.ramBankSize)] = value
		case MBC2:
			m.eram[offset] = value & 0xF // lower 4 bits only
		}
	case address >= 0xC000 && address <= 0xDFFF:
		m.wram[address-0xC000] = value
	case address >= 0xE000 && address <= 0xFDFF:
		// shadow wram
		m.Set(address-0x2000, value)
	case address >= 0xFE00 && address <= 0xFE9F:
		// sprites
		m.cpu.WriteOAM(address, value)
	case address >= 0xFEA0 && address <= 0xFEFF:
		// unusable
	case address >= 0xFF00 && address <= 0xFF7F:
		// memory mapped IO
		if address == 0xFF01 {
			fmt.Printf("%c", value)
		} else if address == c.JoypadRegisterAddress {
			m.cpu.WriteJoypad(value)
		} else if address == c.LYAddress {
			// Reset if game writes to LY
			m.cpu.WriteIO(address, 0)
		} else if address == c.DIVAddress {
			m.cpu.WriteIO(address, 0)
			m.cpu.ResetInternalTimer()
		} else if address == DMAAddress {
			// DMA
			m.performDMA(uint16(value) << 8)
		} else if address == c.TACAddress {
			newVal := value & TACMask
			oldVal := m.cpu.ReadIO(address) & TACMask
			m.cpu.WriteIO(address, 0xF8|newVal)
			if newVal != oldVal {
				m.cpu.ResetCyclesForTimerTick()
			}
		} else if address == 0xFF0A {
			m.cpu.WriteIO(address, 0)
		} else if address == c.STATAddress {
			readOnlyBits := m.cpu.ReadIO(address) & 7
			m.cpu.WriteIO(address, (value&0xF8)|readOnlyBits|0x80)
		} else if address == c.InterruptFlagAddress {
			m.cpu.WriteIO(address, 0xE0|(value&0x1F))
		} else {
			m.cpu.WriteIO(address, value)
		}
	case address >= 0xFF80 && address <= 0xFFFE:
		m.hram[address-0xFF80] = value
	case address == c.InterruptEnableAddress:
		m.interruptEnable = value
	}
}

func (m *Memory) Get(address uint16) byte {
	switch {
	case address < 0x100:
		// TODO: find neater solution
		if m.cpu.BIOSLoaded() {
			return m.bios[address]
		} else {
			return m.rom[address]
		}
	case address < ROMBank00Limit:
		return m.rom[address]
	case address >= ROMBank00Limit && address < ROMBankLimit:
		if !m.bankingEnabled {
			return m.rom[address]
		}
		offset := uint(address - ROMBank00Limit)
		return m.rom[offset+(m.currentROMBank*ROMBankSize)]
	case address >= ROMBankLimit && address <= 0x9FFF:
		// video ram
		return m.cpu.ReadVRAM(address)
	case address >= CartRAMStart && address <= CartRAMEnd:
		// cart ram
		if !m.enableRam || !m.ramAvailable {
			return 0xFF
		}
		offset := uint(address - CartRAMStart)
		return m.eram[offset+(m.currentRAMBank*m.ramBankSize)]
	case address >= 0xC000 && address <= 0xDFFF:
		return m.wram[address-0xC000]
	case address >= 0xE000 && address <= 0xFDFF:
		// shadow wram
		return m.Get(address - 0x2000)
	case address >= 0xFE00 && address <= 0xFE9F:
		// sprites
		return m.cpu.ReadOAM(address)
	case address >= 0xFEA0 && address <= 0xFEFF:
		// unused space
		return 0xFF
	case address >= 0xFF00 && address <= 0xFF7F:
		// memory mapped IO
		if address == c.DIVAddress {
			return byte(m.cpu.GetInternalTimer() >> 8)
		} else if address == c.JoypadRegisterAddress {
			return m.cpu.ReadJoypad()
		} else if address == c.InterruptFlagAddress {
			return 0xE0 | (m.cpu.ReadIO(address) & 0x1F)
		}
		return m.cpu.ReadIO(address)
	case address >= 0xFF80 && address <= 0xFFFE:
		// if address == 0xff85 {
		// 	return 1
		// }
		return m.hram[address-0xFF80]
	case address == c.InterruptEnableAddress:
		return m.interruptEnable
	default:
		panic(fmt.Sprintf("%x\n", address))
	}
}

func (m *Memory) performDMA(address uint16) {
	m.cpu.RunFor(4)
	for i := uint16(0); i < 0xA0; i++ {
		m.cpu.RunFor(4)
		m.Set(0xFE00+i, m.Get(address+i))
	}
}

func (m *Memory) LoadROM(program []byte) {
	cartridgeType := program[CartridgeTypeAddress]
	romSize := program[ROMSizeAddress]
	ramSize := program[RAMSizeAddress]
	if romSize > 0 {
		m.bankingEnabled = true
		m.bankingMode = ROMBanking
	}

	switch cartridgeType {
	case 1:
		m.mbc = MBC1
	case 2, 3:
		m.mbc = MBC1
		m.ramAvailable = true
	case 5, 6:
		m.mbc = MBC2
	case 15, 16, 17, 18, 19:
		panic("MBC3 not implemented")
	}

	switch romSize {
	case 0x0:
		m.rom = make([]byte, 0x8000)
	case 0x1:
		m.rom = make([]byte, 0x10000)
	case 0x2:
		m.rom = make([]byte, 0x20000)
	case 0x3:
		m.rom = make([]byte, 0x40000)
	case 0x4:
		m.rom = make([]byte, 0x80000)
	case 0x5:
		m.rom = make([]byte, 0x100000)
	case 0x6:
		m.rom = make([]byte, 0x200000)
	case 0x7:
		m.rom = make([]byte, 0x400000)
	case 0x8:
		m.rom = make([]byte, 0x800000)
	}

	switch ramSize {
	case 0x0:
		m.ramBankSize = 0x0
	case 0x1:
		m.ramBankSize = 0x800
	case 0x2:
		m.ramBankSize = 0x2000
	case 0x3:
		m.ramBankSize = 0x8000
	}
	m.load(0, program)
}

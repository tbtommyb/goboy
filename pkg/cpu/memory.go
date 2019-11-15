package cpu

import (
	"fmt"

	"github.com/tbtommyb/goboy/pkg/registers"
	"github.com/tbtommyb/goboy/pkg/utils"
)

type Memory struct {
	rom             []byte
	bios            [0x100]byte
	vram            [0x2000]byte
	eram            [0x8000]byte
	wram            [0x2000]byte
	sram            [0x100]byte
	ioram           [0x100]byte
	hram            [0x7F]byte
	interruptEnable byte
	statMode        byte
	cpu             *CPU
	// ramBanks        [0x8000]byte
	currentRAMBank uint16
	enableRam      bool
	bankingMode    BankingMode
	bankingEnabled bool
	mbc            MBC
}

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

const CartridgeTypeAddress = 0x147
const ROMSizeAddress = 0x148

const RAMEnableLimit = 0x2000
const ROMBankNumberLimit = 0x4000
const RAMBankNumberLimit = 0x6000
const ROMRAMModeSelectLimit = 0x8000
const ProgramStartAddress = 0x100
const StackStartAddress = 0xFFFE
const SpriteDataStartAddress = 0x8000
const OAMStart = 0xFE00
const TACMask = 0x7
const DMAAddress = 0xFF46
const ROMBank00Limit = 0x4000
const ROMBankLimit = 0x8000
const ROMBankSize = 0x4000
const CartRAMStart = 0xA000
const CartRAMEnd = 0xBFFF
const RAMBankSize = 0x2000 // TODO: double-check this

func InitMemory(cpu *CPU) *Memory {
	return &Memory{
		bios:           [0x100]byte{},
		vram:           [0x2000]byte{},
		eram:           [0x8000]byte{},
		wram:           [0x2000]byte{},
		ioram:          [0x100]byte{},
		hram:           [0x7F]byte{},
		currentRAMBank: 0,
		cpu:            cpu,
	}
}

func (m *Memory) load(start uint, data []byte) {
	for i := 0; i < len(data); i++ {
		m.rom[start+uint(i)] = data[i]
	}
}

func (m *Memory) set(address uint16, value byte) {
	switch {
	case address < ROMBankLimit:
		if m.bankingEnabled {
			m.handleBanking(address, value)
		}
	case address >= 0x8000 && address <= 0x9FFF:
		// video ram
		m.vram[address-0x8000] = value
	case address >= CartRAMStart && address <= CartRAMEnd:
		offset := address - CartRAMStart
		if m.enableRam {
			switch m.mbc {
			case MBC1:
				m.eram[offset+(m.currentRAMBank*RAMBankSize)] = value
			case MBC2:
				m.eram[offset] = value & 0xF // lower 4 bits only
			}
		}
	case address >= 0xC000 && address <= 0xDFFF:
		m.wram[address-0xC000] = value
	case address >= 0xE000 && address <= 0xFDFF:
		// shadow wram
	case address >= 0xFE00 && address <= 0xFE9F:
		// sprites
		m.cpu.gpu.writeOAM(address, value)
	case address >= 0xFEA0 && address <= 0xFEFF:
		// unusable
	case address >= 0xFF00 && address <= 0xFF7F:
		// memory mapped IO
		localAddress := address - 0xFF00
		if address == 0xFF01 {
			// fmt.Printf("%c", value)
		} else if address == JoypadRegisterAddress {
			m.cpu.setJoypadSelection(value)
		} else if address == LYAddress {
			// Reset if game writes to LY
			m.ioram[localAddress] = 0
		} else if address == DIVAddress {
			m.ioram[localAddress] = 0
			m.cpu.internalTimer = 0
		} else if address == DMAAddress {
			// DMA
			m.performDMA(uint16(value) << 8)
		} else if address == TACAddress {
			newVal := value & TACMask
			oldVal := m.ioram[localAddress] & TACMask
			m.ioram[localAddress] = newVal
			if newVal != oldVal {
				m.cpu.resetCyclesForCurrentTick()
			}
		} else if address == 0xFF0A {
			m.ioram[localAddress] = 0
		} else if address == STATAddress {
			readOnlyBits := m.ioram[localAddress] & 7
			m.ioram[localAddress] = (value & 0xF8) | readOnlyBits
		} else if address == InterruptFlagAddress {
			m.ioram[localAddress] = value & 0x1F
		} else {
			m.ioram[localAddress] = value
		}
	case address >= 0xFF80 && address <= 0xFFFE:
		m.hram[address-0xFF80] = value
	case address == InterruptEnableAddress:
		m.interruptEnable = value & 0x1f
	}
}

func (m *Memory) get(address uint16) byte {
	switch {
	case address < 0x100:
		// TODO: find neater solution
		if m.cpu.loadBIOS {
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
		offset := address - ROMBank00Limit
		return m.rom[offset+(m.cpu.currentROMBank*ROMBankSize)]
	case address >= ROMBankLimit && address <= 0x9FFF:
		// video ram
		return m.vram[address-0x8000]
	case address >= CartRAMStart && address <= CartRAMEnd:
		// cart ram
		offset := address - CartRAMStart
		return m.eram[offset+(m.currentRAMBank*RAMBankSize)]
	case address >= 0xC000 && address <= 0xDFFF:
		return m.wram[address-0xC000]
	case address >= 0xE000 && address <= 0xFDFF:
		// shadow wram
		return 0x00
	case address >= 0xFE00 && address <= 0xFE9F:
		// sprites
		val := m.cpu.gpu.readOAM(address)
		return val
	case address >= 0xFEA0 && address <= 0xFEFF:
		// unused space
		return 0xFF
	case address >= 0xFF00 && address <= 0xFF7F:
		// memory mapped IO
		if address == DIVAddress {
			return byte(m.cpu.internalTimer >> 8)
		} else if address == JoypadRegisterAddress {
			return m.cpu.getJoypadState()
		} else if address == InterruptFlagAddress {
			return m.ioram[address-0xFF00] & 0x1F
		}
		return m.ioram[address-0xFF00]
	case address >= 0xFF80 && address <= 0xFFFE:
		// if address == 0xff85 {
		// 	return 1
		// }
		return m.hram[address-0xFF80]
	case address == InterruptEnableAddress:
		return m.interruptEnable & 0x1F
	default:
		panic(fmt.Sprintf("%x\n", address))
	}
}

func (m *Memory) performDMA(address uint16) {
	for i := uint16(0); i < 0xA0; i++ {
		m.set(0xFE00+i, m.get(address+i))
	}
}

func (m *Memory) handleBanking(address uint16, value byte) {
	switch {
	case address < RAMEnableLimit:
		m.ramBankEnable(address, value)
	case address >= RAMEnableLimit && address < ROMBankNumberLimit:
		m.setLowerROMBankBits(address, value)
	case address >= ROMBankNumberLimit && address < RAMBankNumberLimit:
		switch m.mbc {
		case MBC1:
			if m.bankingMode == ROMBanking {
				m.setUpperROMBankBits(value)
			} else {
				m.currentRAMBank = uint16(value & 0x3)
			}
		}
	case address >= RAMBankNumberLimit && address < ROMRAMModeSelectLimit:
		switch m.mbc {
		case MBC1:
			m.bankingMode = BankingMode(value & 0x1)
			if m.bankingMode == ROMBanking {
				m.currentRAMBank = 0
			}
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
		lowerFiveBits := uint16(value & 0x1F)
		topThreeBits := m.cpu.currentROMBank & 0xE0
		m.cpu.currentROMBank = topThreeBits | lowerFiveBits
	case MBC2:
		if address&0x100 == 0 {
			return
		}
		m.cpu.currentROMBank = uint16(value & 0xF)
	}
	if m.cpu.currentROMBank == 0 {
		m.cpu.currentROMBank++
	}
}

func (m *Memory) setUpperROMBankBits(value byte) {
	bitsFiveAndSix := uint16(value & 0x60)
	remainingBits := m.cpu.currentROMBank & 0x9F

	m.cpu.currentROMBank = remainingBits | bitsFiveAndSix
}

func (cpu *CPU) readMem(address uint16) byte {
	cpu.incrementCycles()
	return cpu.memory.get(address)
}

func (cpu *CPU) LoadBIOS(program []byte) {
	for i := 0; i < len(program); i++ {
		cpu.memory.bios[uint(i)] = program[i]
	}
	cpu.loadBIOS = true
}

func (cpu *CPU) LoadROM(program []byte) {
	cartridgeType := program[CartridgeTypeAddress]
	romSize := program[ROMSizeAddress]
	if romSize > 0 {
		cpu.memory.bankingEnabled = true
		cpu.memory.bankingMode = ROMBanking
	}
	switch romSize {
	case 0x0:
		cpu.memory.rom = make([]byte, 0x8000)
	case 0x1:
		cpu.memory.rom = make([]byte, 0x10000)
	case 0x2:
		cpu.memory.rom = make([]byte, 0x20000)
	case 0x3:
		cpu.memory.rom = make([]byte, 0x40000)
	case 0x4:
		cpu.memory.rom = make([]byte, 0x80000)
	case 0x5:
		cpu.memory.rom = make([]byte, 0x100000)
	case 0x6:
		cpu.memory.rom = make([]byte, 0x200000)
	case 0x7:
		cpu.memory.rom = make([]byte, 0x400000)
	case 0x8:
		cpu.memory.rom = make([]byte, 0x800000)
	}
	cpu.memory.load(0, program)
	switch cartridgeType {
	case 1, 2, 3:
		cpu.memory.mbc = MBC1
	case 5, 6:
		cpu.memory.mbc = MBC2
	case 15, 16, 17, 18, 19:
		panic("MBC3 not implemented")
	}
}

func (cpu *CPU) WriteMem(address uint16, value byte) {
	cpu.incrementCycles()
	cpu.memory.set(address, value)
}

func (cpu *CPU) GetMem(r registers.Pair) byte {
	switch r {
	case registers.BC:
		return cpu.readMem(cpu.GetBC())
	case registers.DE:
		return cpu.readMem(cpu.GetDE())
	case registers.HL:
		return cpu.readMem(cpu.GetHL())
	case registers.SP:
		return cpu.readMem(cpu.GetSP())
	default:
		panic(fmt.Sprintf("GetMem: Invalid register %x", r))
	}
}

func (cpu *CPU) SetMem(r registers.Pair, val byte) byte {
	switch r {
	case registers.BC:
		cpu.WriteMem(cpu.GetBC(), val)
	case registers.DE:
		cpu.WriteMem(cpu.GetDE(), val)
	case registers.HL:
		cpu.WriteMem(cpu.GetHL(), val)
	case registers.SP:
		cpu.WriteMem(cpu.GetSP(), val)
	default:
		panic(fmt.Sprintf("SetMem: Invalid register %x", r))
	}
	return val
}

func (m *Memory) setBitAt(address uint16, bitNumber, bitValue byte) {
	m.set(address, utils.SetBit(bitNumber, m.get(address), bitValue))
}

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
	currentRAMBank  uint
	currentROMBank  uint
	enableRam       bool
	ramAvailable    bool
	ramBankSize     uint
	bankingMode     BankingMode
	bankingEnabled  bool
	mbc             MBC
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
const RAMSizeAddress = 0x149

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

func InitMemory(cpu *CPU) *Memory {
	return &Memory{
		bios:           [0x100]byte{},
		vram:           [0x2000]byte{},
		eram:           [0x8000]byte{},
		wram:           [0x2000]byte{},
		ioram:          [0x100]byte{},
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
			m.cpu.internalTimer = 0
		} else if address == DMAAddress {
			// DMA
			m.performDMA(uint16(value) << 8)
		} else if address == TACAddress {
			newVal := value & TACMask
			oldVal := m.ioram[localAddress] & TACMask
			m.ioram[localAddress] = 0xF8 | newVal
			if newVal != oldVal {
				m.cpu.resetCyclesForCurrentTick()
			}
		} else if address == 0xFF0A {
			m.ioram[localAddress] = 0
		} else if address == STATAddress {
			readOnlyBits := m.ioram[localAddress] & 7
			m.ioram[localAddress] = (value & 0xF8) | readOnlyBits | 0x80
		} else if address == InterruptFlagAddress {
			m.ioram[localAddress] = 0xE0 | (value & 0x1F)
		} else {
			m.ioram[localAddress] = value
		}
	case address >= 0xFF80 && address <= 0xFFFE:
		m.hram[address-0xFF80] = value
	case address == InterruptEnableAddress:
		m.interruptEnable = value
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
		offset := uint(address - ROMBank00Limit)
		return m.rom[offset+(m.currentROMBank*ROMBankSize)]
	case address >= ROMBankLimit && address <= 0x9FFF:
		// video ram
		return m.vram[address-0x8000]
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
			return m.ioram[address-0xFF00]
		}
		return m.ioram[address-0xFF00]
	case address >= 0xFF80 && address <= 0xFFFE:
		// if address == 0xff85 {
		// 	return 1
		// }
		return m.hram[address-0xFF80]
	case address == InterruptEnableAddress:
		return m.interruptEnable
	default:
		panic(fmt.Sprintf("%x\n", address))
	}
}

func (m *Memory) performDMA(address uint16) {
	m.cpu.RunFor(4)
	for i := uint16(0); i < 0xA0; i++ {
		m.cpu.RunFor(4)
		m.set(0xFE00+i, m.get(address+i))
	}
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
	ramSize := program[RAMSizeAddress]
	if romSize > 0 {
		cpu.memory.bankingEnabled = true
		cpu.memory.bankingMode = ROMBanking
	}

	switch cartridgeType {
	case 1:
		cpu.memory.mbc = MBC1
	case 2, 3:
		cpu.memory.mbc = MBC1
		cpu.memory.ramAvailable = true
	case 5, 6:
		cpu.memory.mbc = MBC2
	case 15, 16, 17, 18, 19:
		panic("MBC3 not implemented")
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

	switch ramSize {
	case 0x0:
		cpu.memory.ramBankSize = 0x0
	case 0x1:
		cpu.memory.ramBankSize = 0x800
	case 0x2:
		cpu.memory.ramBankSize = 0x2000
	case 0x3:
		cpu.memory.ramBankSize = 0x8000
	}
	cpu.memory.load(0, program)
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

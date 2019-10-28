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
	eram            [0x2000]byte
	wram            [0x2000]byte
	sram            [0x100]byte
	ioram           [0x100]byte
	hram            [0x7F]byte
	interruptEnable byte
	statMode        byte
	cpu             *CPU
	ramBanks        [0x8000]byte
	currentRamBank  byte
	enableRam       bool
	romBanking      bool
}

const ProgramStartAddress = 0x100
const StackStartAddress = 0xFFFE
const SpriteDataStartAddress = 0x8000
const OAMStart = 0xFE00
const TACMask = 0x7

func (m *Memory) load(start uint, data []byte) {
	for i := 0; i < len(data); i++ {
		m.rom[start+uint(i)] = data[i]
	}
}

func (m *Memory) set(address uint16, value byte) {
	switch {
	case address < 0x8000:
		if !m.romBanking {
			m.rom[address] = value
		}
		m.handleBanking(address, value)
	case address >= 0x8000 && address <= 0x9FFF:
		// video ram
		m.vram[address-0x8000] = value
	case address >= 0xA000 && address <= 0xBFFF:
		if m.enableRam {
			newAddress := address - 0xA000
			m.ramBanks[newAddress+(uint16(m.currentRamBank)*0x2000)] = value
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
		if address == 0xFF01 {
			// fmt.Printf("%c", value)
		} else if address == LYAddress {
			// Reset if game writes to LY
			m.ioram[address-0xFF00] = 0
		} else if address == DIVAddress {
			m.ioram[address-0xFF00] = 0
			m.cpu.internalTimer = 0
		} else if address == 0xFF46 {
			// DMA
			m.performDMA(uint16(value) << 8)
		} else if address == TACAddress {
			newVal := value & TACMask
			oldVal := m.ioram[address-0xFF00] & TACMask
			m.ioram[address-0xFF00] = newVal
			if newVal != oldVal {
				m.cpu.resetCyclesForCurrentTick()
			}
		} else if address == 0xFF0A {
			m.ioram[address-0xFF00] = 0
		} else {
			m.ioram[address-0xFF00] = value
		}
	case address >= 0xFF80 && address <= 0xFFFE:
		m.hram[address-0xFF80] = value
	case address == 0xFFFF:
		m.interruptEnable = value
	}
}

func (m *Memory) setBitAt(address uint16, bitNumber, bitValue byte) {
	m.set(address, utils.SetBit(bitNumber, m.get(address), bitValue))
}

func (m *Memory) performDMA(address uint16) {
	for i := uint16(0); i < 0xA0; i++ {
		m.set(0xFE00+i, m.get(address+i))
	}
}
func (m *Memory) get(address uint16) byte {
	switch {
	case address < 0x100:
		if m.cpu.loadBIOS {
			return m.bios[address]
		} else {
			return m.rom[address]
		}
	case address < 0x4000:
		return m.rom[address]
	case address >= 0x4000 && address <= 0x7FFF:
		if !m.romBanking {
			return m.rom[address]
		}
		newAddress := uint16(address - 0x4000)
		value := m.rom[newAddress+uint16(m.cpu.currentROMBank*0x4000)]
		return value
	case address >= 0x8000 && address <= 0x9FFF:
		// video ram
		return m.vram[address-0x8000]
	case address >= 0xA000 && address <= 0xBFFF:
		// cart ram
		newAddress := address - 0xA000
		return m.eram[newAddress+(uint16(m.currentRamBank)*0x2000)]
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
		// TODO: maybe map to memory instead
		if address == JoypadRegisterAddress {
			return m.cpu.getJoypadState()
		} else if address == DIVAddress {
			return byte(m.cpu.internalTimer >> 8)
		}

		return m.ioram[address-0xFF00]
	case address >= 0xFF80 && address <= 0xFFFE:
		// if address == 0xff85 {
		// 	return 1
		// }
		return m.hram[address-0xFF80]
	case address == 0xFFFF:
		return m.interruptEnable
	default:
		panic(fmt.Sprintf("%x\n", address))
	}
}
func (m *Memory) handleBanking(address uint16, data byte) {
	switch {
	case address < 0x2000:
		if m.cpu.mbc1 || m.cpu.mbc2 {
			m.ramBankEnable(address, data)
		}
	case address >= 0x2000 && address < 0x4000:
		if m.cpu.mbc1 || m.cpu.mbc2 {
			m.changeLowROMBank(data)
		}
	case address >= 0x4000 && address < 0x6000:
		if m.cpu.mbc1 {
			if m.romBanking {
				m.changeHiROMBank(data)
			} else {
				m.ramBankChange(data)
			}
		}
	case address >= 0x6000 && address < 0x8000:
		if m.cpu.mbc1 {
			m.changeROMRAMMode(data)
		}
	}
}

func (m *Memory) ramBankEnable(address uint16, data byte) {
	if m.cpu.mbc2 {
		// TODO: clarify
		if utils.IsSet(4, byte(address)) {
			return
		}
	}
	testData := data & 0xf
	if testData == 0xA {
		m.enableRam = true
	} else if testData == 0x0 {
		m.enableRam = false
	}
}

func (m *Memory) changeLowROMBank(data byte) {
	if m.cpu.mbc2 {
		m.cpu.currentROMBank = uint16(data & 0xf)
		if m.cpu.currentROMBank == 0 {
			m.cpu.currentROMBank++
		}
		return
	}
	lower := data & 31
	m.cpu.currentROMBank &= 224
	m.cpu.currentROMBank |= uint16(lower)
	if m.cpu.currentROMBank == 0 {
		m.cpu.currentROMBank++
	}
}

func (m *Memory) changeHiROMBank(data byte) {
	m.cpu.currentROMBank &= 31

	data &= 224
	m.cpu.currentROMBank |= uint16(data)
	if m.cpu.currentROMBank == 0 {
		m.cpu.currentROMBank++
	}
}

func (m *Memory) ramBankChange(data byte) {
	m.currentRamBank = data & 0x3
}

func (m *Memory) changeROMRAMMode(data byte) {
	newData := data & 0x1
	if newData == 0 {
		m.romBanking = true
	} else {
		m.romBanking = false
	}
	if m.romBanking {
		m.currentRamBank = 0
	}
}

func (cpu *CPU) readMem(address uint16) byte {
	cpu.incrementCycles()
	return cpu.memory.get(address)
}

// TODO: remove this
// func (cpu *CPU) LoadProgram(program []byte) {
// 	cpu.memory.load(cpu.GetPC(), program)
// }

func (cpu *CPU) LoadBIOS(program []byte) {
	for i := 0; i < len(program); i++ {
		cpu.memory.bios[+uint(i)] = program[i]
	}
	cpu.loadBIOS = true
}

func (cpu *CPU) LoadROM(program []byte) {
	cartridgeType := program[0x147]
	romSize := program[0x148]
	if romSize > 0 {
		cpu.memory.romBanking = true
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
	case 1:
		cpu.mbc1 = true
	case 2:
		cpu.mbc1 = true
	case 3:
		cpu.mbc1 = true
	case 5:
		cpu.mbc2 = true
	case 6:
		cpu.mbc2 = true
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

func InitMemory(cpu *CPU) *Memory {
	return &Memory{
		bios:           [0x100]byte{},
		vram:           [0x2000]byte{},
		eram:           [0x2000]byte{},
		wram:           [0x2000]byte{},
		ioram:          [0x100]byte{},
		hram:           [0x7F]byte{},
		ramBanks:       [0x8000]byte{},
		currentRamBank: 0,
		cpu:            cpu,
	}
}

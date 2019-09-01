package cpu

import (
	"fmt"

	"github.com/tbtommyb/goboy/pkg/registers"
)

type Memory struct {
	rom             [0x8000]byte
	vram            [0x2000]byte
	eram            [0x2000]byte
	wram            [0x2000]byte
	sram            [0x100]byte
	ioram           [0x100]byte
	hram            [0x7F]byte
	interruptEnable byte
	statMode        byte
	cpu             *CPU
}

const ProgramStartAddress = 0x100
const StackStartAddress = 0xFFFE

func (m *Memory) load(start uint16, data []byte) {
	var i uint16
	for i = 0; i < uint16(len(data)); i++ {
		m.set(start+i, data[i])
	}
}

func (m *Memory) set(address uint16, value byte) {
	switch {
	case address < 0x8000:
		m.rom[address] = value
	case address >= 0x8000 && address <= 0x9FFF:
		// video ram
		m.vram[address-0x8000] = value
	case address >= 0xA000 && address <= 0xBFFF:
		m.eram[address-0xA000] = value
	case address >= 0xC000 && address <= 0xDFFF:
		m.wram[address-0xC000] = value
	case address >= 0xE000 && address <= 0xFDFF:
		// shadow wram
	case address >= 0xFE00 && address <= 0xFE9F:
		// sprites
		m.sram[address-0xFE00] = value
	case address >= 0xFEA0 && address <= 0xFEFF:
		// unusable
	case address >= 0xFF00 && address <= 0xFF7F:
		// memory mapped IO
		if address == 0xFF01 {
			// fmt.Printf("%c", value)
		} else if address == LYAddress {
			// Reset if game writes to LY
			m.ioram[address-0xFF00] = 0
		} else if address == 0xFF46 {
			// DMA
			m.performDMA(uint16(value) << 8)
		} else if address == TMCAddress {
			currentFreq := m.cpu.getClockFreq()
			m.ioram[address-0xFF00] = value
			newFreq := m.cpu.getClockFreq()

			if currentFreq != newFreq {
				m.cpu.setClockFreq()
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

func (m *Memory) performDMA(address uint16) {
	for i := uint16(0); i < 0xA0; i++ {
		m.set(0xFE00+i, m.get(address+i))
	}
}

func (m *Memory) get(address uint16) byte {
	switch {
	case address < 0x8000:
		return m.rom[address]
	case address >= 0x8000 && address <= 0x9FFF:
		// video ram
		return m.vram[address-0x8000]
	case address >= 0xA000 && address <= 0xBFFF:
		// cart ram
		return m.eram[address-0xA000]
	case address >= 0xC000 && address <= 0xDFFF:
		return m.wram[address-0xC000]
	case address >= 0xE000 && address <= 0xFDFF:
		// shadow wram
		return 0x00
	case address >= 0xFE00 && address <= 0xFE9F:
		// sprites
		val := m.sram[address-0xFE00]
		return val
	case address >= 0xFEA0 && address <= 0xFEFF:
		// unused space
		return 0xFF
	case address >= 0xFF00 && address <= 0xFF7F:
		// memory mapped IO
		// TODO: maybe map to memory instead
		if address == JoypadRegisterAddress {
			return m.cpu.getJoypadState()
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

func (cpu *CPU) readMem(address uint16) byte {
	cpu.incrementCycles()
	return cpu.memory.get(address)
}

// TODO: remove this
func (cpu *CPU) LoadProgram(program []byte) {
	cpu.memory.load(cpu.GetPC(), program)
}

func (cpu *CPU) LoadBIOS(program []byte) {
	cpu.memory.load(0, program)
}

func (cpu *CPU) LoadROM(program []byte) {
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

func InitMemory(cpu *CPU) *Memory {
	return &Memory{
		rom:   [0x8000]byte{},
		vram:  [0x2000]byte{},
		eram:  [0x2000]byte{},
		wram:  [0x2000]byte{},
		ioram: [0x100]byte{},
		hram:  [0x7F]byte{},
		cpu:   cpu,
	}
}

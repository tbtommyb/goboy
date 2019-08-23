package cpu

import (
	"fmt"

	"github.com/tbtommyb/goboy/pkg/registers"
)

var GameBoyColorMap = []int{0xFFFFFFFF, 0xB6B6B6FF, 0x676767FF, 0x000000FF}

type Memory struct {
	rom   [0x8000]byte
	vram  [0x2000]byte
	eram  [0x2000]byte
	wram  [0x2000]byte
	ioram [0x100]byte
	hram  [0x7F]byte
	flag  byte
}

const ProgramStartAddress = 0x100
const StackStartAddress = 0xFFFE

func (m *Memory) load(start uint16, data []byte) {
	var i uint16
	for i = 0; i < uint16(len(data)); i++ {
		m.set(start+i, data[i])
	}
}

func (m *Memory) set(address uint16, value byte) byte {
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
	case address >= 0xFEA0 && address <= 0xFEFF:
		panic(fmt.Sprintf("Invalid write to %x\n", address))
		// unusable
	case address >= 0xFF00 && address <= 0xFF7F:
		// memory mapped IO
		if address == 0xFF01 {
			fmt.Printf("%c", value)
		}
		m.ioram[address-0xFF00] = value
	case address >= 0xFF80 && address <= 0xFFFE:
		m.hram[address-0xFF80] = value
	case address == 0xFFFF:
		m.flag = value
	}
	return value
}

func (m *Memory) get(address uint16) byte {
	switch {
	case address < 0x8000:
		return m.rom[address]
	case address >= 0x8000 && address <= 0x9FFF:
		// video ram
		return m.vram[address-0x8000]
	case address >= 0xA000 && address <= 0xBFFF:
		return m.eram[address-0xA000]
		// cart ram
	case address >= 0xC000 && address <= 0xDFFF:
		return m.wram[address-0xC000]
	case address >= 0xE000 && address <= 0xFDFF:
		// shadow wram
		return 0x00
	case address >= 0xFE00 && address <= 0xFE9F:
		// sprites
		return 0x00
	case address >= 0xFEA0 && address <= 0xFEFF:
		return 0x00
	case address >= 0xFF00 && address <= 0xFF7F:
		// memory mapped IO
		return m.ioram[address-0xFF00]
	case address >= 0xFF80 && address <= 0xFFFE:
		return m.hram[address-0xFF80]
	case address == 0xFFFF:
		return m.flag
	default:
		panic(fmt.Sprintf("%x\n", address))
	}
}

func (cpu *CPU) readMem(address uint16) byte {
	cpu.incrementCycles()
	return cpu.memory.get(address)
}

func (cpu *CPU) LoadProgram(program []byte) {
	cpu.memory.load(cpu.GetPC(), program)
}

func (cpu *CPU) LoadROM(program []byte) {
	cpu.memory.load(0, program)
}

func (cpu *CPU) WriteMem(address uint16, value byte) byte {
	cpu.incrementCycles()
	return cpu.memory.set(address, value)
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

func (m *Memory) scrollY() byte {
	return m.get(0xFF42)
}

func (m *Memory) scrollX() byte {
	return m.get(0xFF43)
}

func (m *Memory) getLY() byte {
	return m.get(0xFF44)
}

func (m *Memory) incrementLY() {
	currentScanline := m.get(0xFF44)
	currentScanline++
	if currentScanline > 153 {
		m.set(0xFF44, 0)
	} else {
		m.set(0xFF44, currentScanline)
	}
}
func (m *Memory) backgroundPixelAt(x uint8, y uint8) int {
	// 32 tiles per row. y>>3 (same as y/8) gets the row. x>>3 (x/8) gets the columns
	tileMapOffset := (uint16(x) >> 3) + (uint16(y)>>3)*32
	tileSelectionAddress := m.bgTileMapStartAddress() + uint16(tileMapOffset)
	tileNumber := m.get(tileSelectionAddress)          // Which one of 256 tiles are to be shown
	tileDataAddress := m.bgTileDataAddress(tileNumber) // Where the 16-bytes of the tile begin

	tileYOffset := (y & 0x7) * 2 // Each row in the tile takes 2 bytes
	tileXOffset := (x & 0x7)     // Each col in the tile is 1 bit
	pixelByte := tileDataAddress + uint16(tileYOffset)
	pixLow := (m.get(pixelByte+1) >> (7 - tileXOffset)) & 0x1
	pixHigh := (m.get(pixelByte) >> (7 - tileXOffset)) & 0x1
	colorNumber := (pixHigh << 1) | pixLow
	return GameBoyColorMap[colorNumber]
}

func (m *Memory) bgTileDataAddress(tileNumber uint8) uint16 {
	tileAddress := uint16(0)
	if ((m.get(0xFF40) >> 4) & 0x1) == 0x1 {
		tileAddress = 0x8000
	} else {
		tileAddress = 0x8800
	}
	return tileAddress + uint16(tileNumber)*16
}

func (m *Memory) bgTileMapStartAddress() uint16 {
	if ((m.get(0xFF40) >> 3) & 0x1) == 0x1 {
		return 0x9C00
	}
	return 0x9800
}

func InitMemory() *Memory {
	return &Memory{
		rom:   [0x8000]byte{},
		vram:  [0x2000]byte{},
		eram:  [0x2000]byte{},
		wram:  [0x2000]byte{},
		ioram: [0x100]byte{},
		hram:  [0x7F]byte{},
	}
}

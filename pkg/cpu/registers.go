package cpu

import (
	"fmt"

	"github.com/tbtommyb/goboy/pkg/registers"
	"github.com/tbtommyb/goboy/pkg/utils"
)

const LCDCAddress = 0xFF40
const STATAddress = 0xFF41
const ScrollYAddress = 0xFF42
const ScrollXAddress = 0xFF43
const LYAddress = 0xFF44
const LYCAddress = 0xFF45
const BGPAddress = 0xFF47
const OBP0Address = 0xFF48
const OBP1Address = 0xFF49
const WindowYAddress = 0xFF4A
const WindowXAddress = 0xFF4B
const JoypadRegisterAddress = 0xFF00

func (cpu *CPU) Get(r registers.Single) byte {
	switch r {
	case registers.M:
		return cpu.readMem(cpu.GetHL())
	default:
		return byte(cpu.r[r])
	}
}

func (cpu *CPU) GetPair(r registers.Pair) (byte, byte) {
	switch r {
	case registers.BC:
		return cpu.Get(registers.B), cpu.Get(registers.C)
	case registers.DE:
		return cpu.Get(registers.D), cpu.Get(registers.E)
	case registers.HL:
		return cpu.Get(registers.H), cpu.Get(registers.L)
	case registers.AF:
		return cpu.Get(registers.A), cpu.GetFlags()
	case registers.SP:
		return utils.SplitPair(cpu.GetSP())
	default:
		panic(fmt.Sprintf("GetPair: Invalid register %x", r))
	}
}

func (cpu *CPU) Set(r registers.Single, val byte) byte {
	switch r {
	case registers.M:
		cpu.WriteMem(cpu.GetHL(), val)
	default:
		cpu.r[r] = val
	}
	return val
}

func (cpu *CPU) SetPair(r registers.Pair, val uint16) uint16 {
	switch r {
	case registers.BC:
		cpu.SetBC(val)
	case registers.DE:
		cpu.SetDE(val)
	case registers.HL:
		cpu.SetHL(val)
	case registers.SP:
		cpu.setSP(val)
	case registers.AF:
		cpu.SetAF(val)
	}
	return val
}

func (cpu *CPU) GetBC() uint16 {
	return utils.MergePair(cpu.Get(registers.B), cpu.Get(registers.C))
}

func (cpu *CPU) SetBC(value uint16) uint16 {
	cpu.Set(registers.B, byte(value>>8))
	cpu.Set(registers.C, byte(value))
	return value
}

func (cpu *CPU) GetDE() uint16 {
	return utils.MergePair(cpu.Get(registers.D), cpu.Get(registers.E))
}

func (cpu *CPU) SetDE(value uint16) uint16 {
	cpu.Set(registers.D, byte(value>>8))
	cpu.Set(registers.E, byte(value))
	return value
}

func (cpu *CPU) GetHL() uint16 {
	return utils.MergePair(cpu.Get(registers.H), cpu.Get(registers.L))
}

func (cpu *CPU) SetHL(value uint16) uint16 {
	cpu.Set(registers.H, byte(value>>8))
	cpu.Set(registers.L, byte(value))
	return value
}

func (cpu *CPU) SetAF(value uint16) uint16 {
	cpu.Set(registers.A, byte(value>>8))
	cpu.flags = byte(value & 0xf0)
	return value
}

func (cpu *CPU) GetAF() uint16 {
	return utils.MergePair(cpu.Get(registers.A), cpu.flags)
}

func (cpu *CPU) getLCDC() byte {
	return cpu.memory.get(LCDCAddress)
}

func (cpu *CPU) setLCDC(value byte) byte {
	return cpu.memory.set(LCDCAddress, value)
}

func (cpu *CPU) getSTAT() byte {
	return cpu.memory.get(STATAddress)
}

func (cpu *CPU) setSTAT(status byte) {
	cpu.memory.set(STATAddress, status)
}

func (cpu *CPU) getScrollY() byte {
	return cpu.memory.get(ScrollYAddress)
}

func (cpu *CPU) setScrollY(value byte) byte {
	return cpu.memory.set(ScrollYAddress, value)
}

func (cpu *CPU) getScrollX() byte {
	return cpu.memory.get(ScrollXAddress)
}

func (cpu *CPU) setScrollX(value byte) byte {
	return cpu.memory.set(ScrollXAddress, value)
}

func (cpu *CPU) getWindowY() byte {
	return cpu.memory.get(WindowYAddress)
}

func (cpu *CPU) getWindowX() byte {
	return cpu.memory.get(WindowXAddress)
}

func (cpu *CPU) getLY() byte {
	return cpu.memory.get(LYAddress)
}

func (cpu *CPU) setLY(value byte) {
	cpu.memory.ioram[LYAddress-0xFF00] = value
}

func (cpu *CPU) getLYC() byte {
	return cpu.memory.get(LYCAddress)
}

func (cpu *CPU) getBGP() byte {
	return cpu.memory.get(BGPAddress)
}

func (cpu *CPU) setBGP(value byte) byte {
	return cpu.memory.set(BGPAddress, value)
}

func (cpu *CPU) setOBP0(value byte) byte {
	return cpu.memory.set(OBP0Address, value)
}

func (cpu *CPU) getOBP0() byte {
	return cpu.memory.get(OBP0Address)
}

func (cpu *CPU) setOBP1(value byte) byte {
	return cpu.memory.set(OBP1Address, value)
}

func (cpu *CPU) getOBP1() byte {
	return cpu.memory.get(OBP1Address)
}

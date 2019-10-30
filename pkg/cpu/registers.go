package cpu

import (
	"fmt"

	"github.com/tbtommyb/goboy/pkg/registers"
	"github.com/tbtommyb/goboy/pkg/utils"
)

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

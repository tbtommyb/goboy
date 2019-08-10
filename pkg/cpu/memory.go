package cpu

import (
	"fmt"

	"github.com/tbtommyb/goboy/pkg/registers"
)

type Memory []byte

const ProgramStartAddress = 0x150
const StackStartAddress = 0xFF80

func (m Memory) load(start uint16, data []byte) {
	var i uint16
	for i = 0; i < uint16(len(data)); i++ {
		m[start+i] = data[i]
	}
}

func (m Memory) set(address uint16, value byte) byte {
	m[address] = value
	return value
}

func (m Memory) get(address uint16) byte {
	return m[address]
}

func (cpu *CPU) readMem(address uint16) byte {
	cpu.incrementCycles()
	return cpu.memory.get(address)
}

func (cpu *CPU) LoadProgram(program []byte) {
	cpu.memory.load(cpu.GetPC(), program)
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

func InitMemory() Memory {
	return make(Memory, 0xFFFF)
}

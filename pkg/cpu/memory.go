package cpu

import "fmt"

type AddressType byte
type Memory []byte

const ProgramStartAddress = 0x150
const StackStartAddress = 0xFF80
const (
	RelativeN  AddressType = 0x0
	RelativeC              = 0x2
	RelativeNN             = 0xA
)

func (m Memory) load(start int, data []byte) {
	for i := 0; i < len(data); i++ {
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
	cpu.memory.load(ProgramStartAddress, program)
}

func (cpu *CPU) WriteMem(address uint16, value byte) byte {
	cpu.incrementCycles()
	return cpu.memory.set(address, value)
}

func (cpu *CPU) GetMem(r RegisterPair) byte {
	switch r {
	case BC:
		return cpu.readMem(cpu.GetBC())
	case DE:
		return cpu.readMem(cpu.GetDE())
	case HL:
		return cpu.readMem(cpu.GetHL())
	case SP:
		return cpu.readMem(cpu.GetSP())
	default:
		panic(fmt.Sprintf("GetMem: Invalid register %x", r))
	}
}

func (cpu *CPU) SetMem(r RegisterPair, val byte) byte {
	switch r {
	case BC:
		cpu.WriteMem(cpu.GetBC(), val)
	case DE:
		cpu.WriteMem(cpu.GetDE(), val)
	case HL:
		cpu.WriteMem(cpu.GetHL(), val)
	case SP:
		cpu.WriteMem(cpu.GetSP(), val)
	default:
		panic(fmt.Sprintf("SetMem: Invalid register %x", r))
	}
	return val
}

func InitMemory() Memory {
	return make(Memory, 0xFFFF)
}

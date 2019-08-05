package cpu

import "fmt"

type Register byte
type RegisterPair byte
type Registers map[Register]byte

const (
	A Register = 0x7
	B          = 0x0
	C          = 0x1
	D          = 0x2
	E          = 0x3
	H          = 0x4
	L          = 0x5
	M          = 0x6 // memory reference through H:L
)

const (
	BC RegisterPair = 0x0
	DE              = 0x1
	HL              = 0x2
	SP              = 0x3
	AF              = 0x4
)

func (cpu *CPU) Get(r Register) byte {
	switch r {
	case M:
		return cpu.readMem(cpu.GetHL())
	default:
		return byte(cpu.r[r])
	}
}

func (cpu *CPU) GetPair(r RegisterPair) (byte, byte) {
	switch r {
	case BC:
		return cpu.Get(B), cpu.Get(C)
	case DE:
		return cpu.Get(D), cpu.Get(E)
	case HL:
		return cpu.Get(H), cpu.Get(L)
	case AF:
		return cpu.Get(A), cpu.GetFlags()
	default:
		panic(fmt.Sprintf("GetPair: Invalid register %x", r))
	}
}

func (cpu *CPU) Set(r Register, val byte) byte {
	switch r {
	case M:
		cpu.WriteMem(cpu.GetHL(), val)
	default:
		cpu.r[r] = val
	}
	return val
}

func (cpu *CPU) SetPair(r RegisterPair, val uint16) uint16 {
	switch r {
	case BC:
		cpu.SetBC(val)
	case DE:
		cpu.SetDE(val)
	case HL:
		cpu.SetHL(val)
	case SP:
		cpu.setSP(val)
	}
	return val
}

func (cpu *CPU) GetBC() uint16 {
	return mergePair(cpu.Get(B), cpu.Get(C))
}

func (cpu *CPU) SetBC(value uint16) uint16 {
	cpu.Set(B, byte(value>>8))
	cpu.Set(C, byte(value))
	return value
}

func (cpu *CPU) GetDE() uint16 {
	return mergePair(cpu.Get(D), cpu.Get(E))
}

func (cpu *CPU) SetDE(value uint16) uint16 {
	cpu.Set(D, byte(value>>8))
	cpu.Set(E, byte(value))
	return value
}

func (cpu *CPU) GetHL() uint16 {
	return mergePair(cpu.Get(H), cpu.Get(L))
}

func (cpu *CPU) SetHL(value uint16) uint16 {
	cpu.Set(H, byte(value>>8))
	cpu.Set(L, byte(value))
	return value
}

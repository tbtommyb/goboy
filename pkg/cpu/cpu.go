package cpu

import "fmt"

// TODO: consider what needs to be exported

type Register byte

// TODO: implement this
// type RegisterPair byte

// const (
// 	BC RegisterPair = 0x0
// 	DE = 0x1
// 	HL = 0x2
// 	SP = 0x3
// 	AF = 0x4
// )
const (
	A  Register = 0x7
	B           = 0x0
	C           = 0x1
	D           = 0x2
	E           = 0x3
	H           = 0x4
	L           = 0x5
	M           = 0x6 // memory reference through H:L
	BC          = 0x8 // pairs have 0x8 added as an offset to create unique values
	DE          = 0x9
	HL          = 0xA
	SP          = 0xB
	AF          = 0xC
)

type AddressType byte

const (
	RelativeN  AddressType = 0x0
	RelativeC              = 0x2
	RelativeNN             = 0xA
)

type Registers map[Register]byte

type CPU struct {
	r      Registers
	flags  Register
	SP, PC uint16
	memory Memory
	cycles uint
}

func (cpu *CPU) Get(r Register) byte {
	if r > A {
		panic(fmt.Sprintf("Get: invalid register %d", r))
	}
	switch r {
	case M:
		cpu.IncrementCycles()
		return cpu.memory.Get(cpu.GetHL())
	default:
		return byte(cpu.r[r])
	}
}

func (cpu *CPU) GetMem(r Register) byte {
	cpu.IncrementCycles()
	switch r {
	case BC:
		return cpu.memory.Get(cpu.GetBC())
	case DE:
		return cpu.memory.Get(cpu.GetDE())
	case HL:
		return cpu.memory.Get(cpu.GetHL())
	case SP:
		return cpu.memory.Get(cpu.GetSP())
	default:
		panic(fmt.Sprintf("GetMem: Invalid register %x", r))
	}
}

func (cpu *CPU) GetPair(r Register) (byte, byte) {
	cpu.IncrementCycles()
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
	if r > A {
		panic(fmt.Sprintf("Get: invalid register %d", r))
	}
	switch r {
	case M:
		cpu.IncrementCycles()
		cpu.memory.Set(cpu.GetHL(), val)
	default:
		cpu.r[r] = val
	}
	return val
}

func (cpu *CPU) SetMem(r Register, val byte) byte {
	cpu.IncrementCycles()
	switch r {
	case BC:
		cpu.memory.Set(cpu.GetBC(), val)
	case DE:
		cpu.memory.Set(cpu.GetDE(), val)
	case HL:
		cpu.memory.Set(cpu.GetHL(), val)
	case SP:
		cpu.memory.Set(cpu.GetSP(), val)
	default:
		panic(fmt.Sprintf("SetMem: Invalid register %x", r))
	}
	return val
}

func (cpu *CPU) SetPair(r Register, val uint16) uint16 {
	switch r {
	case BC:
		cpu.SetBC(val)
	case DE:
		cpu.SetDE(val)
	case HL:
		cpu.SetHL(val)
	case SP:
		cpu.SetSP(val)
	}
	return val
}

func (cpu *CPU) GetFlags() byte {
	return byte(cpu.flags)
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

func (cpu *CPU) GetSP() uint16 {
	return cpu.SP
}

func (cpu *CPU) SetSP(value uint16) uint16 {
	cpu.SP = value
	return value
}

func (cpu *CPU) IncrementSP() {
	cpu.SP += 1
}

func (cpu *CPU) DecrementSP() {
	cpu.SP -= 1
}

func (cpu *CPU) GetPC() uint16 {
	return cpu.PC
}

func (cpu *CPU) IncrementPC() {
	cpu.PC += 1
}

func (cpu *CPU) GetCycles() uint {
	return cpu.cycles
}

func (cpu *CPU) IncrementCycles() {
	cpu.cycles += 1
}

// TODO: create separate stack structure
func (cpu *CPU) PushStack(val byte) byte {
	cpu.DecrementSP()
	return cpu.SetMem(SP, val)
}

func (cpu *CPU) PopStack() byte {
	val := cpu.GetMem(SP)
	cpu.IncrementSP()
	return val
}

func (cpu *CPU) computeOffset(offset uint16) uint16 {
	cpu.IncrementCycles()
	return 0xFF00 + offset
}

func (cpu *CPU) fetchAndIncrement() byte {
	value := cpu.memory[cpu.GetPC()]
	cpu.IncrementPC()
	cpu.IncrementCycles()
	return value
}

func mergePair(high, low byte) uint16 {
	return uint16(high)<<8 | uint16(low)
}

func (cpu *CPU) Run() {
	for opcode := cpu.fetchAndIncrement(); opcode != 0; opcode = cpu.fetchAndIncrement() {
		instr := Decode(opcode)
		switch i := instr.(type) {
		case Move:
			cpu.Set(i.dest, cpu.Get(i.source))
		case MoveImmediate:
			i.immediate = cpu.fetchAndIncrement()
			cpu.Set(i.dest, i.immediate)
		case LoadIndirect:
			cpu.Set(i.dest, cpu.GetMem(i.source))
		case StoreIndirect:
			address := mergePair(cpu.GetPair(i.dest))
			cpu.memory.Set(address, cpu.Get(i.source))
		case LoadRelative:
			var source uint16
			switch i.addressType {
			case RelativeC:
				source = cpu.computeOffset(uint16(cpu.Get(C)))
			case RelativeN:
				source = cpu.computeOffset(uint16(cpu.fetchAndIncrement()))
			case RelativeNN:
				source = mergePair(cpu.fetchAndIncrement(), cpu.fetchAndIncrement())
				cpu.IncrementCycles()
			}

			cpu.Set(A, cpu.memory.Get(source))
		case StoreRelative:
			var dest uint16
			switch i.addressType {
			case RelativeC:
				dest = cpu.computeOffset(uint16(cpu.Get(C)))
			case RelativeN:
				dest = cpu.computeOffset(uint16(cpu.fetchAndIncrement()))
			case RelativeNN:
				dest = mergePair(cpu.fetchAndIncrement(), cpu.fetchAndIncrement())
				cpu.IncrementCycles()
			}
			cpu.memory.Set(dest, cpu.Get(A))
		case LoadIncrement:
			cpu.Set(A, cpu.GetMem(HL))
			cpu.SetHL(cpu.GetHL() + 1)
		case LoadDecrement:
			cpu.Set(A, cpu.GetMem(HL))
			cpu.SetHL(cpu.GetHL() - 1)
		case StoreIncrement:
			cpu.SetMem(HL, cpu.Get(A))
			cpu.SetHL(cpu.GetHL() + 1)
		case StoreDecrement:
			cpu.SetMem(HL, cpu.Get(A))
			cpu.SetHL(cpu.GetHL() - 1)
		case LoadRegisterPairImmediate:
			var immediate uint16
			immediate |= uint16(cpu.fetchAndIncrement())
			immediate |= uint16(cpu.fetchAndIncrement()) << 8
			cpu.SetPair(i.dest, immediate)
		case HLtoSP:
			cpu.SetPair(SP, cpu.GetHL())
			cpu.IncrementCycles() // TODO: remove need for this
		case Push:
			high, low := cpu.GetPair(i.source)
			cpu.PushStack(high)
			cpu.PushStack(low)
		case Pop:
			low := cpu.PopStack()
			high := cpu.PopStack()
			cpu.SetPair(i.dest, mergePair(high, low))
		case InvalidInstruction:
			panic(fmt.Sprintf("Invalid Instruction: %x", instr.Opcode()))
		}
	}
}

// TODO: RunProgram convenience method?
func (cpu *CPU) LoadProgram(program []byte) {
	cpu.memory.Load(ProgramStartAddress, program)
}

func Init() CPU {
	return CPU{
		r: Registers{
			A: 0, B: 0, C: 0, D: 0, E: 0, H: 0, L: 0,
		}, SP: StackStartAddress, PC: ProgramStartAddress,
		memory: InitMemory(),
	}
}

package cpu

import "fmt"

// TODO: consider what needs to be exported

type Register byte
type RegisterPair byte
type AddressType byte
type Registers map[Register]byte
type CPU struct {
	r      Registers
	flags  Register
	SP, PC uint16
	memory Memory
	cycles uint
}

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

const (
	RelativeN  AddressType = 0x0
	RelativeC              = 0x2
	RelativeNN             = 0xA
)

func (cpu *CPU) Get(r Register) byte {
	switch r {
	case M:
		return cpu.getMem(cpu.GetHL())
	default:
		return byte(cpu.r[r])
	}
}

func (cpu *CPU) GetMem(r RegisterPair) byte {
	switch r {
	case BC:
		return cpu.getMem(cpu.GetBC())
	case DE:
		return cpu.getMem(cpu.GetDE())
	case HL:
		return cpu.getMem(cpu.GetHL())
	case SP:
		return cpu.getMem(cpu.GetSP())
	default:
		panic(fmt.Sprintf("GetMem: Invalid register %x", r))
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
		cpu.writeMem(cpu.GetHL(), val)
	default:
		cpu.r[r] = val
	}
	return val
}

func (cpu *CPU) SetMem(r RegisterPair, val byte) byte {
	switch r {
	case BC:
		cpu.writeMem(cpu.GetBC(), val)
	case DE:
		cpu.writeMem(cpu.GetDE(), val)
	case HL:
		cpu.writeMem(cpu.GetHL(), val)
	case SP:
		cpu.writeMem(cpu.GetSP(), val)
	default:
		panic(fmt.Sprintf("SetMem: Invalid register %x", r))
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
	return 0xFF00 + offset
}

func (cpu *CPU) fetchAndIncrement() byte {
	value := cpu.memory.Get(cpu.GetPC())
	cpu.IncrementPC()
	cpu.IncrementCycles()
	return value
}

func mergePair(high, low byte) uint16 {
	return uint16(high)<<8 | uint16(low)
}

func (cpu *CPU) fetchRelative(i RelativeAddressingInstruction) uint16 {
	switch addressType := i.AddressType(); addressType {
	case RelativeC:
		return cpu.computeOffset(uint16(cpu.Get(C)))
	case RelativeN:
		return cpu.computeOffset(uint16(cpu.fetchAndIncrement()))
	case RelativeNN:
		return mergePair(cpu.fetchAndIncrement(), cpu.fetchAndIncrement())
	default:
		panic(fmt.Sprintf("fetchRelative: invalid address type %d", addressType))
	}
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
			cpu.writeMem(address, cpu.Get(i.source))
		case LoadRelative:
			cpu.Set(A, cpu.getMem(cpu.fetchRelative(i)))
		case StoreRelative:
			cpu.writeMem(cpu.fetchRelative(i), cpu.Get(A))
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
			cpu.IncrementCycles() // TODO: remove need for this
		case Pop:
			low := cpu.PopStack()
			high := cpu.PopStack()
			cpu.SetPair(i.dest, mergePair(high, low))
		case LoadHLSP:
			// TODO: flags
			immediate := int8(cpu.fetchAndIncrement())
			cpu.SetPair(HL, cpu.GetSP()+uint16(immediate))
			cpu.IncrementCycles() // TODO: remove need for this
		case StoreSP:
			var immediate uint16
			immediate |= uint16(cpu.fetchAndIncrement())
			immediate |= uint16(cpu.fetchAndIncrement()) << 8
			cpu.writeMem(immediate, byte(cpu.GetSP()))
			cpu.writeMem(immediate+1, byte(cpu.GetSP()>>8))
		case InvalidInstruction:
			panic(fmt.Sprintf("Invalid Instruction: %x", instr.Opcode()))
		}
	}
}

// TODO: RunProgram convenience method?
func (cpu *CPU) LoadProgram(program []byte) {
	cpu.memory.Load(ProgramStartAddress, program)
}

func (cpu *CPU) writeMem(address uint16, value byte) byte {
	cpu.IncrementCycles()
	return cpu.memory.Set(address, value)
}

func (cpu *CPU) getMem(address uint16) byte {
	cpu.IncrementCycles()
	return cpu.memory.Get(address)
}

func Init() CPU {
	return CPU{
		r: Registers{
			A: 0, B: 0, C: 0, D: 0, E: 0, H: 0, L: 0,
		}, SP: StackStartAddress, PC: ProgramStartAddress,
		memory: InitMemory(),
	}
}

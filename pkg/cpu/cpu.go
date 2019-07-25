package cpu

import "fmt"

// TODO: consider what needs to be exported

type Register byte

const (
	A  Register = 0x7
	B           = 0x0
	C           = 0x1
	D           = 0x2
	E           = 0x3
	H           = 0x4
	L           = 0x5
	M           = 0x6 // memory reference through H:L
	BC          = 0x8
	DE          = 0x9
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
	SP, PC uint16
	memory Memory
	cycles uint
}

func (cpu *CPU) Get(r Register) byte {
	if r == BC {
		return cpu.memory.Get(cpu.GetBC())
	}
	if r == DE {
		return cpu.memory.Get(cpu.GetDE())
	}
	if r == M {
		return cpu.memory.Get(cpu.GetHL())
	}
	return byte(cpu.r[r])
}

func (cpu *CPU) GetBC() uint16 {
	cpu.IncrementCycles()
	return uint16(cpu.Get(B))<<8 | uint16(cpu.Get(C))
}

func (cpu *CPU) GetDE() uint16 {
	cpu.IncrementCycles()
	return uint16(cpu.Get(D))<<8 | uint16(cpu.Get(E))
}

func (cpu *CPU) GetHL() uint16 {
	cpu.IncrementCycles()
	return uint16(cpu.Get(H))<<8 | uint16(cpu.Get(L))
}

func (cpu *CPU) SetHL(value uint16) uint16 {
	cpu.set(H, byte(value>>8))
	cpu.set(L, byte(value))
	return value
}

func (cpu *CPU) IncrementHL(current uint16) uint16 {
	return cpu.SetHL(uint16(int(current + 1)))
}

func (cpu *CPU) DecrementHL(current uint16) uint16 {
	return cpu.SetHL(uint16(int(current - 1)))
}

func (cpu *CPU) set(r Register, val byte) byte {
	if r == BC {
		cpu.memory.Set(cpu.GetBC(), val)
		return val
	}
	if r == DE {
		cpu.memory.Set(cpu.GetDE(), val)
		return val
	}
	if r == M {
		cpu.memory.Set(cpu.GetHL(), val)
		return val
	}
	cpu.r[r] = val
	return val
}

func (cpu *CPU) GetSP() uint16 {
	return cpu.SP
}

func (cpu *CPU) IncrementSP() {
	cpu.SP += 1
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

func (cpu *CPU) Run() {
	for opcode := cpu.fetchAndIncrement(); opcode != 0; opcode = cpu.fetchAndIncrement() {
		instr := Decode(opcode)
		switch i := instr.(type) {
		case Move:
			cpu.set(i.dest, cpu.Get(i.source))
		case MoveImmediate:
			i.immediate = cpu.fetchAndIncrement()
			cpu.set(i.dest, i.immediate)
		case MoveIndirect:
			cpu.set(i.dest, cpu.Get(i.source))
		case LoadRelative:
			var source uint16
			switch i.addressType {
			case RelativeC:
				source = cpu.computeOffset(uint16(cpu.Get(C)))
			case RelativeN:
				source = cpu.computeOffset(uint16(cpu.fetchAndIncrement()))
			case RelativeNN:
				source |= uint16(cpu.fetchAndIncrement()) << 8
				source |= uint16(cpu.fetchAndIncrement())
				cpu.IncrementCycles()
			}

			cpu.set(A, cpu.memory.Get(source))
		case StoreRelative:
			var dest uint16
			switch i.addressType {
			case RelativeC:
				dest = cpu.computeOffset(uint16(cpu.Get(C)))
			case RelativeN:
				dest = cpu.computeOffset(uint16(cpu.fetchAndIncrement()))
			case RelativeNN:
				dest |= uint16(cpu.fetchAndIncrement()) << 8
				dest |= uint16(cpu.fetchAndIncrement())
				cpu.IncrementCycles()
			}
			cpu.memory.Set(dest, cpu.Get(A))
		case LoadIncrement:
			hl := cpu.GetHL()
			cpu.set(A, cpu.memory.Get(hl))
			cpu.IncrementHL(hl)
		case LoadDecrement:
			hl := cpu.GetHL()
			cpu.set(A, cpu.memory.Get(hl))
			cpu.DecrementHL(hl)
		case StoreIncrement:
			hl := cpu.GetHL()
			cpu.memory.Set(hl, cpu.Get(A))
			cpu.IncrementHL(hl)
		case StoreDecrement:
			hl := cpu.GetHL()
			cpu.memory.Set(hl, cpu.Get(A))
			cpu.DecrementHL(hl)
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
		}, SP: 0, PC: ProgramStartAddress,
		memory: InitMemory(),
	}
}

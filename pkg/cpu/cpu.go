package cpu

import "fmt"

// TODO: consider what needs to be exported

type Register byte

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

type Registers map[Register]byte

type CPU struct {
	r      Registers
	SP, PC uint16
	memory Memory
	cycles uint
}

func (cpu *CPU) Get(r Register) byte {
	if r == M {
		return cpu.memory.Get(cpu.GetHL())
	}
	return byte(cpu.r[r])
}

func (cpu *CPU) GetBC() uint16 {
	return uint16(cpu.Get(B))<<8 | uint16(cpu.Get(C))
}

func (cpu *CPU) GetDE() uint16 {
	return uint16(cpu.Get(D))<<8 | uint16(cpu.Get(E))
}

func (cpu *CPU) GetHL() uint16 {
	cpu.IncrementCycles()
	return uint16(cpu.Get(H))<<8 | uint16(cpu.Get(L))
}

func (cpu *CPU) set(r Register, val byte) byte {
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
		case Load:
			cpu.set(i.dest, cpu.Get(i.source))
		case LoadImmediate:
			i.immediate = cpu.fetchAndIncrement()
			cpu.set(i.dest, i.immediate)
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

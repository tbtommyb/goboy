package cpu

// TODO: consider what needs to be exported

type Register byte

const (
	A Register = 0x7
	B          = 0x0
	C          = 0x1
	D          = 0x5
	E          = 0x3
	F          = 0x2 // TODO: made up number
	H          = 0x4
	L          = 0x6 // TODO: should be 0x5 but avoiding duplicate with D
)

type Registers map[Register]byte

type CPU struct {
	r      Registers
	SP, PC uint16
	memory Memory
}

func (cpu *CPU) Get(r Register) byte {
	return byte(cpu.r[r])
}

func (cpu *CPU) GetBC() uint16 {
	return uint16(cpu.Get(B))<<8 | uint16(cpu.Get(C))
}

func (cpu *CPU) GetDE() uint16 {
	return uint16(cpu.Get(D))<<8 | uint16(cpu.Get(E))
}

func (cpu *CPU) GetHL() uint16 {
	return uint16(cpu.Get(H))<<8 | uint16(cpu.Get(L))
}

func (cpu *CPU) set(r Register, val byte) byte {
	cpu.r[r] = val
	return cpu.Get(r)
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

func (cpu *CPU) fetchAndIncrement() byte {
	value := cpu.memory[cpu.GetPC()]
	cpu.IncrementPC()
	return value
}

func (cpu *CPU) Run() {
	opcode := cpu.fetchAndIncrement()
	for ok := true; ok; ok = (opcode != 0) {
		instr := cpu.Decode(opcode)
		switch i := instr.(type) {
		case LoadRegister:
			cpu.set(i.dest, cpu.Get(i.source))
		case LoadImmediate:
			cpu.set(i.dest, i.immediate)
		case InvalidInstruction:
			return
		}
		opcode = cpu.fetchAndIncrement()
	}
}

// TODO: RunProgram convenience method?
func (cpu *CPU) LoadProgram(program []byte) {
	cpu.memory.Load(ProgramStartAddress, program)
}

func Init() CPU {
	return CPU{
		r: Registers{
			A: 0, B: 0, C: 0, D: 0, E: 0, F: 0, H: 0, L: 0,
		}, SP: 0, PC: ProgramStartAddress,
		memory: InitMemory(),
	}
}

package cpu

const LoadRegisterMask = 0xC0
const LoadRegisterPattern = 0x40
const LoadImmediateMask = 0xC7
const LoadImmediatePattern = 0x6
const LoadRegisterMemoryMask = 0xC7
const LoadRegisterMemoryPattern = 0x46
const StoreMemoryRegisterMask = 0xF8
const StoreMemoryRegisterPattern = 0x70

const DestRegisterMask = 0x38
const DestRegisterShift = 3
const SourceRegisterMask = 0x7

type Instruction interface {
	Opcode() []byte
	MicroOps(*CPU) []MicroOp
}

type InvalidInstruction struct{ opcode byte }

func (i InvalidInstruction) Opcode() []byte { return []byte{i.opcode} }

type EmptyInstruction struct{}

func (i EmptyInstruction) Opcode() []byte { return []byte{0} }

type LoadRegister struct {
	source, dest Register
}

func (i LoadRegister) Opcode() []byte {
	return []byte{byte(LoadRegisterPattern | i.source | i.dest<<DestRegisterShift)}
}

type MicroOp func(*byte) byte

type registerSet struct {
	dest Register
	val  byte
}

func (cpu *CPU) fetchImmediate() MicroOp {
	// cycles
	return func(*byte) byte {
		return cpu.fetchAndIncrement()
	}
}

func (cpu *CPU) fetchRegister(r Register) MicroOp {
	return func(*byte) byte {
		return cpu.Get(r)
	}
}

func (cpu *CPU) registerSet(r Register) MicroOp {
	// cycles
	return func(val *byte) byte {
		return cpu.set(r, *val)
	}
}

func (i LoadRegister) MicroOps(cpu *CPU) []MicroOp {
	return []MicroOp{
		cpu.fetchRegister(i.source), cpu.registerSet(i.dest),
	}
}

func (i LoadImmediate) MicroOps(cpu *CPU) []MicroOp {
	return []MicroOp{
		cpu.fetchImmediate(), cpu.registerSet(i.dest),
	}
}

type LoadImmediate struct {
	dest Register
}

func (i LoadImmediate) Opcode() byte {
	return byte(LoadImmediatePattern | i.dest<<DestRegisterShift)
}

type LoadRegisterMemory struct {
	dest Register
}

func (i LoadRegisterMemory) Opcode() byte {
	return byte(LoadRegisterMemoryPattern | i.dest<<DestRegisterShift)
}

type StoreMemoryRegister struct {
	source Register
}

func (i StoreMemoryRegister) Opcode() byte {
	return byte(StoreMemoryRegisterPattern | i.source)
}

func (cpu *CPU) Decode(op byte) Instruction {
	switch {
	// case op&LoadRegisterMemoryMask == LoadRegisterMemoryPattern:
	// 	// LD D, (HL), 0b01ddd110
	// 	// This bit pattern conflicts with the value of L
	// 	// so needs to come before LoadRegister check
	// 	// OR take H, L out of const Register assignment?
	// 	dest := Register(op & DestRegisterMask >> DestRegisterShift)

	// 	return LoadRegisterMemory{dest}
	// case op&StoreMemoryRegisterMask == StoreMemoryRegisterPattern:
	// 	source := Register(op & SourceRegisterMask)
	// 	return StoreMemoryRegister{source}
	case op&LoadRegisterMask == LoadRegisterPattern:
		// LD D, S. 0b01dddsss
		source := Register(op & SourceRegisterMask)
		dest := Register(op & DestRegisterMask >> DestRegisterShift)
		return LoadRegister{source, dest}
	case op&LoadImmediateMask == LoadImmediatePattern:
		// LD D, n. 0b00ddd110
		dest := Register(op & DestRegisterMask >> DestRegisterShift) // TODO extract this
		return LoadImmediate{dest: dest}
		// case op == 0:
		// 	return EmptyInstruction{}
		// default:
		// 	return InvalidInstruction{opcode: op}
	default:
		panic("!!")

	}
}

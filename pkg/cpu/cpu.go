package cpu

import "fmt"

type CPU struct {
	r      Registers
	flags  byte
	SP, PC uint16
	memory Memory
	cycles uint
}

func (cpu *CPU) GetPC() uint16 {
	return cpu.PC
}

func (cpu *CPU) incrementPC() {
	cpu.PC += 1
}

func (cpu *CPU) GetCycles() uint {
	return cpu.cycles
}

func (cpu *CPU) incrementCycles() {
	cpu.cycles += 1
}

func (cpu *CPU) fetchAndIncrement() byte {
	value := cpu.memory.get(cpu.GetPC())
	cpu.incrementPC()
	cpu.incrementCycles()
	return value
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

func addOp(a, b, carry byte) (byte, FlagSet) {
	result := a + b + carry
	flagSet := FlagSet{
		Zero:      result == 0,
		Negative:  false,
		HalfCarry: isAddHalfCarry(a, b, carry),
		FullCarry: isAddFullCarry(a, b, carry),
	}
	return result, flagSet
}

func subtractOp(a, b, carry byte) (byte, FlagSet) {
	result := a - b - carry
	flagSet := FlagSet{
		Zero:      result == 0,
		Negative:  true,
		HalfCarry: isSubHalfCarry(a, b, carry),
		FullCarry: isSubFullCarry(a, b, carry),
	}
	return result, flagSet
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
			cpu.WriteMem(address, cpu.Get(i.source))
		case LoadRelative:
			cpu.Set(A, cpu.readMem(cpu.fetchRelative(i)))
		case StoreRelative:
			cpu.WriteMem(cpu.fetchRelative(i), cpu.Get(A))
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
			cpu.incrementCycles() // TODO: remove need for this
		case Push:
			high, low := cpu.GetPair(i.source)
			cpu.pushStack(high)
			cpu.pushStack(low)
			cpu.incrementCycles() // TODO: remove need for this
		case Pop:
			low := cpu.popStack()
			high := cpu.popStack()
			cpu.SetPair(i.dest, mergePair(high, low))
		case LoadHLSP:
			// TODO: flags
			immediate := int8(cpu.fetchAndIncrement())
			cpu.SetPair(HL, cpu.GetSP()+uint16(immediate))
			cpu.incrementCycles() // TODO: remove need for this
		case StoreSP:
			var immediate uint16
			immediate |= uint16(cpu.fetchAndIncrement())
			immediate |= uint16(cpu.fetchAndIncrement()) << 8
			cpu.WriteMem(immediate, byte(cpu.GetSP()))
			cpu.WriteMem(immediate+1, byte(cpu.GetSP()>>8))
		case Add:
			a := cpu.Get(A)
			b := cpu.Get(i.source)
			var carry byte
			if i.withCarry && cpu.isSet(FullCarry) {
				carry = 1
			}
			result, flagSet := addOp(a, b, carry)
			cpu.Set(A, result)
			cpu.setFlags(flagSet)
		case AddImmediate:
			a := cpu.Get(A)
			b := cpu.fetchAndIncrement()
			var carry byte
			if i.withCarry && cpu.isSet(FullCarry) {
				carry = 1
			}
			result, flagSet := addOp(a, b, carry)
			cpu.Set(A, result)
			cpu.setFlags(flagSet)
		case Subtract:
			a := cpu.Get(A)
			b := cpu.Get(i.source)
			var carry byte
			if i.withCarry && cpu.isSet(FullCarry) {
				carry = 1
			}
			result, flagSet := subtractOp(a, b, carry)
			cpu.Set(A, result)
			cpu.setFlags(flagSet)
		case SubtractImmediate:
			a := cpu.Get(A)
			b := cpu.fetchAndIncrement()
			var carry byte
			if i.withCarry && cpu.isSet(FullCarry) {
				carry = 1
			}
			result, flagSet := subtractOp(a, b, carry)
			cpu.Set(A, result)
			cpu.setFlags(flagSet)
		case InvalidInstruction:
			panic(fmt.Sprintf("Invalid Instruction: %x", instr.Opcode()))
		}
	}
}

func Init() CPU {
	return CPU{
		r: Registers{
			A: 0, B: 0, C: 0, D: 0, E: 0, H: 0, L: 0,
		}, SP: StackStartAddress, PC: ProgramStartAddress,
		memory: InitMemory(),
	}
}

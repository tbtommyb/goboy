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

func addOp(args ...byte) (byte, FlagSet) {
	a, b, carry := args[0], args[1], args[2]
	result := a + b + carry
	flagSet := FlagSet{
		Zero:      result == 0,
		Negative:  false,
		HalfCarry: isAddHalfCarry(a, b, carry),
		FullCarry: isAddFullCarry(a, b, carry),
	}
	return result, flagSet
}

func subOp(args ...byte) (byte, FlagSet) {
	a, b, carry := args[0], args[1], args[2]
	result := a - b - carry
	flagSet := FlagSet{
		Zero:      result == 0,
		Negative:  true,
		HalfCarry: isSubHalfCarry(a, b, carry),
		FullCarry: isSubFullCarry(a, b, carry),
	}
	return result, flagSet
}

func andOp(args ...byte) (byte, FlagSet) {
	a, b := args[0], args[1]
	result := a & b
	flagSet := FlagSet{
		Zero:      result == 0,
		Negative:  false,
		HalfCarry: true,
		FullCarry: false,
	}
	return result, flagSet
}

func orOp(args ...byte) (byte, FlagSet) {
	a, b := args[0], args[1]
	result := a | b
	flagSet := FlagSet{
		Zero:      result == 0,
		Negative:  false,
		HalfCarry: false,
		FullCarry: false,
	}
	return result, flagSet
}

func xorOp(args ...byte) (byte, FlagSet) {
	a, b := args[0], args[1]
	result := a ^ b
	flagSet := FlagSet{
		Zero:      result == 0,
		Negative:  false,
		HalfCarry: false,
		FullCarry: false,
	}
	return result, flagSet
}

func cmpOp(args ...byte) FlagSet {
	a, b := args[0], args[1]
	result := a ^ b
	return FlagSet{
		Zero:      result == 0,
		Negative:  true,
		HalfCarry: isSubHalfCarry(a, b, 0),
		FullCarry: isSubFullCarry(a, b, 0),
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
			cpu.setSP(cpu.GetHL())
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
			// TODO flags
			immediate := int8(cpu.fetchAndIncrement())
			cpu.SetHL(cpu.GetSP() + uint16(immediate))
			cpu.incrementCycles() // TODO: remove need for this
		case StoreSP:
			var immediate uint16
			immediate |= uint16(cpu.fetchAndIncrement())
			immediate |= uint16(cpu.fetchAndIncrement()) << 8
			cpu.WriteMem(immediate, byte(cpu.GetSP()))
			cpu.WriteMem(immediate+1, byte(cpu.GetSP()>>8))
		case Add:
			carry := cpu.carryBit(i.withCarry, FullCarry)
			cpu.perform(addOp, cpu.Get(A), cpu.Get(i.source), carry)
		case AddImmediate:
			carry := cpu.carryBit(i.withCarry, FullCarry)
			cpu.perform(addOp, cpu.Get(A), cpu.fetchAndIncrement(), carry)
		case Subtract:
			carry := cpu.carryBit(i.withCarry, FullCarry)
			cpu.perform(subOp, cpu.Get(A), cpu.Get(i.source), carry)
		case SubtractImmediate:
			carry := cpu.carryBit(i.withCarry, FullCarry)
			cpu.perform(subOp, cpu.Get(A), cpu.fetchAndIncrement(), carry)
		case And:
			cpu.perform(andOp, cpu.Get(A), cpu.Get(i.source))
		case AndImmediate:
			cpu.perform(andOp, cpu.Get(A), cpu.fetchAndIncrement())
		case Or:
			cpu.perform(orOp, cpu.Get(A), cpu.Get(i.source))
		case OrImmediate:
			cpu.perform(orOp, cpu.Get(A), cpu.fetchAndIncrement())
		case Xor:
			cpu.perform(xorOp, cpu.Get(A), cpu.Get(i.source))
		case XorImmediate:
			cpu.perform(xorOp, cpu.Get(A), cpu.fetchAndIncrement())
		case Cmp:
			flagSet := cmpOp(cpu.Get(A), cpu.Get(i.source))
			cpu.setFlags(flagSet)
		case CmpImmediate:
			flagSet := cmpOp(cpu.Get(A), cpu.fetchAndIncrement())
			cpu.setFlags(flagSet)
		case Increment:
			a := cpu.Get(i.dest)
			result := a + 1
			cpu.Set(i.dest, result)
			cpu.setFlags(FlagSet{
				Zero:      result == 0,
				HalfCarry: isAddHalfCarry(a, 1, 0),
				FullCarry: cpu.isSet(FullCarry),
				Negative:  false,
			})
		case Decrement:
			a := cpu.Get(i.dest)
			result := a - 1
			cpu.Set(i.dest, result)
			cpu.setFlags(FlagSet{
				Zero:      result == 0,
				HalfCarry: isSubHalfCarry(a, 1, 0),
				FullCarry: cpu.isSet(FullCarry),
				Negative:  true,
			})
		case AddPair:
			a := cpu.GetHL()
			b := mergePair(cpu.GetPair(i.source))
			result := a + b
			cpu.SetHL(result)
			cpu.setFlags(FlagSet{
				Zero:      cpu.isSet(Zero),
				Negative:  false,
				HalfCarry: isAddHalfCarry16(a, b),
				FullCarry: isAddFullCarry16(a, b),
			})
			cpu.incrementCycles()
		case AddSP:
			a := cpu.GetSP()
			b := cpu.fetchAndIncrement()
			result := a + uint16(b)
			cpu.setSP(result)
			cpu.setFlags(FlagSet{
				Zero:      false,
				Negative:  false,
				HalfCarry: isAddHalfCarry16(a, uint16(b)),
				FullCarry: isAddFullCarry16(a, uint16(b)),
			})
			cpu.fetchAndIncrement()
		case InvalidInstruction:
			panic(fmt.Sprintf("Invalid Instruction: %x", instr.Opcode()))
		}
	}
}

func (cpu *CPU) perform(f func(...byte) (byte, FlagSet), args ...byte) {
	result, flagSet := f(args...)
	cpu.Set(A, result)
	cpu.setFlags(flagSet)
}

func (cpu *CPU) carryBit(withCarry bool, flag Flag) byte {
	var carry byte
	if withCarry && cpu.isSet(flag) {
		carry = 1
	}
	return carry
}

func Init() CPU {
	return CPU{
		r: Registers{
			A: 0, B: 0, C: 0, D: 0, E: 0, H: 0, L: 0,
		}, SP: StackStartAddress, PC: ProgramStartAddress,
		memory: InitMemory(),
	}
}

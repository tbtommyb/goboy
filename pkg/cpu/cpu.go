package cpu

import (
	"fmt"
	"math/bits"
)

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

func rotateOp(i RotateInstruction, value, flag byte) (byte, FlagSet) {
	var result byte
	var flags FlagSet
	switch i.Direction() {
	case RotateLeft:
		result = bits.RotateLeft8(value, 1)
		if !i.WithCopy() {
			result = setBit(0, result, flag)
		}
		flags = FlagSet{
			FullCarry: bits.LeadingZeros8(value) == 0,
			Zero:      result == 0,
		}
	case RotateRight:
		result = bits.RotateLeft8(value, -1)
		if !i.WithCopy() {
			result = setBit(7, result, flag)
		}
		flags = FlagSet{
			FullCarry: bits.TrailingZeros8(value) == 0,
			Zero:      result == 0,
		}
	}
	return result, flags
}

func shiftOp(i RotateInstruction, value, flag byte) (byte, FlagSet) {
	var result byte
	var flags FlagSet
	switch i.Direction() {
	case RotateLeft:
		result = value << 1
		flags = FlagSet{
			FullCarry: bits.LeadingZeros8(value) == 0,
			Zero:      result == 0,
		}
	case RotateRight:
		if i.WithCopy() {
			result = byte(int8(value) >> 1) // Sign-extend right shift
		} else {
			result = value >> 1
		}

		flags = FlagSet{
			FullCarry: bits.TrailingZeros8(value) == 0,
			Zero:      result == 0,
		}
	}
	return result, flags
}

func (cpu *CPU) perform(f func(...byte) (byte, FlagSet), args ...byte) {
	result, flagSet := f(args...)
	cpu.Set(A, result)
	cpu.setFlags(flagSet)
}

func (cpu *CPU) Run(instr Instruction) {
	switch i := instr.(type) {
	case Move:
		cpu.Set(i.dest, cpu.Get(i.source))
	case MoveImmediate:
		cpu.Set(i.dest, i.immediate)
	case LoadIndirect:
		cpu.Set(i.dest, cpu.GetMem(i.source))
	case StoreIndirect:
		address := mergePair(cpu.GetPair(i.dest))
		cpu.WriteMem(address, cpu.Get(i.source))
	case LoadRelative:
		source := cpu.computeOffset(uint16(cpu.Get(C)))
		cpu.Set(A, cpu.readMem(source))
	case LoadRelativeImmediateN:
		source := cpu.computeOffset(uint16(i.immediate))
		cpu.Set(A, cpu.readMem(source))
	case LoadRelativeImmediateNN:
		cpu.Set(A, cpu.readMem(i.immediate))
	case StoreRelative:
		source := cpu.computeOffset(uint16(cpu.Get(C)))
		cpu.WriteMem(source, cpu.Get(A))
	case StoreRelativeImmediateN:
		dest := cpu.computeOffset(uint16(i.immediate))
		cpu.WriteMem(dest, cpu.Get(A))
	case StoreRelativeImmediateNN:
		cpu.WriteMem(i.immediate, cpu.Get(A))
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
		cpu.SetPair(i.dest, i.immediate)
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
		a := uint16(i.immediate)
		b := cpu.GetSP()
		cpu.SetHL(a + b)
		cpu.incrementCycles() // TODO: remove need for this
		cpu.setFlags(FlagSet{
			HalfCarry: isAddHalfCarry16(a, b),
			FullCarry: isAddFullCarry16(a, b),
		})
	case StoreSP:
		cpu.WriteMem(i.immediate, byte(cpu.GetSP()))
		cpu.WriteMem(i.immediate+1, byte(cpu.GetSP()>>8))
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
			HalfCarry: isAddHalfCarry16(a, uint16(b)),
			FullCarry: isAddFullCarry16(a, uint16(b)),
		})
		cpu.incrementCycles()
	case IncrementPair:
		a := mergePair(cpu.GetPair(i.dest))
		cpu.SetPair(i.dest, a+1)
		cpu.incrementCycles()
	case DecrementPair:
		a := mergePair(cpu.GetPair(i.dest))
		cpu.SetPair(i.dest, a-1)
		cpu.incrementCycles()
	case RotateA:
		result, flagSet := rotateOp(i, cpu.Get(A), cpu.getFlag(FullCarry))
		cpu.Set(A, result)
		cpu.setFlags(flagSet)
	case RotateOperand:
		var result byte
		var flagSet FlagSet
		operand := cpu.fetchAndIncrement()
		i.direction = rotationDirection(operand)
		i.withCopy = rotationCopy(operand)
		i.source = source(operand)
		i.action = rotationAction(operand)

		switch i.action {
		case RotateAction:
			result, flagSet = rotateOp(i, cpu.Get(i.source), cpu.getFlag(FullCarry))
		case ShiftAction:
			result, flagSet = shiftOp(i, cpu.Get(i.source), cpu.getFlag(FullCarry))
		}
		cpu.Set(i.source, result)
		cpu.setFlags(flagSet)
	case InvalidInstruction:
		panic(fmt.Sprintf("Invalid Instruction: %x", instr.Opcode()))
	}
}

func (cpu *CPU) LoadAndRun() {
	Decode(cpu, cpu.Run)
}

func Init() CPU {
	return CPU{
		r: Registers{
			A: 0, B: 0, C: 0, D: 0, E: 0, H: 0, L: 0,
		}, SP: StackStartAddress, PC: ProgramStartAddress,
		memory: InitMemory(),
	}
}

func (cpu *CPU) next() byte {
	value := cpu.memory.get(cpu.GetPC())
	cpu.incrementPC()
	cpu.incrementCycles()
	return value
}

func (cpu *CPU) fetchAndIncrement() byte {
	value := cpu.memory.get(cpu.GetPC())
	cpu.incrementPC()
	cpu.incrementCycles()
	return value
}

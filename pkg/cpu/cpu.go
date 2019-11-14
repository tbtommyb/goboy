package cpu

import (
	"fmt"
	"math/bits"
	"sync"

	"github.com/tbtommyb/goboy/pkg/decoder"
	"github.com/tbtommyb/goboy/pkg/display"
	in "github.com/tbtommyb/goboy/pkg/instructions"
	"github.com/tbtommyb/goboy/pkg/registers"
	"github.com/tbtommyb/goboy/pkg/utils"
)

var GameboyClockSpeed = 4194304

type CPU struct {
	r                    registers.Registers
	flags                byte
	SP, PC               uint16
	memory               *Memory
	cycles               uint
	IME                  bool
	halt                 bool
	Display              *display.Display
	gpu                  *GPU
	loadBIOS             bool
	internalTimer        uint16
	cyclesForCurrentTick int
	currentROMBank       uint16
	interrupts           chan Interrupt
	complete             chan bool
	pcMutex              sync.Mutex
	joypadInternalState  Joypad
}

func (cpu *CPU) GetPC() uint16 {
	return cpu.PC
}

func (cpu *CPU) incrementPC() {
	cpu.pcMutex.Lock()
	cpu.PC += 1
	cpu.pcMutex.Unlock()
}

func (cpu *CPU) setPC(value uint16) {
	cpu.incrementCycles()
	cpu.pcMutex.Lock()
	cpu.PC = value
	cpu.pcMutex.Unlock()
}

func (cpu *CPU) GetCycles() uint {
	return cpu.cycles
}

func (cpu *CPU) incrementCycles() {
	cpu.cycles += 1
}

func (cpu *CPU) decrementCycles() {
	cpu.cycles -= 1
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
	return FlagSet{
		Zero:      a == b,
		Negative:  true,
		HalfCarry: isSubHalfCarry(a, b, 0),
		FullCarry: isSubFullCarry(a, b, 0),
	}
}

func shiftOp(i in.Shift, value, flag byte) (byte, FlagSet) {
	var result byte
	var flags FlagSet
	switch i.GetDirection() {
	case in.Left:
		result = value << 1
		flags = FlagSet{
			FullCarry: bits.LeadingZeros8(value) == 0,
			Zero:      result == 0,
		}
	case in.Right:
		if i.IsWithCopy() {
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

func swapOp(value byte) (byte, FlagSet) {
	result := (value&0xf)<<4 | (value&0xf0)>>4
	flags := FlagSet{
		Zero: result == 0,
	}
	return result, flags
}

func rotateLeftOp(value byte, withCarry bool) (byte, FlagSet) {
	bit7 := value >> 7
	value = value << 1
	if withCarry {
		value = value | 0x1
	}

	var fc bool
	if bit7 != 0 {
		fc = true
	}
	return value, FlagSet{
		Negative:  false,
		HalfCarry: false,
		Zero:      value == 0,
		FullCarry: fc,
	}
}

func rotateLeftCarryOp(value byte) (byte, FlagSet) {
	result := (value << 1) | (value >> 7)
	fc := (value >> 7) > 0

	return result, FlagSet{
		Negative:  false,
		HalfCarry: false,
		Zero:      value == 0,
		FullCarry: fc,
	}
}

func rotateRightOp(value byte, withCarry bool) (byte, FlagSet) {
	var fc bool
	if value&0x1 == 0x1 {
		fc = true
	}

	value = value >> 1
	if withCarry {
		value = value | 0x80
	}
	return value, FlagSet{
		Zero:      value == 0,
		Negative:  false,
		HalfCarry: false,
		FullCarry: fc,
	}
}

func rotateRightCarryOp(value byte) (byte, FlagSet) {
	result := (value << 7) | (value >> 1)
	fc := value&0x1 > 0

	return result, FlagSet{
		Negative:  false,
		HalfCarry: false,
		Zero:      value == 0,
		FullCarry: fc,
	}
}

func (cpu *CPU) perform(f func(...byte) (byte, FlagSet), args ...byte) {
	result, flagSet := f(args...)
	cpu.Set(registers.A, result)
	cpu.setFlags(flagSet)
}

func (cpu *CPU) Execute(instr in.Instruction) {
	switch i := instr.(type) {
	case in.Move:
		cpu.Set(i.Dest, cpu.Get(i.Source))
	case in.MoveImmediate:
		cpu.Set(i.Dest, i.Immediate)
	case in.LoadIndirect:
		cpu.Set(i.Dest, cpu.GetMem(i.Source))
	case in.StoreIndirect:
		address := utils.MergePair(cpu.GetPair(i.Dest))
		cpu.WriteMem(address, cpu.Get(i.Source))
	case in.LoadRelative:
		source := cpu.computeOffset(uint16(cpu.Get(registers.C)))
		cpu.Set(registers.A, cpu.readMem(source))
	case in.LoadRelativeImmediateN:
		source := cpu.computeOffset(uint16(i.Immediate))
		cpu.Set(registers.A, cpu.readMem(source))
	case in.LoadRelativeImmediateNN:
		cpu.Set(registers.A, cpu.readMem(i.Immediate))
	case in.StoreRelative:
		source := cpu.computeOffset(uint16(cpu.Get(registers.C)))
		cpu.WriteMem(source, cpu.Get(registers.A))
	case in.StoreRelativeImmediateN:
		dest := cpu.computeOffset(uint16(i.Immediate))
		cpu.WriteMem(dest, cpu.Get(registers.A))
	case in.StoreRelativeImmediateNN:
		cpu.WriteMem(i.Immediate, cpu.Get(registers.A))
	case in.LoadIncrement:
		cpu.Set(registers.A, cpu.GetMem(registers.HL))
		cpu.SetHL(cpu.GetHL() + 1)
	case in.LoadDecrement:
		cpu.Set(registers.A, cpu.GetMem(registers.HL))
		cpu.SetHL(cpu.GetHL() - 1)
	case in.StoreIncrement:
		cpu.SetMem(registers.HL, cpu.Get(registers.A))
		cpu.SetHL(cpu.GetHL() + 1)
	case in.StoreDecrement:
		cpu.SetMem(registers.HL, cpu.Get(registers.A))
		cpu.SetHL(cpu.GetHL() - 1)
	case in.LoadRegisterPairImmediate:
		cpu.SetPair(i.Dest, i.Immediate)
	case in.HLtoSP:
		cpu.setSP(cpu.GetHL())
	case in.Push:
		high, low := cpu.GetPair(i.Source)
		cpu.pushStack(high)
		cpu.pushStack(low)
		cpu.incrementCycles() // TODO: remove need for this
	case in.Pop:
		low := cpu.popStack()
		high := cpu.popStack()
		cpu.SetPair(i.Dest, utils.MergePair(high, low))
	case in.LoadHLSP:
		a := uint16(i.Immediate)
		b := cpu.GetSP()
		cpu.SetHL(a + b)
		cpu.incrementCycles() // TODO: remove need for this
		cpu.setFlags(FlagSet{
			HalfCarry: lowerByteHalfCarry(byte(a), byte(b)),
			FullCarry: lowerByteFullCarry(byte(a), byte(b)),
		})
	case in.StoreSP:
		cpu.WriteMem(i.Immediate, byte(cpu.GetSP()))
		cpu.WriteMem(i.Immediate+1, byte(cpu.GetSP()>>8))
	case in.Add:
		carry := cpu.carryBit(i.WithCarry, FullCarry)
		cpu.perform(addOp, cpu.Get(registers.A), cpu.Get(i.Source), carry)
	case in.AddImmediate:
		carry := cpu.carryBit(i.WithCarry, FullCarry)
		result, flags := addOp(cpu.Get(registers.A), i.Immediate, carry)
		cpu.Set(registers.A, result)
		cpu.setFlags(flags)
	case in.Subtract:
		carry := cpu.carryBit(i.WithCarry, FullCarry)
		cpu.perform(subOp, cpu.Get(registers.A), cpu.Get(i.Source), carry)
	case in.SubtractImmediate:
		carry := cpu.carryBit(i.WithCarry, FullCarry)
		cpu.perform(subOp, cpu.Get(registers.A), i.Immediate, carry)
	case in.And:
		cpu.perform(andOp, cpu.Get(registers.A), cpu.Get(i.Source))
	case in.AndImmediate:
		cpu.perform(andOp, cpu.Get(registers.A), i.Immediate)
	case in.Or:
		cpu.perform(orOp, cpu.Get(registers.A), cpu.Get(i.Source))
	case in.OrImmediate:
		cpu.perform(orOp, cpu.Get(registers.A), i.Immediate)
	case in.Xor:
		cpu.perform(xorOp, cpu.Get(registers.A), cpu.Get(i.Source))
	case in.XorImmediate:
		cpu.perform(xorOp, cpu.Get(registers.A), i.Immediate)
	case in.Cmp:
		flagSet := cmpOp(cpu.Get(registers.A), cpu.Get(i.Source))
		cpu.setFlags(flagSet)
	case in.CmpImmediate:
		flagSet := cmpOp(cpu.Get(registers.A), i.Immediate)
		cpu.setFlags(flagSet)
	case in.Increment:
		a := cpu.Get(i.Dest)
		result := a + 1
		cpu.Set(i.Dest, result)
		cpu.setFlags(FlagSet{
			Zero:      result == 0,
			HalfCarry: isAddHalfCarry(a, 1, 0),
			FullCarry: cpu.isSet(FullCarry),
			Negative:  false,
		})
	case in.Decrement:
		a := cpu.Get(i.Dest)
		result := a - 1
		cpu.Set(i.Dest, result)
		cpu.setFlags(FlagSet{
			Zero:      result == 0,
			HalfCarry: isSubHalfCarry(a, 1, 0),
			FullCarry: cpu.isSet(FullCarry),
			Negative:  true,
		})
	case in.AddPair:
		a := cpu.GetHL()
		b := utils.MergePair(cpu.GetPair(i.Source))
		result := a + b
		cpu.SetHL(result)
		cpu.setFlags(FlagSet{
			Zero:      cpu.isSet(Zero),
			HalfCarry: isAddHalfCarry16(a, b),
			FullCarry: isAddFullCarry16(a, b),
		})
		cpu.incrementCycles()
	case in.AddSP:
		a := cpu.GetSP()
		b := uint16(i.Immediate)
		result := a + b
		cpu.setSP(result)
		cpu.setFlags(FlagSet{
			HalfCarry: lowerByteHalfCarry(byte(a), byte(b)),
			FullCarry: lowerByteFullCarry(byte(a), byte(b)),
		})
		cpu.incrementCycles()
	case in.IncrementPair:
		a := utils.MergePair(cpu.GetPair(i.Dest))
		cpu.SetPair(i.Dest, a+1)
		cpu.incrementCycles()
	case in.DecrementPair:
		a := utils.MergePair(cpu.GetPair(i.Dest))
		cpu.SetPair(i.Dest, a-1)
		cpu.incrementCycles()
	case in.RL:
		value, flags := rotateLeftOp(cpu.Get(i.Source), cpu.isSet(FullCarry))
		cpu.Set(i.Source, value)
		cpu.setFlags(flags)
	case in.RLA:
		value, flags := rotateLeftOp(cpu.Get(registers.A), cpu.isSet(FullCarry))
		cpu.Set(registers.A, value)
		flags.Zero = false // BLARGG
		cpu.setFlags(flags)
	case in.RLC:
		value, flags := rotateLeftCarryOp(cpu.Get(i.Source))
		cpu.Set(i.Source, value)
		cpu.setFlags(flags)
	case in.RLCA:
		value, flags := rotateLeftCarryOp(cpu.Get(registers.A))
		flags.Zero = false
		cpu.Set(registers.A, value)
		cpu.setFlags(flags)
	case in.RR:
		value, flags := rotateRightOp(cpu.Get(i.Source), cpu.isSet(FullCarry))
		cpu.Set(i.Source, value)
		cpu.setFlags(flags)
	case in.RRA:
		value, flags := rotateRightOp(cpu.Get(registers.A), cpu.isSet(FullCarry))
		flags.Zero = false
		cpu.Set(registers.A, value)
		cpu.setFlags(flags)
	case in.RRC:
		value, flags := rotateRightCarryOp(cpu.Get(i.Source))
		cpu.Set(i.Source, value)
		cpu.setFlags(flags)
	case in.RRCA:
		value, flags := rotateRightCarryOp(cpu.Get(registers.A))
		flags.Zero = false
		cpu.Set(registers.A, value)
		cpu.setFlags(flags)
	case in.Shift:
		result, flagSet := shiftOp(i, cpu.Get(i.Source), cpu.getFlag(FullCarry))
		cpu.Set(i.Source, result)
		cpu.setFlags(flagSet)
	case in.Swap:
		result, flagSet := swapOp(cpu.Get(i.Source))
		cpu.Set(i.Source, result)
		cpu.setFlags(flagSet)
	case in.Bit:
		cpu.setFlags(FlagSet{
			Negative:  false,
			HalfCarry: true,
			Zero:      !utils.IsSet(i.BitNumber, cpu.Get(i.Source)),
			FullCarry: cpu.isSet(FullCarry),
		})
	case in.Set:
		bit := i.BitNumber
		result := utils.SetBit(bit, cpu.Get(i.Source), 1)
		cpu.Set(i.Source, result)
	case in.Reset:
		bit := i.BitNumber
		result := utils.SetBit(bit, cpu.Get(i.Source), 0)
		cpu.Set(i.Source, result)
	case in.JumpImmediate:
		cpu.setPC(i.Immediate)
	case in.JumpImmediateConditional:
		if cpu.conditionMet(i.Condition) {
			cpu.setPC(i.Immediate)
		}
	case in.JumpRelative:
		// -2 to account for decoder having moved past immediate value. Refactor?
		cpu.setPC(cpu.GetPC() - 2 + uint16(i.Immediate))
	case in.JumpRelativeConditional:
		if cpu.conditionMet(i.Condition) {
			cpu.setPC(cpu.GetPC() - 2 + uint16(i.Immediate))
		}
	case in.JumpMemory:
		cpu.setPC(cpu.GetHL())
		cpu.decrementCycles() // TODO: hack attack
	case in.Call:
		high, low := utils.SplitPair(cpu.GetPC())
		cpu.pushStack(high)
		cpu.pushStack(low)
		cpu.setPC(i.Immediate)
	case in.CallConditional:
		if cpu.conditionMet(i.Condition) {
			high, low := utils.SplitPair(cpu.GetPC())
			cpu.pushStack(high)
			cpu.pushStack(low)
			cpu.setPC(i.Immediate)
		}
	case in.Return:
		cpu.setPC(utils.ReverseMergePair(cpu.popStack(), cpu.popStack()))
	case in.ReturnInterrupt:
		cpu.setPC(utils.ReverseMergePair(cpu.popStack(), cpu.popStack()))
		cpu.enableInterrupts()
	case in.ReturnConditional:
		if cpu.conditionMet(i.Condition) {
			cpu.setPC(utils.ReverseMergePair(cpu.popStack(), cpu.popStack()))
		}
		cpu.incrementCycles()
	case in.RST:
		high, low := utils.SplitPair(cpu.GetPC())
		cpu.pushStack(high)
		cpu.pushStack(low)
		cpu.setPC(uint16(i.Operand << in.OperandShift))
	case in.DAA:
		if !cpu.isSet(Negative) {
			if cpu.isSet(FullCarry) || cpu.Get(registers.A) > 0x99 {
				cpu.Set(registers.A, cpu.Get(registers.A)+0x60)
				cpu.setFlag(FullCarry, true)
			}
			if cpu.isSet(HalfCarry) || (cpu.Get(registers.A)&0x0f) > 0x9 {
				cpu.Set(registers.A, cpu.Get(registers.A)+0x6)
			}
		} else {
			if cpu.isSet(FullCarry) {
				cpu.Set(registers.A, cpu.Get(registers.A)-0x60)
				cpu.setFlag(FullCarry, true)
			}
			if cpu.isSet(HalfCarry) {
				cpu.Set(registers.A, cpu.Get(registers.A)-0x6)
			}
		}
		cpu.setFlags(FlagSet{
			Zero:      cpu.Get(registers.A) == 0,
			HalfCarry: false,
			FullCarry: cpu.isSet(FullCarry),
			Negative:  cpu.isSet(Negative),
		})
	case in.Complement:
		cpu.Set(registers.A, ^cpu.Get(registers.A))
		cpu.setFlags(FlagSet{
			Negative:  true,
			HalfCarry: true,
			Zero:      cpu.isSet(Zero),
			FullCarry: cpu.isSet(FullCarry),
		})
	case in.CCF:
		cpu.setFlags(FlagSet{
			FullCarry: !cpu.isSet(FullCarry),
			Negative:  false,
			HalfCarry: false,
			Zero:      cpu.isSet(Zero),
		})
	case in.SCF:
		cpu.setFlags(FlagSet{
			FullCarry: true,
			Negative:  false,
			HalfCarry: false,
			Zero:      cpu.isSet(Zero),
		})
	case in.EnableInterrupt:
		cpu.enableInterrupts()
	case in.DisableInterrupt:
		cpu.disableInterrupts()
	case in.Nop:
		cpu.incrementCycles()
	case in.Stop:
		// TODO: implement
		cpu.halt = true
	case in.Halt:
		cpu.halt = true
	case in.InvalidInstruction:
		fmt.Sprintf("Invalid Instruction: %x", instr.Opcode())
	}
}

func (cpu *CPU) Run() {
	for !cpu.halt {
		cpu.Execute(decoder.Decode(cpu))
	}
	return
}

func (cpu *CPU) Step() uint {
	if cpu.halt {
		return 1 // nop
	}
	// TODO: find more efficient solution
	if cpu.GetPC() == 0x100 && cpu.loadBIOS {
		cpu.loadBIOS = false
	}
	initialCycles := cpu.GetCycles()
	cpu.Execute(decoder.Decode(cpu))
	return cpu.GetCycles() - initialCycles
}

func Init(loadBIOS bool) *CPU {
	cpu := &CPU{
		loadBIOS:       loadBIOS,
		r:              registers.Registers{},
		currentROMBank: 1,
		interrupts:     make(chan Interrupt, len(Interrupts)),
	}
	memory := InitMemory(cpu)
	cpu.memory = memory
	gpu := InitGPU(cpu)
	cpu.gpu = gpu
	if !cpu.loadBIOS {
		cpu.emulateBootSequence()
	}
	return cpu
}

func (cpu *CPU) AttachDisplay(d DisplayInterface) {
	cpu.gpu.display = d
}

func (cpu *CPU) UpdateDisplay() {
	cpu.gpu.update()
}

func (cpu *CPU) Next() byte {
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

func (cpu *CPU) emulateBootSequence() {
	cpu.SetAF(0x01B0)
	cpu.SetBC(0x0013)
	cpu.SetDE(0x00D8)
	cpu.SetHL(0x014D)
	cpu.setSP(0xFFFE)
	cpu.setLCDC(0x91)
	cpu.setBGP(0xFC)
	cpu.setOBP0(0xFF)
	cpu.setOBP1(0xFF)
	cpu.setPC(0x100)
}

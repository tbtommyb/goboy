package instructions

import (
	"github.com/tbtommyb/goboy/pkg/conditions"
	"github.com/tbtommyb/goboy/pkg/registers"
)

type Direction byte

const (
	Left Direction = iota
	Right
)

type Instruction interface {
	Opcode() []byte
}

type InvalidInstruction struct{ ErrorOpcode byte }

func (i InvalidInstruction) Opcode() []byte { return []byte{i.ErrorOpcode} }

type Nop struct{}

func (i Nop) Opcode() []byte { return []byte{NopPattern} }

type Move struct {
	Source, Dest registers.Single
}

func (i Move) Opcode() []byte {
	return []byte{byte(MovePattern | i.Source | i.Dest<<DestRegisterShift)}
}

type MoveImmediate struct {
	Dest      registers.Single
	Immediate byte
}

func (i MoveImmediate) Opcode() []byte {
	return []byte{byte(MoveImmediatePattern | i.Dest<<DestRegisterShift), i.Immediate}
}

type LoadIndirect struct {
	Source registers.Pair
	Dest   registers.Single
}

func (i LoadIndirect) Opcode() []byte {
	return []byte{byte(LoadIndirectPattern | i.Source<<PairRegisterShift)}
}

type StoreIndirect struct {
	Source registers.Single
	Dest   registers.Pair
}

func (i StoreIndirect) Opcode() []byte {
	return []byte{byte(StoreIndirectPattern | i.Dest<<PairRegisterShift)}
}

type LoadRelative struct {
}

func (i LoadRelative) Opcode() []byte {
	return []byte{LoadRelativePattern}
}

type LoadRelativeImmediateN struct {
	Immediate byte
}

func (i LoadRelativeImmediateN) Opcode() []byte {
	return []byte{LoadRelativeImmediateNPattern, i.Immediate}
}

type LoadRelativeImmediateNN struct {
	Immediate uint16
}

func (i LoadRelativeImmediateNN) Opcode() []byte {
	return []byte{LoadRelativeImmediateNNPattern, byte(i.Immediate), byte(i.Immediate >> 8)}
}

type StoreRelative struct {
}

func (i StoreRelative) Opcode() []byte {
	return []byte{StoreRelativePattern}
}

type StoreRelativeImmediateN struct {
	Immediate byte
}

func (i StoreRelativeImmediateN) Opcode() []byte {
	return []byte{StoreRelativeImmediateNPattern, i.Immediate}
}

type StoreRelativeImmediateNN struct {
	Immediate uint16
}

func (i StoreRelativeImmediateNN) Opcode() []byte {
	return []byte{StoreRelativeImmediateNNPattern, byte(i.Immediate), byte(i.Immediate >> 8)}
}

type LoadIncrement struct{}

func (i LoadIncrement) Opcode() []byte {
	return []byte{(LoadIncrementPattern)}
}

type StoreIncrement struct{}

func (i StoreIncrement) Opcode() []byte {
	return []byte{(StoreIncrementPattern)}
}

type LoadDecrement struct{}

func (i LoadDecrement) Opcode() []byte {
	return []byte{(LoadDecrementPattern)}
}

type StoreDecrement struct{}

func (i StoreDecrement) Opcode() []byte {
	return []byte{(StoreDecrementPattern)}
}

// TODO: maybe make this more generic?
type HLtoSP struct{}

func (i HLtoSP) Opcode() []byte {
	return []byte{(HLtoSPPattern)}
}

type LoadRegisterPairImmediate struct {
	Dest      registers.Pair
	Immediate uint16
}

func (i LoadRegisterPairImmediate) Opcode() []byte {
	return []byte{byte(LoadRegisterPairImmediatePattern | i.Dest<<PairRegisterShift), byte(i.Immediate), byte(i.Immediate >> 8)}
}

type Push struct {
	Source registers.Pair
}

func (i Push) Opcode() []byte {
	register := MuxPairs(i.Source)
	return []byte{byte(PushPattern | register<<PairRegisterShift)}
}

type Pop struct {
	Dest registers.Pair
}

func (i Pop) Opcode() []byte {
	register := MuxPairs(i.Dest)
	return []byte{byte(PopPattern | register<<PairRegisterShift)}
}

type LoadHLSP struct {
	Immediate int8
}

func (i LoadHLSP) Opcode() []byte {
	return []byte{byte(LoadHLSPPattern), byte(i.Immediate)}
}

type StoreSP struct {
	Immediate uint16
}

func (i StoreSP) Opcode() []byte {
	return []byte{byte(StoreSPPattern), byte(i.Immediate), byte(i.Immediate >> 8)}
}

type Add struct {
	Source    registers.Single
	WithCarry bool
}

func (i Add) Opcode() []byte {
	var carry byte
	if i.WithCarry {
		carry = 1
	}
	return []byte{byte(AddPattern|i.Source) | carry<<CarryShift}
}

type AddImmediate struct {
	Immediate byte
	WithCarry bool
}

func (i AddImmediate) Opcode() []byte {
	var carry byte
	if i.WithCarry {
		carry = 1
	}
	return []byte{byte(AddImmediatePattern | carry<<CarryShift), i.Immediate}
}

type Subtract struct {
	Source    registers.Single
	WithCarry bool
}

func (i Subtract) Opcode() []byte {
	var carry byte
	if i.WithCarry {
		carry = 1
	}
	return []byte{byte(SubtractPattern|i.Source) | carry<<CarryShift}
}

type SubtractImmediate struct {
	Immediate byte
	WithCarry bool
}

func (i SubtractImmediate) Opcode() []byte {
	var carry byte
	if i.WithCarry {
		carry = 1
	}
	return []byte{byte(SubtractImmediatePattern | carry<<CarryShift), i.Immediate}
}

type And struct {
	Source registers.Single
}

func (i And) Opcode() []byte {
	return []byte{byte(AndPattern | i.Source)}
}

type AndImmediate struct {
	Immediate byte
}

func (i AndImmediate) Opcode() []byte {
	return []byte{AndImmediatePattern, i.Immediate}
}

type Or struct {
	Source registers.Single
}

func (i Or) Opcode() []byte {
	return []byte{byte(OrPattern | i.Source)}
}

type OrImmediate struct {
	Immediate byte
}

func (i OrImmediate) Opcode() []byte {
	return []byte{OrImmediatePattern, i.Immediate}
}

type Xor struct {
	Source registers.Single
}

func (i Xor) Opcode() []byte {
	return []byte{byte(XorPattern | i.Source)}
}

type XorImmediate struct {
	Immediate byte
}

func (i XorImmediate) Opcode() []byte {
	return []byte{XorImmediatePattern, i.Immediate}
}

type Cmp struct {
	Source registers.Single
}

func (i Cmp) Opcode() []byte {
	return []byte{byte(CmpPattern | i.Source)}
}

type CmpImmediate struct {
	Immediate byte
}

func (i CmpImmediate) Opcode() []byte {
	return []byte{CmpImmediatePattern, i.Immediate}
}

type Increment struct {
	Dest registers.Single
}

func (i Increment) Opcode() []byte {
	return []byte{byte(IncrementPattern | i.Dest<<DestRegisterShift)}
}

type Decrement struct {
	Dest registers.Single
}

func (i Decrement) Opcode() []byte {
	return []byte{byte(DecrementPattern | i.Dest<<DestRegisterShift)}
}

type AddPair struct {
	Source registers.Pair
}

func (i AddPair) Opcode() []byte {
	return []byte{byte(AddPairPattern | i.Source<<PairRegisterShift)}
}

type AddSP struct {
	Immediate int8
}

func (i AddSP) Opcode() []byte {
	return []byte{AddSPPattern, byte(i.Immediate)}
}

type IncrementPair struct {
	Dest registers.Pair
}

func (i IncrementPair) Opcode() []byte {
	return []byte{byte(IncrementPairPattern | i.Dest<<PairRegisterShift)}
}

type DecrementPair struct {
	Dest registers.Pair
}

func (i DecrementPair) Opcode() []byte {
	return []byte{byte(DecrementPairPattern | i.Dest<<PairRegisterShift)}
}

type RotateInstruction interface {
	GetDirection() Direction
	IsWithCopy() bool
}

type RLCA struct{}

func (i RLCA) Opcode() []byte { return []byte{RLCAPattern} }

type RLA struct{}

func (i RLA) Opcode() []byte { return []byte{RLAPattern} }

type RRCA struct{}

func (i RRCA) Opcode() []byte { return []byte{RRCAPattern} }

type RRA struct{}

func (i RRA) Opcode() []byte { return []byte{RRAPattern} }

type RLC struct{ Source registers.Single }

func (i RLC) Opcode() []byte { return []byte{Prefix, byte(RLCPattern | i.Source)} }

type RL struct{ Source registers.Single }

func (i RL) Opcode() []byte { return []byte{Prefix, byte(RLPattern | i.Source)} }

type RRC struct{ Source registers.Single }

func (i RRC) Opcode() []byte { return []byte{Prefix, byte(RRCPattern | i.Source)} }

type RR struct{ Source registers.Single }

func (i RR) Opcode() []byte { return []byte{Prefix, byte(RRPattern | i.Source)} }

type Shift struct {
	Direction Direction
	Source    registers.Single
	WithCopy  bool
}

func (i Shift) Opcode() []byte {
	var copyBit byte
	if i.Direction == Right && !i.WithCopy {
		copyBit = 1
	}
	return []byte{Prefix, ShiftPattern | byte(i.Direction<<RotateDirectionShift) | byte(copyBit<<ShiftCopyShift) | byte(i.Source)}
}

func (i Shift) GetDirection() Direction {
	return i.Direction
}

func (i Shift) IsWithCopy() bool {
	return i.WithCopy
}

type Swap struct {
	Source registers.Single
}

func (i Swap) Opcode() []byte {
	return []byte{Prefix, SwapPattern | byte(i.Source)}
}

type Bit struct {
	Source    registers.Single
	BitNumber byte
}

func (i Bit) Opcode() []byte {
	return []byte{Prefix, BitPattern | byte(i.Source) | byte(i.BitNumber<<BitNumberShift)}
}

type Set struct {
	Source    registers.Single
	BitNumber byte
}

func (i Set) Opcode() []byte {
	return []byte{Prefix, SetPattern | byte(i.Source) | byte(i.BitNumber<<BitNumberShift)}
}

type Reset struct {
	Source    registers.Single
	BitNumber byte
}

func (i Reset) Opcode() []byte {
	return []byte{Prefix, ResetPattern | byte(i.Source) | byte(i.BitNumber<<BitNumberShift)}
}

type JumpImmediate struct {
	Immediate uint16
}

func (i JumpImmediate) Opcode() []byte {
	return []byte{JumpImmediatePattern, byte(i.Immediate), byte(i.Immediate >> 8)}
}

type JumpImmediateConditional struct {
	Immediate uint16
	Condition conditions.Condition
}

func (i JumpImmediateConditional) Opcode() []byte {
	return []byte{JumpImmediateConditionalPattern | byte(i.Condition<<ConditionShift), byte(i.Immediate), byte(i.Immediate >> 8)}
}

type JumpRelative struct {
	Immediate int8
}

func (i JumpRelative) Opcode() []byte {
	return []byte{JumpRelativePattern, byte(i.Immediate - 2)}
}

type JumpRelativeConditional struct {
	Immediate int8
	Condition conditions.Condition
}

func (i JumpRelativeConditional) Opcode() []byte {
	return []byte{JumpRelativeConditionalPattern | byte(i.Condition<<ConditionShift), byte(i.Immediate - 2)}
}

type JumpMemory struct{}

func (i JumpMemory) Opcode() []byte {
	return []byte{JumpMemoryPattern}
}

type Call struct {
	Immediate uint16
}

func (i Call) Opcode() []byte {
	return []byte{CallPattern, byte(i.Immediate), byte(i.Immediate >> 8)}
}

type CallConditional struct {
	Immediate uint16
	Condition conditions.Condition
}

func (i CallConditional) Opcode() []byte {
	return []byte{CallConditionalPattern | byte(i.Condition<<ConditionShift), byte(i.Immediate), byte(i.Immediate >> 8)}
}

type Return struct{}

func (i Return) Opcode() []byte {
	return []byte{ReturnPattern}
}

type ReturnInterrupt struct{}

func (i ReturnInterrupt) Opcode() []byte {
	return []byte{ReturnInterruptPattern}
}

type ReturnConditional struct {
	Condition conditions.Condition
}

func (i ReturnConditional) Opcode() []byte {
	return []byte{ReturnConditionalPattern | byte(i.Condition<<ConditionShift)}
}

type RST struct {
	Operand byte
}

func (i RST) Opcode() []byte {
	return []byte{RSTPattern | byte(i.Operand<<3)}
}

type DAA struct{}

func (i DAA) Opcode() []byte {
	return []byte{DAAPattern}
}

type Complement struct{}

func (i Complement) Opcode() []byte {
	return []byte{ComplementPattern}
}

type CCF struct{}

func (i CCF) Opcode() []byte {
	return []byte{CCFPattern}
}

type SCF struct{}

func (i SCF) Opcode() []byte {
	return []byte{SCFPattern}
}

type DisableInterrupt struct{}

func (i DisableInterrupt) Opcode() []byte {
	return []byte{DisableInterruptPattern}
}

type EnableInterrupt struct{}

func (i EnableInterrupt) Opcode() []byte {
	return []byte{EnableInterruptPattern}
}

type Halt struct{}

func (i Halt) Opcode() []byte {
	return []byte{HaltPattern}
}

type Stop struct{}

func (i Stop) Opcode() []byte {
	return []byte{StopPattern, NopPattern}
}

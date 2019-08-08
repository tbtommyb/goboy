package instructions

import "github.com/tbtommyb/goboy/pkg/registers"

type RotationDirection byte
type RotationAction byte

const (
	RotateLeft RotationDirection = iota
	RotateRight
)

const (
	RotateAction RotationAction = iota
	ShiftAction
)

type Instruction interface {
	Opcode() []byte
}

type InvalidInstruction struct{ ErrorOpcode byte }

func (i InvalidInstruction) Opcode() []byte { return []byte{i.ErrorOpcode} }

type EmptyInstruction struct{}

func (i EmptyInstruction) Opcode() []byte { return []byte{0} }

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
	return []byte{LoadRelativeImmediateNNPattern, byte(i.Immediate >> 8), byte(i.Immediate)}
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
	return []byte{StoreRelativeImmediateNNPattern, byte(i.Immediate >> 8), byte(i.Immediate)}
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
	Immediate byte
}

func (i AddSP) Opcode() []byte {
	return []byte{AddSPPattern, i.Immediate}
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
	GetDirection() RotationDirection
	IsWithCopy() bool
}

type RotateA struct {
	Direction RotationDirection
	WithCopy  bool
}

func (i RotateA) Opcode() []byte {
	var copyBit byte = 1
	if i.WithCopy {
		copyBit = 0
	}
	return []byte{byte(RotateAPattern | byte(i.Direction<<RotateDirectionShift) | copyBit<<RotateCopyShift)}
}

func (i RotateA) GetDirection() RotationDirection {
	return i.Direction
}

func (i RotateA) IsWithCopy() bool {
	return i.WithCopy
}

type RotateOperand struct {
	Action    RotationAction
	Direction RotationDirection
	WithCopy  bool
	Source    registers.Single
}

func (i RotateOperand) Opcode() []byte {
	var copyBit byte = 1
	if i.WithCopy {
		copyBit = 0
	}
	return []byte{RotateOperandPrefix, byte(i.Direction<<RotateDirectionShift) | byte(copyBit<<RotateCopyShift) | byte(i.Source) | byte(i.Action<<RotateActionShift)}
}

func (i RotateOperand) GetDirection() RotationDirection {
	return i.Direction
}

func (i RotateOperand) IsWithCopy() bool {
	return i.WithCopy
}

func GetRotationDirection(opcode byte) RotationDirection {
	if opcode&RotateDirectionMask > 0 {
		return RotateRight
	}
	return RotateLeft
}

func GetWithRotationCopy(opcode byte) bool {
	return opcode&RotateCopyMask == 0
}

func GetRotationAction(opcode byte) RotationAction {
	if opcode&RotateActionMask > 0 {
		return ShiftAction
	}
	return RotateAction
}

func Source(opcode byte) registers.Single {
	return registers.Single(opcode & SourceRegisterMask)
}

func Dest(opcode byte) registers.Single {
	return registers.Single(opcode & DestRegisterMask >> DestRegisterShift)
}

func Pair(opcode byte) registers.Pair {
	return registers.Pair(opcode & PairRegisterMask >> PairRegisterShift)
}

// AF and SP use same bit pattern in different instructions
func MuxPairs(r registers.Pair) registers.Pair {
	if r == registers.AF {
		r = registers.SP
	}
	return r
}

func DemuxPairs(opcode byte) registers.Pair {
	reg := Pair(opcode)
	if reg == registers.SP {
		reg = registers.AF
	}
	return reg
}

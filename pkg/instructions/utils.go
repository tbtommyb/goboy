package instructions

import (
	"github.com/tbtommyb/goboy/pkg/conditions"
	"github.com/tbtommyb/goboy/pkg/registers"
)

func GetDirection(opcode byte) Direction {
	if opcode&RotateDirectionMask > 0 {
		return Right
	}
	return Left
}

func GetWithCopyRotation(opcode byte) bool {
	return opcode&RotateCopyMask == 0
}

func GetWithCopyShift(opcode byte) bool {
	return opcode&ShiftCopyMask == ShiftCopyPattern
}

func BitNumber(opcode byte) byte {
	return (opcode & BitNumberMask) >> BitNumberShift
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

func GetCondition(opcode byte) conditions.Condition {
	return conditions.Condition((opcode & ConditionMask) >> ConditionShift)
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

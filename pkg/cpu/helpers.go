package cpu

import (
	"fmt"

	"github.com/tbtommyb/goboy/pkg/conditions"
)

func (cpu *CPU) computeOffset(offset uint16) uint16 {
	return 0xFF00 + offset
}

func (cpu *CPU) carryBit(withCarry bool, flag Flag) byte {
	var carry byte
	if withCarry && cpu.isSet(flag) {
		carry = 1
	}
	return carry
}

func (cpu *CPU) conditionMet(cond conditions.Condition) bool {
	switch cond {
	case conditions.NC:
		if !cpu.isSet(FullCarry) {
			return true
		}
	case conditions.C:
		if cpu.isSet(FullCarry) {
			return true
		}
	case conditions.NZ:
		if !cpu.isSet(Zero) {
			return true
		}
	case conditions.Z:
		if cpu.isSet(Zero) {
			return true
		}
	default:
		panic(fmt.Sprintf("Invalid condition %v", cond))
	}
	return false
}

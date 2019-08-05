package cpu

const MoveMask = 0xC0
const MovePattern = 0x40

const MoveImmediateMask = 0xC7
const MoveImmediatePattern = 0x6

const LoadIncrementPattern = 0x2A
const LoadDecrementPattern = 0x3A
const StoreIncrementPattern = 0x22
const StoreDecrementPattern = 0x32

const MoveIndirectMask = 0xEF
const LoadIndirectPattern = 0xA
const StoreIndirectPattern = 0x2

const DestRegisterMask = 0x38
const DestRegisterShift = 3
const SourceRegisterMask = 0x7
const PairRegisterShift = 4
const PairRegisterMask = 0x30

const MoveRelativeMask = 0xF0
const LoadRelativePattern = 0xF0
const StoreRelativePattern = 0xE0
const MoveRelativeAddressing = 0xF

const LoadRegisterPairImmediateMask = 0xCF
const LoadRegisterPairImmediatePattern = 0x1

const HLtoSPPattern = 0xF9

const PushPopMask = 0xCF
const PushPattern = 0xC5
const PopPattern = 0xC1

const LoadHLSPPattern = 0xF8
const StoreSPPattern = 0x8

const AddMask = 0xF0
const AddPattern = 0x80
const AddImmediateMask = 0xF7
const AddImmediatePattern = 0xC6
const CarryMask = 0x8
const CarryShift = 3

const SubtractMask = 0xF0
const SubtractPattern = 0x90
const SubtractImmediateMask = 0xF7
const SubtractImmediatePattern = 0xD6

const AndMask = 0xF8
const AndPattern = 0xA0
const AndImmediatePattern = 0xE6

const OrMask = 0xF8
const OrPattern = 0xB0
const OrImmediatePattern = 0xF6

const XorMask = 0xF8
const XorPattern = 0xA8
const XorImmediatePattern = 0xEE

const CmpMask = 0xF8
const CmpPattern = 0xB8
const CmpImmediatePattern = 0xFE

func Decode(op byte) Instruction {
	switch {
	case op&MoveMask == MovePattern:
		// LD D, S. 0b01dd dsss
		return Move{source: source(op), dest: dest(op)}
	case op&MoveImmediateMask == MoveImmediatePattern:
		// LD D, n. 0b00dd d110
		return MoveImmediate{dest: dest(op)}
	case op^LoadIncrementPattern == 0:
		// LDI A, (HL) 0b0010 1010
		return LoadIncrement{}
	case op^StoreIncrementPattern == 0:
		// LDI (HL), A. 0b0010 0010
		return StoreIncrement{}
	case op^LoadDecrementPattern == 0:
		// LDD A, (HL) 0b0011 1010
		return LoadDecrement{}
	case op^StoreDecrementPattern == 0:
		// LDD (HL), A 0b0011 0010
		return StoreDecrement{}
	case op&MoveIndirectMask == LoadIndirectPattern:
		// LD, r, (pair). 0b00dd 1010
		return LoadIndirect{dest: A, source: pair(op)}
	case op&MoveIndirectMask == StoreIndirectPattern:
		// LD (pair), r. 0b00ss 0010
		return StoreIndirect{source: A, dest: pair(op)}
	case op^HLtoSPPattern == 0:
		// TODO: ordering dependence with LoadRelativePattern
		// LD SP, HL. 0b 1111 1001
		return HLtoSP{}
	case op&MoveRelativeMask == LoadRelativePattern && isAddressing(op):
		// LD A, (C). 0b1111 0010
		// LD A, n. 0b1111 0000
		// LD A, nn. 0b1111 1010
		return LoadRelative{addressType: addressType(op)}
	case op&MoveRelativeMask == StoreRelativePattern && isAddressing(op):
		// LD (C), A. 0b1110 0010
		// LD n, A. 0b1110 0000
		// LD nn, A. 0b1110 1010
		return StoreRelative{addressType: addressType(op)}
	case op&LoadRegisterPairImmediateMask == LoadRegisterPairImmediatePattern:
		// LD dd, nn. 0b00dd 0001
		return LoadRegisterPairImmediate{dest: pair(op)}
	case op&PushPopMask == PushPattern:
		// PUSH qq. 0b11qq 0101
		return Push{source: demuxPairs(op)}
	case op&PushPopMask == PopPattern:
		// POP qq. 0b11qq 0001
		return Pop{dest: demuxPairs(op)}
	case op == LoadHLSPPattern:
		// LDHL SP, e. 0b1111 1000
		return LoadHLSP{}
	case op == StoreSPPattern:
		// LD nn, SP. 0b0000 1000
		return StoreSP{}
	case op&AddMask == AddPattern:
		// ADD A, r. 0b1000 0rrr
		// ADC A, r. 0b1000 1rrr
		withCarry := (op & CarryMask) > 0
		return Add{source: source(op), withCarry: withCarry}
	case op&AddImmediateMask == AddImmediatePattern:
		// ADD A n. 0b1100 0110
		withCarry := (op & CarryMask) > 0
		return AddImmediate{withCarry: withCarry}
	case op&SubtractMask == SubtractPattern:
		// SUB A, r. 0b1001 0rrr
		// SBC A, r. 0b1001 1rrr
		withCarry := (op & CarryMask) > 0
		return Subtract{source: source(op), withCarry: withCarry}
	case op&SubtractImmediateMask == SubtractImmediatePattern:
		// SUB A n. 0b1101 0110
		// SBC A n. 0b1101 1110
		withCarry := (op & CarryMask) > 0
		return SubtractImmediate{withCarry: withCarry}
	case op&AndMask == AndPattern:
		// AND A r. 0b1010 0rrr
		return And{source: source(op)}
	case op == AndImmediatePattern:
		// AND A n. 0b1110 0110
		return AndImmediate{}
	case op&OrMask == OrPattern:
		// OR A r. 0b1011 0rrr
		return Or{source: source(op)}
	case op == OrImmediatePattern:
		// OR A n. 0b1111 0110
		return OrImmediate{}
	case op&XorMask == XorPattern:
		// XOR A r. 0b1010 1rrr
		return Xor{source: source(op)}
	case op == XorImmediatePattern:
		// OR A n. 0b1110 1110
		return XorImmediate{}
	case op&CmpMask == CmpPattern:
		// CP A r. 0b1011 1rrr
		return Cmp{source: source(op)}
	case op == CmpImmediatePattern:
		// OR A n. 0b1111 1110
		return CmpImmediate{}
	case op == 0:
		return EmptyInstruction{}
	default:
		return InvalidInstruction{opcode: op}
	}
}

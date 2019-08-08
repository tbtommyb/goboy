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

const LoadRelativePattern = 0xF2
const LoadRelativeImmediateNPattern = 0xF0
const LoadRelativeImmediateNNPattern = 0xFA
const StoreRelativePattern = 0xE2
const StoreRelativeImmediateNPattern = 0xE0
const StoreRelativeImmediateNNPattern = 0xEA
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

const IncrementMask = 0xC7
const IncrementPattern = 0x4
const DecrementMask = 0xC7
const DecrementPattern = 0x5

const AddPairMask = 0xCF
const AddPairPattern = 0x9

const AddSPPattern = 0xE8

const IncrementPairMask = 0xCF
const IncrementPairPattern = 0x3
const DecrementPairMask = 0xCF
const DecrementPairPattern = 0xB

const RotateMask = 0xE7
const RotateAPattern = 0x7
const RotateDirectionMask = 0x8
const RotateDirectionShift = 3
const RotateCopyMask = 0x10
const RotateCopyShift = 4
const RotateOperandPrefix = 0xCB
const RotateActionMask = 0x20
const RotateActionShift = 5

// func Decode(op byte) Instruction {
func Decode(il InstructionIterator, handle func(Instruction)) {
	for op := il.next(); op != 0; op = il.next() {
		switch {
		case op&MoveMask == MovePattern:
			// LD D, S. 0b01dd dsss
			handle(Move{source: source(op), dest: dest(op)})
		case op&MoveImmediateMask == MoveImmediatePattern:
			// LD D, n. 0b00dd d110
			handle(MoveImmediate{dest: dest(op), immediate: il.next()})
		case op^LoadIncrementPattern == 0:
			// LDI A, (HL) 0b0010 1010
			handle(LoadIncrement{})
		case op^StoreIncrementPattern == 0:
			// LDI (HL), A. 0b0010 0010
			handle(StoreIncrement{})
		case op^LoadDecrementPattern == 0:
			// LDD A, (HL) 0b0011 1010
			handle(LoadDecrement{})
		case op^StoreDecrementPattern == 0:
			// LDD (HL), A 0b0011 0010
			handle(StoreDecrement{})
		case op&MoveIndirectMask == LoadIndirectPattern:
			// LD, r, (pair). 0b00dd 1010
			handle(LoadIndirect{dest: A, source: pair(op)})
		case op&MoveIndirectMask == StoreIndirectPattern:
			// LD (pair), r. 0b00ss 0010
			handle(StoreIndirect{source: A, dest: pair(op)})
		case op^HLtoSPPattern == 0:
			// LD SP, HL. 0b 1111 1001
			handle(HLtoSP{})
		case op == LoadRelativePattern:
			// LD A, (C). 0b1111 0010
			handle(LoadRelative{})
		case op == LoadRelativeImmediateNPattern:
			// LD A, n. 0b1111 0000
			handle(LoadRelativeImmediateN{immediate: il.next()})
		case op == LoadRelativeImmediateNNPattern:
			// LD A, nn. 0b1111 1010
			immediate := mergePair(il.next(), il.next())
			handle(LoadRelativeImmediateNN{immediate})
		case op == StoreRelativePattern:
			// LD (C), A. 0b1110 0010
			handle(StoreRelative{})
		case op == StoreRelativeImmediateNPattern:
			// LD n, A. 0b1110 0000
			handle(StoreRelativeImmediateN{immediate: il.next()})
		case op == StoreRelativeImmediateNNPattern:
			// LD nn, A. 0b1110 1010
			immediate := mergePair(il.next(), il.next())
			handle(StoreRelativeImmediateNN{immediate})
		case op&LoadRegisterPairImmediateMask == LoadRegisterPairImmediatePattern:
			// LD dd, nn. 0b00dd 0001
			var immediate uint16
			immediate |= uint16(il.next())
			immediate |= uint16(il.next()) << 8
			handle(LoadRegisterPairImmediate{dest: pair(op), immediate: immediate})
		// case op&PushPopMask == PushPattern:
		// 	// PUSH qq. 0b11qq 0101
		// 	handle(Push{source: demuxPairs(op)}
		// case op&PushPopMask == PopPattern:
		// 	// POP qq. 0b11qq 0001
		// 	handle(Pop{dest: demuxPairs(op)}
		// case op == LoadHLSPPattern:
		// 	// LDHL SP, e. 0b1111 1000
		// 	handle(LoadHLSP{}
		// case op == StoreSPPattern:
		// 	// LD nn, SP. 0b0000 1000
		// 	handle(StoreSP{}
		// case op&AddMask == AddPattern:
		// 	// ADD A, r. 0b1000 0rrr
		// 	// ADC A, r. 0b1000 1rrr
		// 	withCarry := (op & CarryMask) > 0
		// 	handle(Add{source: source(op), withCarry: withCarry}
		// case op&AddImmediateMask == AddImmediatePattern:
		// 	// ADD A n. 0b1100 0110
		// 	withCarry := (op & CarryMask) > 0
		// 	handle(AddImmediate{withCarry: withCarry}
		// case op&SubtractMask == SubtractPattern:
		// 	// SUB A, r. 0b1001 0rrr
		// 	// SBC A, r. 0b1001 1rrr
		// 	withCarry := (op & CarryMask) > 0
		// 	handle(Subtract{source: source(op), withCarry: withCarry}
		// case op&SubtractImmediateMask == SubtractImmediatePattern:
		// 	// SUB A n. 0b1101 0110
		// 	// SBC A n. 0b1101 1110
		// 	withCarry := (op & CarryMask) > 0
		// 	handle(SubtractImmediate{withCarry: withCarry}
		// case op&AndMask == AndPattern:
		// 	// AND A r. 0b1010 0rrr
		// 	handle(And{source: source(op)}
		// case op == AndImmediatePattern:
		// 	// AND A n. 0b1110 0110
		// 	handle(AndImmediate{}
		// case op&OrMask == OrPattern:
		// 	// OR A r. 0b1011 0rrr
		// 	handle(Or{source: source(op)}
		// case op == OrImmediatePattern:
		// 	// OR A n. 0b1111 0110
		// 	handle(OrImmediate{}
		// case op&XorMask == XorPattern:
		// 	// XOR A r. 0b1010 1rrr
		// 	handle(Xor{source: source(op)}
		// case op == XorImmediatePattern:
		// 	// OR A n. 0b1110 1110
		// 	handle(XorImmediate{}
		// case op&CmpMask == CmpPattern:
		// 	// CP A r. 0b1011 1rrr
		// 	handle(Cmp{source: source(op)}
		// case op == CmpImmediatePattern:
		// 	// OR A n. 0b1111 1110
		// 	handle(CmpImmediate{}
		// case op&IncrementMask == IncrementPattern:
		// 	// INC r. 0b00rr r100
		// 	handle(Increment{dest: dest(op)}
		// case op&DecrementMask == DecrementPattern:
		// 	// DEC r. 0b00rr r101
		// 	handle(Decrement{dest: dest(op)}
		// case op&AddPairMask == AddPairPattern:
		// 	// ADD HL, ss. 0b00ss 1001
		// 	handle(AddPair{source: pair(op)}
		// case op == AddSPPattern:
		// 	// ADD SP, n. 0b1110 1000
		// 	handle(AddSP{}
		// case op&IncrementPairMask == IncrementPairPattern:
		// 	// INC ss. 0b00ss 0011
		// 	handle(IncrementPair{dest: pair(op)}
		// case op&DecrementPairMask == DecrementPairPattern:
		// 	// INC ss. 0b00ss 1011
		// 	handle(DecrementPair{dest: pair(op)}
		// case op&RotateMask == RotateAPattern:
		// 	// RLCA. 0b0000 0111
		// 	// RLA. 0b00001 0111
		// 	// RRCA. 0b0000 1111
		// 	// RRA. 0b0001 1111
		// 	handle(RotateA{direction: rotationDirection(op), withCopy: rotationCopy(op)}
		// case op == RotateOperandPrefix:
		// 	// RLC r. 0b11000 1011, 0001 0rrr
		// 	// RL r. 0b11000 1011, 0001 0rrr
		// 	// RRC r. 0b11000 1011, 0000 1rrr
		// 	// RR r. 0b11000 1011, 0001 1rrr
		// 	handle(RotateOperand{}
		case op == 0:
			handle(EmptyInstruction{})
		default:
			handle(InvalidInstruction{opcode: op})
		}
	}
}

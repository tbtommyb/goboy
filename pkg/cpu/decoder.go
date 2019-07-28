package cpu

const MoveMask = 0xC0
const MovePattern = 0x40

const MoveImmediateMask = 0xC7
const MoveImmediatePattern = 0x6

const LoadIncrementPattern = 0x2A
const StoreIncrementPattern = 0x22
const LoadDecrementPattern = 0x3A
const StoreDecrementPattern = 0x32

const MoveIndirectMask = 0xEF
const LoadIndirectPattern = 0xA
const StoreIndirectPattern = 0x2

const DestRegisterMask = 0x38
const DestRegisterShift = 3
const SourceRegisterMask = 0x7
const PairRegisterShift = 4
const PairRegisterMask = 0x30
const PairRegisterBaseValue = 0x8 // Used to give pairs unique numbers

const MoveRelative = 0xF0
const LoadRelativePattern = 0xF0
const StoreRelativePattern = 0xE0
const MoveRelativeAddressing = 0xF

const LoadRegisterPairImmediateMask = 0xCF
const LoadRegisterPairImmediatePattern = 0x1

const HLtoSPPattern = 0xF9

const PushPopMask = 0xCF
const PushPattern = 0xC5
const PopPattern = 0xC1

type Instruction interface {
	Opcode() []byte
}

type InvalidInstruction struct{ opcode byte }

func (i InvalidInstruction) Opcode() []byte { return []byte{i.opcode} }

type EmptyInstruction struct{}

func (i EmptyInstruction) Opcode() []byte { return []byte{0} }

type Move struct {
	source, dest Register
}

func (i Move) Opcode() []byte {
	return []byte{byte(MovePattern | i.source | i.dest<<DestRegisterShift)}
}

type MoveImmediate struct {
	dest      Register
	immediate byte
}

func (i MoveImmediate) Opcode() []byte {
	return []byte{byte(MoveImmediatePattern | i.dest<<DestRegisterShift), i.immediate}
}

type MoveIndirect struct {
	source, dest Register
}

func (i MoveIndirect) Opcode() []byte {
	if i.dest == A {
		return []byte{byte(LoadIndirectPattern | (i.source-PairRegisterBaseValue)<<PairRegisterShift)}
	} else {
		return []byte{byte(StoreIndirectPattern | (i.dest-PairRegisterBaseValue)<<PairRegisterShift)}
	}
}

type LoadRelative struct {
	addressType AddressType
	immediate   uint16
}

// TODO: check immediate ordering
func (i LoadRelative) Opcode() []byte {
	opcode := []byte{byte(LoadRelativePattern | i.addressType)}
	switch i.addressType {
	case RelativeN:
		opcode = append(opcode, uint8(i.immediate))
	case RelativeNN:
		opcode = append(opcode, uint8(i.immediate>>8))
		opcode = append(opcode, uint8(i.immediate))
	}
	return opcode
}

type StoreRelative struct {
	addressType AddressType
	immediate   uint16
}

func (i StoreRelative) Opcode() []byte {
	opcode := []byte{byte(StoreRelativePattern | i.addressType)}
	switch i.addressType {
	case RelativeN:
		opcode = append(opcode, uint8(i.immediate))
	case RelativeNN:
		opcode = append(opcode, uint8(i.immediate>>8))
		opcode = append(opcode, uint8(i.immediate))
	}
	return opcode
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
	dest      Register
	immediate uint16
}

func (i LoadRegisterPairImmediate) Opcode() []byte {
	return []byte{byte(LoadRegisterPairImmediatePattern | (i.dest-PairRegisterBaseValue)<<PairRegisterShift), byte(i.immediate), byte(i.immediate >> 8)}
}

type Push struct {
	source Register
}

func (i Push) Opcode() []byte {
	register := muxPairs(i.source)
	return []byte{byte(PushPattern | register<<PairRegisterShift)}
}

type Pop struct {
	dest Register
}

func (i Pop) Opcode() []byte {
	register := muxPairs(i.dest)
	return []byte{byte(PopPattern | register<<PairRegisterShift)}
}

func source(opcode byte) Register {
	return Register(opcode & SourceRegisterMask)
}

func dest(opcode byte) Register {
	return Register(opcode & DestRegisterMask >> DestRegisterShift)
}

func pair(opcode byte) Register {
	return Register((opcode & PairRegisterMask >> PairRegisterShift) | PairRegisterBaseValue)
}

// TODO: janky AF
func muxPairs(r Register) Register {
	if r == AF {
		r = SP
	}
	return r
}

func demuxPairs(opcode byte) Register {
	reg := pair(opcode)
	if reg == SP {
		reg = AF
	}
	return reg
}

func addressType(opcode byte) AddressType {
	return AddressType(opcode & MoveRelativeAddressing)
}

// TODO: reliance on this feels like antipattern
func isAddressing(opcode byte) bool {
	address := addressType(opcode)

	return address == RelativeN || address == RelativeC || address == RelativeNN
}

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
		return MoveIndirect{dest: A, source: pair(op)}
	case op&MoveIndirectMask == StoreIndirectPattern:
		// LD (pair), r. 0b00ss 0010
		return MoveIndirect{source: A, dest: pair(op)}
	case op^HLtoSPPattern == 0:
		// TODO: ordering dependence with LoadRelativePattern
		// LD SP, HL. 0b 1111 1001
		return HLtoSP{}
	case op&MoveRelative == LoadRelativePattern && isAddressing(op):
		// LD A, (C). 0b1111 0010
		// LD A, n. 0b1111 0000
		// LD A, nn. 0b1111 1010
		return LoadRelative{addressType: addressType(op)}
	case op&MoveRelative == StoreRelativePattern && isAddressing(op):
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
	case op == 0:
		return EmptyInstruction{}
	default:
		return InvalidInstruction{opcode: op}
	}
}

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
const AddImmediateMask = 0xC7
const AddImmediatePattern = 0xC6
const CarryMask = 0x8
const CarryShift = 3

type Instruction interface {
	Opcode() []byte
}

type RelativeAddressingInstruction interface {
	AddressType() AddressType
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

type LoadIndirect struct {
	source RegisterPair
	dest   Register
}

func (i LoadIndirect) Opcode() []byte {
	return []byte{byte(LoadIndirectPattern | i.source<<PairRegisterShift)}
}

type StoreIndirect struct {
	source Register
	dest   RegisterPair
}

func (i StoreIndirect) Opcode() []byte {
	return []byte{byte(StoreIndirectPattern | i.dest<<PairRegisterShift)}
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

func (i LoadRelative) AddressType() AddressType {
	return i.addressType
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

func (i StoreRelative) AddressType() AddressType {
	return i.addressType
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
	dest      RegisterPair
	immediate uint16
}

func (i LoadRegisterPairImmediate) Opcode() []byte {
	return []byte{byte(LoadRegisterPairImmediatePattern | i.dest<<PairRegisterShift), byte(i.immediate), byte(i.immediate >> 8)}
}

type Push struct {
	source RegisterPair
}

func (i Push) Opcode() []byte {
	register := muxPairs(i.source)
	return []byte{byte(PushPattern | register<<PairRegisterShift)}
}

type Pop struct {
	dest RegisterPair
}

func (i Pop) Opcode() []byte {
	register := muxPairs(i.dest)
	return []byte{byte(PopPattern | register<<PairRegisterShift)}
}

type LoadHLSP struct {
	immediate int8
}

func (i LoadHLSP) Opcode() []byte {
	return []byte{byte(LoadHLSPPattern), byte(i.immediate)}
}

type StoreSP struct {
	immediate uint16
}

func (i StoreSP) Opcode() []byte {
	return []byte{byte(StoreSPPattern), byte(i.immediate), byte(i.immediate >> 8)}
}

type Add struct {
	source    Register
	withCarry bool
}

func (i Add) Opcode() []byte {
	var carry byte
	if i.withCarry {
		carry = 1
	}
	return []byte{byte(AddPattern|i.source) | carry<<CarryShift}
}

type AddImmediate struct {
	immediate byte
	withCarry bool
}

func (i AddImmediate) Opcode() []byte {
	var carry byte
	if i.withCarry {
		carry = 1
	}
	return []byte{byte(AddImmediatePattern | carry<<CarryShift), i.immediate}
}

func source(opcode byte) Register {
	return Register(opcode & SourceRegisterMask)
}

func dest(opcode byte) Register {
	return Register(opcode & DestRegisterMask >> DestRegisterShift)
}

func pair(opcode byte) RegisterPair {
	return RegisterPair(opcode & PairRegisterMask >> PairRegisterShift)
}

// AF and SP use same bit pattern in different instructions
func muxPairs(r RegisterPair) RegisterPair {
	if r == AF {
		r = SP
	}
	return r
}

func demuxPairs(opcode byte) RegisterPair {
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
	case op == 0:
		return EmptyInstruction{}
	default:
		return InvalidInstruction{opcode: op}
	}
}

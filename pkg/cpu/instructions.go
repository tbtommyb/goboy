package cpu

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

type Subtract struct {
	source    Register
	withCarry bool
}

func (i Subtract) Opcode() []byte {
	var carry byte
	if i.withCarry {
		carry = 1
	}
	return []byte{byte(SubtractPattern|i.source) | carry<<CarryShift}
}

type SubtractImmediate struct {
	immediate byte
	withCarry bool
}

func (i SubtractImmediate) Opcode() []byte {
	var carry byte
	if i.withCarry {
		carry = 1
	}
	return []byte{byte(SubtractImmediatePattern | carry<<CarryShift), i.immediate}
}

type And struct {
	source Register
}

func (i And) Opcode() []byte {
	return []byte{byte(AndPattern | i.source)}
}

type AndImmediate struct {
	immediate byte
}

func (i AndImmediate) Opcode() []byte {
	return []byte{AndImmediatePattern, i.immediate}
}

type Or struct {
	source Register
}

func (i Or) Opcode() []byte {
	return []byte{byte(OrPattern | i.source)}
}

type OrImmediate struct {
	immediate byte
}

func (i OrImmediate) Opcode() []byte {
	return []byte{OrImmediatePattern, i.immediate}
}

type Xor struct {
	source Register
}

func (i Xor) Opcode() []byte {
	return []byte{byte(XorPattern | i.source)}
}

type XorImmediate struct {
	immediate byte
}

func (i XorImmediate) Opcode() []byte {
	return []byte{XorImmediatePattern, i.immediate}
}

type Cmp struct {
	source Register
}

func (i Cmp) Opcode() []byte {
	return []byte{byte(CmpPattern | i.source)}
}

type CmpImmediate struct {
	immediate byte
}

func (i CmpImmediate) Opcode() []byte {
	return []byte{CmpImmediatePattern, i.immediate}
}

type Increment struct {
	dest Register
}

func (i Increment) Opcode() []byte {
	return []byte{byte(IncrementPattern | i.dest<<DestRegisterShift)}
}

type Decrement struct {
	dest Register
}

func (i Decrement) Opcode() []byte {
	return []byte{byte(DecrementPattern | i.dest<<DestRegisterShift)}
}

type AddPair struct {
	source RegisterPair
}

func (i AddPair) Opcode() []byte {
	return []byte{byte(AddPairPattern | i.source<<PairRegisterShift)}
}

type AddSP struct {
	immediate byte
}

func (i AddSP) Opcode() []byte {
	return []byte{AddSPPattern, i.immediate}
}

type IncrementPair struct {
	dest RegisterPair
}

func (i IncrementPair) Opcode() []byte {
	return []byte{byte(IncrementPairPattern | i.dest<<PairRegisterShift)}
}

type DecrementPair struct {
	dest RegisterPair
}

func (i DecrementPair) Opcode() []byte {
	return []byte{byte(DecrementPairPattern | i.dest<<PairRegisterShift)}
}

type RotateLeftCopyA struct{}

func (i RotateLeftCopyA) Opcode() []byte {
	return []byte{RotateLeftCopyAPattern}
}

package cpu

const LoadMask = 0xC0
const LoadPattern = 0x40

const LoadImmediateMask = 0xC7
const LoadImmediatePattern = 0x6

const DestRegisterMask = 0x38
const DestRegisterShift = 3
const SourceRegisterMask = 0x7

const LoadStorePairMask = 0xEF
const LoadPairPattern = 0xA
const StorePairPattern = 0x2

const PairRegisterShift = 4
const PairRegisterMask = 0x30
const PairRegisterBaseValue = 0x8 // Used to give pairs unique numbers

const LoadStoreRelativeMask = 0xF8
const LoadRelativePattern = 0xF0
const StoreRelativePattern = 0xE0
const LoadStoreRelativeType = 0x7

const LoadNNPattern = 0xFA
const StoreNNPattern = 0xEA

const LoadIncrementMask = 0xEF
const LoadIncrementPattern = 0x2A
const LoadIncrementBitMask = 0x10
const LoadIncrementBitShift = 4
const StoreIncrementPattern = 0x22

type Instruction interface {
	Opcode() []byte
}

type InvalidInstruction struct{ opcode byte }

func (i InvalidInstruction) Opcode() []byte { return []byte{i.opcode} }

type EmptyInstruction struct{}

func (i EmptyInstruction) Opcode() []byte { return []byte{0} }

type Load struct {
	source, dest Register
}

func (i Load) Opcode() []byte {
	return []byte{byte(LoadPattern | i.source | i.dest<<DestRegisterShift)}
}

type LoadImmediate struct {
	dest      Register
	immediate byte
}

func (i LoadImmediate) Opcode() []byte {
	return []byte{byte(LoadImmediatePattern | i.dest<<DestRegisterShift), i.immediate}
}

type LoadPair struct {
	source, dest Register
}

func (i LoadPair) Opcode() []byte {
	if i.dest == A {
		return []byte{byte(LoadPairPattern | (i.source-PairRegisterBaseValue)<<PairRegisterShift)}
	} else {
		return []byte{byte(StorePairPattern | (i.dest-PairRegisterBaseValue)<<PairRegisterShift)}
	}
}

type LoadRelative struct {
	addressType AddressType
	immediate   byte
}

func (i LoadRelative) Opcode() []byte {
	opcode := []byte{byte(LoadRelativePattern | i.addressType)}
	if i.addressType == RelativeN {
		opcode = append(opcode, i.immediate)
	}
	return opcode
}

type StoreRelative struct {
	addressType AddressType
	immediate   byte
}

func (i StoreRelative) Opcode() []byte {
	opcode := []byte{byte(StoreRelativePattern | i.addressType)}
	if i.addressType == RelativeN {
		opcode = append(opcode, i.immediate)
	}
	return opcode
}

type LoadNN struct{ immediate uint16 }

func (i LoadNN) Opcode() []byte {
	return []byte{byte(LoadNNPattern), byte(i.immediate >> 8), byte(i.immediate)}
}

type StoreNN struct{ immediate uint16 }

func (i StoreNN) Opcode() []byte {
	return []byte{byte(StoreNNPattern), byte(i.immediate >> 8), byte(i.immediate)}
}

type LoadIncrement struct {
	increment int
}

func (i LoadIncrement) Opcode() []byte {
	var incrementPattern byte
	if i.increment == 1 {
		incrementPattern = 0
	} else {
		incrementPattern = 1
	}
	return []byte{byte(LoadIncrementPattern | incrementPattern<<LoadIncrementBitShift)}
}

type StoreIncrement struct {
	increment int
}

func (i StoreIncrement) Opcode() []byte {
	var incrementPattern byte
	if i.increment == 1 {
		incrementPattern = 0
	} else {
		incrementPattern = 1
	}
	return []byte{byte(StoreIncrementPattern | incrementPattern<<LoadIncrementBitShift)}
}

func Decode(op byte) Instruction {
	switch {
	case op&LoadMask == LoadPattern:
		// LD D, S. 0b01dddsss
		source := Register(op & SourceRegisterMask)
		dest := Register(op & DestRegisterMask >> DestRegisterShift)
		return Load{source, dest}
	case op&LoadImmediateMask == LoadImmediatePattern:
		// LD D, n. 0b00ddd110
		dest := Register(op & DestRegisterMask >> DestRegisterShift) // TODO: extract this
		return LoadImmediate{dest: dest}
	case op&LoadStorePairMask == LoadPairPattern:
		// LD, r, (pair). 0b00dd1010
		// TODO: cover HLI and HLD with this instruction
		pair := Register((op & PairRegisterMask >> PairRegisterShift) | PairRegisterBaseValue)
		return LoadPair{dest: A, source: pair}
	case op&LoadStorePairMask == StorePairPattern:
		// LD (pair), r. 0b00ss0010
		pair := Register((op & PairRegisterMask >> PairRegisterShift) | PairRegisterBaseValue)
		return LoadPair{source: A, dest: pair}
	case op&LoadStoreRelativeMask == LoadRelativePattern:
		// LD A, (C). 0b11110010
		// LD A, n. 0b11110000
		addressType := AddressType(op & LoadStoreRelativeType)
		return LoadRelative{addressType: addressType}
	case op&LoadStoreRelativeMask == StoreRelativePattern:
		// LD A, (C). 0b11110010
		// LD A, n. 0b11100000
		addressType := AddressType(op & LoadStoreRelativeType)
		return StoreRelative{addressType: addressType}
	case op^LoadNNPattern == 0:
		// LD A, (nn). 0b11111010
		return LoadNN{}
	case op^StoreNNPattern == 0:
		// LD (nn), A. 0b11101010
		return StoreNN{}
	case op&LoadIncrementMask == LoadIncrementPattern:
		// LD A, (HLI) 0b001i1010
		// TODO: handle this bit better
		var increment int
		isDecrement := (op&LoadIncrementBitMask)>>LoadIncrementBitShift == 1
		if isDecrement {
			increment = -1
		} else {
			increment = 1
		}
		return LoadIncrement{increment}
	case op&LoadIncrementMask == StoreIncrementPattern:
		// LD (HLI), A. 0b001i0010
		var increment int
		isDecrement := (op&LoadIncrementBitMask)>>LoadIncrementBitShift == 1
		if isDecrement {
			increment = -1
		} else {
			increment = 1
		}
		return StoreIncrement{increment}
	case op == 0:
		return EmptyInstruction{}
	default:
		return InvalidInstruction{opcode: op}
	}
}

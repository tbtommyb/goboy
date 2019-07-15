package cpu

const LoadMask = 0xC0
const LoadPattern = 0x40

const LoadImmediateMask = 0xC7
const LoadImmediatePattern = 0x6

const DestRegisterMask = 0x38
const DestRegisterShift = 3
const SourceRegisterMask = 0x7

const LoadStorePairMask = 0xCF
const LoadPairPattern = 0xA
const StorePairPattern = 0x2

const PairRegisterShift = 4
const PairRegisterMask = 0x30
const PairRegisterBaseValue = 0x8 // Used to give pairs unique numbers

const LoadRelativeCPattern = 0xF2
const LoadRelativeNPattern = 0xF0
const StoreRelativeCPattern = 0xE2
const StoreRelativeNPattern = 0xE0

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

type LoadRelativeC struct{}

func (i LoadRelativeC) Opcode() []byte { return []byte{byte(LoadRelativeCPattern)} }

type StoreRelativeC struct{}

func (i StoreRelativeC) Opcode() []byte { return []byte{byte(StoreRelativeCPattern)} }

type LoadRelativeN struct{ immediate byte }

func (i LoadRelativeN) Opcode() []byte { return []byte{byte(LoadRelativeNPattern), i.immediate} }

type StoreRelativeN struct{ immediate byte }

func (i StoreRelativeN) Opcode() []byte { return []byte{byte(StoreRelativeNPattern), i.immediate} }

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
		// LD, r, (pair). 0b00ss1010
		// TODO: cover HLI and HLD with this instruction
		pair := Register((op & PairRegisterMask >> PairRegisterShift) | PairRegisterBaseValue)
		return LoadPair{dest: A, source: pair}
	case op&LoadStorePairMask == StorePairPattern:
		pair := Register((op & PairRegisterMask >> PairRegisterShift) | PairRegisterBaseValue)
		return LoadPair{source: A, dest: pair}
	case op^LoadRelativeCPattern == 0:
		return LoadRelativeC{}
	case op^StoreRelativeCPattern == 0:
		return StoreRelativeC{}
	case op^LoadRelativeNPattern == 0:
		return LoadRelativeN{}
	case op^StoreRelativeNPattern == 0:
		return StoreRelativeN{}
	case op == 0:
		return EmptyInstruction{}
	default:
		return InvalidInstruction{opcode: op}
	}
}

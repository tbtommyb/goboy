package cpu

const LoadMask = 0xC0
const LoadPattern = 0x40
const LoadImmediateMask = 0xC7
const LoadImmediatePattern = 0x6
const LoadStorePairMask = 0xCF
const LoadPairPattern = 0xA
const StorePairPattern = 0x2
const LoadOrStorePairFlag = 0x8
const LoadOrStorePairShift = 3

const DestRegisterMask = 0x38
const DestRegisterShift = 3
const SourceRegisterMask = 0x7

const PairRegisterShift = 4
const PairRegisterMask = 0x30
const PairRegisterBaseValue = 0x8 // Used to give pairs unique numbers

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
	case op == 0:
		return EmptyInstruction{}
	default:
		return InvalidInstruction{opcode: op}
	}
}

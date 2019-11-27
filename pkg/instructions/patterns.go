package instructions

const DestRegisterMask = 0x38
const DestRegisterShift = 3
const SourceRegisterMask = 0x7
const PairRegisterShift = 4
const PairRegisterMask = 0x30

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

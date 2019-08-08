package cpu

import (
	"testing"
)

type DoubleOpcodeTestCase struct {
	opcodes  []byte
	expected Instruction
}

func compare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, el := range a {
		if el != b[i] {
			return false
		}
	}
	return true
}

var testCases = []DoubleOpcodeTestCase{
	{[]byte{0x6, 0x12}, MoveImmediate{dest: B, immediate: 0x12}},
	{[]byte{0x36, 0x34}, MoveImmediate{dest: M, immediate: 0x34}},
	{[]byte{0xF2}, LoadRelative{}},
	{[]byte{0xF0, 0x11}, LoadRelativeImmediateN{immediate: 0x11}},
	{[]byte{0xFA, 0x11, 0x22}, LoadRelativeImmediateNN{immediate: 0x1122}},
	{[]byte{0xE2}, StoreRelative{}},
	{[]byte{0xE0, 0x11}, StoreRelativeImmediateN{immediate: 0x11}},
	{[]byte{0xEA, 0x11, 0x22}, StoreRelativeImmediateNN{immediate: 0x1122}},
	{[]byte{0xFF}, InvalidInstruction{opcode: 0xFF}},
	{[]byte{0x00}, EmptyInstruction{}},
	{[]byte{0x77}, Move{source: A, dest: M}},
	{[]byte{0x46}, Move{source: M, dest: B}},
	{[]byte{0x47}, Move{source: A, dest: B}},
	{[]byte{0xA}, LoadIndirect{dest: A, source: BC}},
	{[]byte{0x1A}, LoadIndirect{dest: A, source: DE}},
	{[]byte{0x2}, StoreIndirect{dest: BC, source: A}},
	{[]byte{0x12}, StoreIndirect{dest: DE, source: A}},
	{[]byte{0x2A}, LoadIncrement{}},
	{[]byte{0x3A}, LoadDecrement{}},
	{[]byte{0x22}, StoreIncrement{}},
	{[]byte{0x32}, StoreDecrement{}},
	{[]byte{0x1, 0x34, 0x12}, LoadRegisterPairImmediate{dest: BC, immediate: 0x1234}},
	{[]byte{0x11, 0x34, 0x12}, LoadRegisterPairImmediate{dest: DE, immediate: 0x1234}},
	{[]byte{0x21, 0x34, 0x12}, LoadRegisterPairImmediate{dest: HL, immediate: 0x1234}},
	{[]byte{0x31, 0x34, 0x12}, LoadRegisterPairImmediate{dest: SP, immediate: 0x1234}},
	{[]byte{0xF9}, HLtoSP{}},
	{[]byte{0xC5}, Push{source: BC}},
	{[]byte{0xD5}, Push{source: DE}},
	{[]byte{0xE5}, Push{source: HL}},
	{[]byte{0xF5}, Push{source: AF}},
	{[]byte{0xC1}, Pop{dest: BC}},
	{[]byte{0xD1}, Pop{dest: DE}},
	{[]byte{0xE1}, Pop{dest: HL}},
	{[]byte{0xF1}, Pop{dest: AF}},
	{[]byte{0xF8, 0x12}, LoadHLSP{immediate: 0x12}},
	{[]byte{0x8, 0x44, 0x55}, StoreSP{immediate: 0x5544}},
}

func TestDecoder(t *testing.T) {
	for _, testCase := range testCases {
		il := InstructionList{instructions: testCase.opcodes}
		Decode(&il, func(actual Instruction) {
			if actual != testCase.expected {
				t.Errorf("Expected %#v, got %#v", testCase.expected, actual)
			}
		})
	}
}

func TestOpcode(t *testing.T) {
	for _, testCase := range testCases {
		il := InstructionList{instructions: testCase.opcodes}
		Decode(&il, func(actual Instruction) {
			result := actual.Opcode()
			ok := compare(result, testCase.opcodes)
			if !ok {
				t.Errorf("Expected %#v, got %#v", testCase.opcodes, result)
			}
		})
	}
}

// var testCases = map[byte]Instruction{
// 0x8:  StoreSP{},
// 0x81: Add{source: C},
// 0x8F: Add{source: A, withCarry: true},
// 0x8E: Add{source: M, withCarry: true},
// 0xC6: AddImmediate{},
// 0xCE: AddImmediate{withCarry: true},
// 0x91: Subtract{source: C},
// 0x96: Subtract{source: M},
// 0xD6: SubtractImmediate{},
// 0x99: Subtract{source: C, withCarry: true},
// 0x9E: Subtract{source: M, withCarry: true},
// 0xDE: SubtractImmediate{withCarry: true},
// 0xA1: And{source: C},
// 0xA6: And{source: M},
// 0xE6: AndImmediate{},
// 0xB1: Or{source: C},
// 0xB6: Or{source: M},
// 0xF6: OrImmediate{},
// 0xA9: Xor{source: C},
// 0xAE: Xor{source: M},
// 0xEE: XorImmediate{},
// 0xB9: Cmp{source: C},
// 0xBE: Cmp{source: M},
// 0xFE: CmpImmediate{},
// 0xC:  Increment{dest: C},
// 0x34: Increment{dest: M},
// 0xD:  Decrement{dest: C},
// 0x35: Decrement{dest: M},
// 0x29: AddPair{source: HL},
// 0xE8: AddSP{},
// 0x33: IncrementPair{dest: SP},
// 0x3B: DecrementPair{dest: SP},
// 0x7:  RotateA{withCopy: true, direction: RotateLeft},
// 0x17: RotateA{direction: RotateLeft},
// 0xF:  RotateA{withCopy: true, direction: RotateRight},
// 0x1F: RotateA{direction: RotateRight},
// 0xCB: RotateOperand{},

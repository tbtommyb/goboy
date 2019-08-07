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
	{[]byte{0x8, 0xEE, 0xFF}, StoreSP{immediate: 0xFFEE}},
	{[]byte{0x81}, Add{source: C}},
	{[]byte{0x8F}, Add{source: A, withCarry: true}},
	{[]byte{0x8E}, Add{source: M, withCarry: true}},
	{[]byte{0xC6, 0x23}, AddImmediate{immediate: 0x23}},
	{[]byte{0xCE, 0x23}, AddImmediate{withCarry: true, immediate: 0x23}},
	{[]byte{0x91}, Subtract{source: C}},
	{[]byte{0x96}, Subtract{source: M}},
	{[]byte{0xD6, 0x1}, SubtractImmediate{immediate: 0x1}},
	{[]byte{0x99}, Subtract{source: C, withCarry: true}},
	{[]byte{0x9E}, Subtract{source: M, withCarry: true}},
	{[]byte{0xDE, 0x9}, SubtractImmediate{withCarry: true, immediate: 0x9}},
	{[]byte{0xA1}, And{source: C}},
	{[]byte{0xA6}, And{source: M}},
	{[]byte{0xE6, 0x10}, AndImmediate{immediate: 0x10}},
	{[]byte{0xB1}, Or{source: C}},
	{[]byte{0xB6}, Or{source: M}},
	{[]byte{0xF6, 0xAA}, OrImmediate{immediate: 0xAA}},
	{[]byte{0xA9}, Xor{source: C}},
	{[]byte{0xAE}, Xor{source: M}},
	{[]byte{0xEE, 0xF}, XorImmediate{immediate: 0xF}},
	{[]byte{0xB9}, Cmp{source: C}},
	{[]byte{0xBE}, Cmp{source: M}},
	{[]byte{0xFE, 0xAB}, CmpImmediate{immediate: 0xAB}},
	{[]byte{0xC}, Increment{dest: C}},
	{[]byte{0x34}, Increment{dest: M}},
	{[]byte{0xD}, Decrement{dest: C}},
	{[]byte{0x35}, Decrement{dest: M}},
	{[]byte{0x29}, AddPair{source: HL}},
	{[]byte{0xE8, 0x5}, AddSP{immediate: 0x5}},
	{[]byte{0x33}, IncrementPair{dest: SP}},
	{[]byte{0x3B}, DecrementPair{dest: SP}},
	{[]byte{0x7}, RotateA{withCopy: true, direction: RotateLeft}},
	{[]byte{0x17}, RotateA{direction: RotateLeft}},
	{[]byte{0xF}, RotateA{withCopy: true, direction: RotateRight}},
	{[]byte{0x1F}, RotateA{direction: RotateRight}},
	{[]byte{0xCB, 0x9}, RotateOperand{withCopy: true, direction: RotateRight, source: C}},
	{[]byte{0xCB, 0x16}, RotateOperand{withCopy: false, direction: RotateLeft, source: M}},
	{[]byte{0xCB, 0x29}, RotateOperand{action: ShiftAction, withCopy: true, direction: RotateRight, source: C}},
	{[]byte{0xCB, 0x3E}, RotateOperand{action: ShiftAction, withCopy: false, direction: RotateRight, source: M}},
	{[]byte{0xCB, 0x26}, RotateOperand{action: ShiftAction, withCopy: true, direction: RotateLeft, source: M}},
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

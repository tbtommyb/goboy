package decoder

import (
	"testing"

	"github.com/tbtommyb/goboy/pkg/disassembler"
	in "github.com/tbtommyb/goboy/pkg/instructions"
	"github.com/tbtommyb/goboy/pkg/registers"
)

type DoubleOpcodeTestCase struct {
	opcodes  []byte
	expected in.Instruction
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
	{[]byte{0x6, 0x12}, in.MoveImmediate{Dest: registers.B, Immediate: 0x12}},
	{[]byte{0x36, 0x34}, in.MoveImmediate{Dest: registers.M, Immediate: 0x34}},
	{[]byte{0xF2}, in.LoadRelative{}},
	{[]byte{0xF0, 0x11}, in.LoadRelativeImmediateN{Immediate: 0x11}},
	{[]byte{0xFA, 0x11, 0x22}, in.LoadRelativeImmediateNN{Immediate: 0x1122}},
	{[]byte{0xE2}, in.StoreRelative{}},
	{[]byte{0xE0, 0x11}, in.StoreRelativeImmediateN{Immediate: 0x11}},
	{[]byte{0xEA, 0x11, 0x22}, in.StoreRelativeImmediateNN{Immediate: 0x1122}},
	{[]byte{0xFF}, in.InvalidInstruction{ErrorOpcode: 0xFF}},
	{[]byte{0x00}, in.EmptyInstruction{}},
	{[]byte{0x77}, in.Move{Source: registers.A, Dest: registers.M}},
	{[]byte{0x46}, in.Move{Source: registers.M, Dest: registers.B}},
	{[]byte{0x47}, in.Move{Source: registers.A, Dest: registers.B}},
	{[]byte{0xA}, in.LoadIndirect{Dest: registers.A, Source: registers.BC}},
	{[]byte{0x1A}, in.LoadIndirect{Dest: registers.A, Source: registers.DE}},
	{[]byte{0x2}, in.StoreIndirect{Dest: registers.BC, Source: registers.A}},
	{[]byte{0x12}, in.StoreIndirect{Dest: registers.DE, Source: registers.A}},
	{[]byte{0x2A}, in.LoadIncrement{}},
	{[]byte{0x3A}, in.LoadDecrement{}},
	{[]byte{0x22}, in.StoreIncrement{}},
	{[]byte{0x32}, in.StoreDecrement{}},
	{[]byte{0x1, 0x34, 0x12}, in.LoadRegisterPairImmediate{Dest: registers.BC, Immediate: 0x1234}},
	{[]byte{0x11, 0x34, 0x12}, in.LoadRegisterPairImmediate{Dest: registers.DE, Immediate: 0x1234}},
	{[]byte{0x21, 0x34, 0x12}, in.LoadRegisterPairImmediate{Dest: registers.HL, Immediate: 0x1234}},
	{[]byte{0x31, 0x34, 0x12}, in.LoadRegisterPairImmediate{Dest: registers.SP, Immediate: 0x1234}},
	{[]byte{0xF9}, in.HLtoSP{}},
	{[]byte{0xC5}, in.Push{Source: registers.BC}},
	{[]byte{0xD5}, in.Push{Source: registers.DE}},
	{[]byte{0xE5}, in.Push{Source: registers.HL}},
	{[]byte{0xF5}, in.Push{Source: registers.AF}},
	{[]byte{0xC1}, in.Pop{Dest: registers.BC}},
	{[]byte{0xD1}, in.Pop{Dest: registers.DE}},
	{[]byte{0xE1}, in.Pop{Dest: registers.HL}},
	{[]byte{0xF1}, in.Pop{Dest: registers.AF}},
	{[]byte{0xF8, 0x12}, in.LoadHLSP{Immediate: 0x12}},
	{[]byte{0x8, 0x44, 0x55}, in.StoreSP{Immediate: 0x5544}},
	{[]byte{0x8, 0xEE, 0xFF}, in.StoreSP{Immediate: 0xFFEE}},
	{[]byte{0x81}, in.Add{Source: registers.C}},
	{[]byte{0x8F}, in.Add{Source: registers.A, WithCarry: true}},
	{[]byte{0x8E}, in.Add{Source: registers.M, WithCarry: true}},
	{[]byte{0xC6, 0x23}, in.AddImmediate{Immediate: 0x23}},
	{[]byte{0xCE, 0x23}, in.AddImmediate{WithCarry: true, Immediate: 0x23}},
	{[]byte{0x91}, in.Subtract{Source: registers.C}},
	{[]byte{0x96}, in.Subtract{Source: registers.M}},
	{[]byte{0xD6, 0x1}, in.SubtractImmediate{Immediate: 0x1}},
	{[]byte{0x99}, in.Subtract{Source: registers.C, WithCarry: true}},
	{[]byte{0x9E}, in.Subtract{Source: registers.M, WithCarry: true}},
	{[]byte{0xDE, 0x9}, in.SubtractImmediate{WithCarry: true, Immediate: 0x9}},
	{[]byte{0xA1}, in.And{Source: registers.C}},
	{[]byte{0xA6}, in.And{Source: registers.M}},
	{[]byte{0xE6, 0x10}, in.AndImmediate{Immediate: 0x10}},
	{[]byte{0xB1}, in.Or{Source: registers.C}},
	{[]byte{0xB6}, in.Or{Source: registers.M}},
	{[]byte{0xF6, 0xAA}, in.OrImmediate{Immediate: 0xAA}},
	{[]byte{0xA9}, in.Xor{Source: registers.C}},
	{[]byte{0xAE}, in.Xor{Source: registers.M}},
	{[]byte{0xEE, 0xF}, in.XorImmediate{Immediate: 0xF}},
	{[]byte{0xB9}, in.Cmp{Source: registers.C}},
	{[]byte{0xBE}, in.Cmp{Source: registers.M}},
	{[]byte{0xFE, 0xAB}, in.CmpImmediate{Immediate: 0xAB}},
	{[]byte{0xC}, in.Increment{Dest: registers.C}},
	{[]byte{0x34}, in.Increment{Dest: registers.M}},
	{[]byte{0xD}, in.Decrement{Dest: registers.C}},
	{[]byte{0x35}, in.Decrement{Dest: registers.M}},
	{[]byte{0x29}, in.AddPair{Source: registers.HL}},
	{[]byte{0xE8, 0x5}, in.AddSP{Immediate: 0x5}},
	{[]byte{0x33}, in.IncrementPair{Dest: registers.SP}},
	{[]byte{0x3B}, in.DecrementPair{Dest: registers.SP}},
	{[]byte{0x7}, in.RotateA{WithCopy: true, Direction: in.Left}},
	{[]byte{0x17}, in.RotateA{Direction: in.Left}},
	{[]byte{0xF}, in.RotateA{WithCopy: true, Direction: in.Right}},
	{[]byte{0x1F}, in.RotateA{Direction: in.Right}},
	{[]byte{0xCB, 0x9}, in.RotateOperand{WithCopy: true, Direction: in.Right, Source: registers.C}},
	{[]byte{0xCB, 0x16}, in.RotateOperand{WithCopy: false, Direction: in.Left, Source: registers.M}},
	{[]byte{0xCB, 0x29}, in.Shift{WithCopy: true, Direction: in.Right, Source: registers.C}},
	{[]byte{0xCB, 0x3E}, in.Shift{Direction: in.Right, Source: registers.M}},
	{[]byte{0xCB, 0x26}, in.Shift{Direction: in.Left, Source: registers.M}},
	{[]byte{0xCB, 0x36}, in.Swap{Source: registers.M}},
}

func TestDecoder(t *testing.T) {
	for _, testCase := range testCases {
		il := disassembler.Disassembler{Instructions: testCase.opcodes}
		Decode(&il, func(actual in.Instruction) {
			if actual != testCase.expected {
				t.Errorf("Expected %#v, got %#v", testCase.expected, actual)
			}
		})
	}
}

func TestOpcode(t *testing.T) {
	for _, testCase := range testCases {
		il := disassembler.Disassembler{Instructions: testCase.opcodes}
		Decode(&il, func(actual in.Instruction) {
			result := actual.Opcode()
			ok := compare(result, testCase.opcodes)
			if !ok {
				t.Errorf("Expected %#v, got %#v", testCase.opcodes, result)
			}
		})
	}
}

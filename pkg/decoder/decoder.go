package decoder

import (
	in "github.com/tbtommyb/goboy/pkg/instructions"
	"github.com/tbtommyb/goboy/pkg/registers"
	"github.com/tbtommyb/goboy/pkg/utils"
)

type Iterator interface {
	Next() byte
}

func Decode(il Iterator, handle func(in.Instruction)) {
	for op := il.Next(); op != 0; op = il.Next() {
		switch {
		case op&in.MoveMask == in.MovePattern:
			// LD D, S. 0b01dd dsss
			handle(in.Move{Source: in.Source(op), Dest: in.Dest(op)})
		case op&in.MoveImmediateMask == in.MoveImmediatePattern:
			// LD D, n. 0b00dd d110
			handle(in.MoveImmediate{Dest: in.Dest(op), Immediate: il.Next()})
		case op^in.LoadIncrementPattern == 0:
			// LDI A, (HL) 0b0010 1010
			handle(in.LoadIncrement{})
		case op^in.StoreIncrementPattern == 0:
			// LDI (HL), A. 0b0010 0010
			handle(in.StoreIncrement{})
		case op^in.LoadDecrementPattern == 0:
			// LDD A, (HL) 0b0011 1010
			handle(in.LoadDecrement{})
		case op^in.StoreDecrementPattern == 0:
			// LDD (HL), A 0b0011 0010
			handle(in.StoreDecrement{})
		case op&in.MoveIndirectMask == in.LoadIndirectPattern:
			// LD, r, (pair). 0b00dd 1010
			handle(in.LoadIndirect{Dest: registers.A, Source: in.Pair(op)})
		case op&in.MoveIndirectMask == in.StoreIndirectPattern:
			// LD (pair), r. 0b00ss 0010
			handle(in.StoreIndirect{Source: registers.A, Dest: in.Pair(op)})
		case op^in.HLtoSPPattern == 0:
			// LD SP, HL. 0b 1111 1001
			handle(in.HLtoSP{})
		case op == in.LoadRelativePattern:
			// LD A, (C). 0b1111 0010
			handle(in.LoadRelative{})
		case op == in.LoadRelativeImmediateNPattern:
			// LD A, n. 0b1111 0000
			handle(in.LoadRelativeImmediateN{Immediate: il.Next()})
		case op == in.LoadRelativeImmediateNNPattern:
			// LD A, nn. 0b1111 1010
			Immediate := utils.MergePair(il.Next(), il.Next())
			handle(in.LoadRelativeImmediateNN{Immediate})
		case op == in.StoreRelativePattern:
			// LD (C), A. 0b1110 0010
			handle(in.StoreRelative{})
		case op == in.StoreRelativeImmediateNPattern:
			// LD n, A. 0b1110 0000
			handle(in.StoreRelativeImmediateN{Immediate: il.Next()})
		case op == in.StoreRelativeImmediateNNPattern:
			// LD nn, A. 0b1110 1010
			Immediate := utils.MergePair(il.Next(), il.Next())
			handle(in.StoreRelativeImmediateNN{Immediate})
		case op&in.LoadRegisterPairImmediateMask == in.LoadRegisterPairImmediatePattern:
			// LD dd, nn. 0b00dd 0001
			Immediate := utils.ReverseMergePair(il.Next(), il.Next())
			handle(in.LoadRegisterPairImmediate{Dest: in.Pair(op), Immediate: Immediate})
		case op&in.PushPopMask == in.PushPattern:
			// PUSH qq. 0b11qq 0101
			handle(in.Push{Source: in.DemuxPairs(op)})
		case op&in.PushPopMask == in.PopPattern:
			// POP qq. 0b11qq 0001
			handle(in.Pop{Dest: in.DemuxPairs(op)})
		case op == in.LoadHLSPPattern:
			// LDHL SP, e. 0b1111 1000
			handle(in.LoadHLSP{Immediate: int8(il.Next())})
		case op == in.StoreSPPattern:
			// LD nn, SP. 0b0000 1000
			Immediate := utils.ReverseMergePair(il.Next(), il.Next())
			handle(in.StoreSP{Immediate})
		case op&in.AddMask == in.AddPattern:
			// ADD A, r. 0b1000 0rrr
			// ADC A, r. 0b1000 1rrr
			WithCarry := (op & in.CarryMask) > 0
			handle(in.Add{Source: in.Source(op), WithCarry: WithCarry})
		case op&in.AddImmediateMask == in.AddImmediatePattern:
			// ADD A n. 0b1100 0110
			WithCarry := (op & in.CarryMask) > 0
			handle(in.AddImmediate{WithCarry: WithCarry, Immediate: il.Next()})
		case op&in.SubtractMask == in.SubtractPattern:
			// SUB A, r. 0b1001 0rrr
			// SBC A, r. 0b1001 1rrr
			WithCarry := (op & in.CarryMask) > 0
			handle(in.Subtract{Source: in.Source(op), WithCarry: WithCarry})
		case op&in.SubtractImmediateMask == in.SubtractImmediatePattern:
			// SUB A n. 0b1101 0110
			// SBC A n. 0b1101 1110
			WithCarry := (op & in.CarryMask) > 0
			handle(in.SubtractImmediate{WithCarry: WithCarry, Immediate: il.Next()})
		case op&in.AndMask == in.AndPattern:
			// AND A r. 0b1010 0rrr
			handle(in.And{Source: in.Source(op)})
		case op == in.AndImmediatePattern:
			// AND A n. 0b1110 0110
			handle(in.AndImmediate{Immediate: il.Next()})
		case op&in.OrMask == in.OrPattern:
			// OR A r. 0b1011 0rrr
			handle(in.Or{Source: in.Source(op)})
		case op == in.OrImmediatePattern:
			// OR A n. 0b1111 0110
			handle(in.OrImmediate{Immediate: il.Next()})
		case op&in.XorMask == in.XorPattern:
			// XOR A r. 0b1010 1rrr
			handle(in.Xor{Source: in.Source(op)})
		case op == in.XorImmediatePattern:
			// OR A n. 0b1110 1110
			handle(in.XorImmediate{Immediate: il.Next()})
		case op&in.CmpMask == in.CmpPattern:
			// CP A r. 0b1011 1rrr
			handle(in.Cmp{Source: in.Source(op)})
		case op == in.CmpImmediatePattern:
			// OR A n. 0b1111 1110
			handle(in.CmpImmediate{Immediate: il.Next()})
		case op&in.IncrementMask == in.IncrementPattern:
			// INC r. 0b00rr r100
			handle(in.Increment{Dest: in.Dest(op)})
		case op&in.DecrementMask == in.DecrementPattern:
			// DEC r. 0b00rr r101
			handle(in.Decrement{Dest: in.Dest(op)})
		case op&in.AddPairMask == in.AddPairPattern:
			// ADD HL, ss. 0b00ss 1001
			handle(in.AddPair{Source: in.Pair(op)})
		case op == in.AddSPPattern:
			// ADD SP, n. 0b1110 1000
			handle(in.AddSP{Immediate: il.Next()})
		case op&in.IncrementPairMask == in.IncrementPairPattern:
			// INC ss. 0b00ss 0011
			handle(in.IncrementPair{Dest: in.Pair(op)})
		case op&in.DecrementPairMask == in.DecrementPairPattern:
			// INC ss. 0b00ss 1011
			handle(in.DecrementPair{Dest: in.Pair(op)})
		case op&in.RotateMask == in.RotateAPattern:
			// RLCA. 0b0000 0111
			// RLA. 0b00001 0111
			// RRCA. 0b0000 1111
			// RRA. 0b0001 1111
			handle(in.RotateA{Direction: in.GetDirection(op), WithCopy: in.GetWithCopyRotation(op)})
		case op == in.Prefix:
			operand := il.Next()
			switch {
			case operand&in.SwapMask == in.SwapPattern:
				// SWAP m. 0b1100 1011, 0011 0rrr
				handle(in.Swap{Source: in.Source(operand)})
			case operand&in.ShiftMask == in.ShiftPattern:
				// SLA m. 0b1100 1011, 0010 0rrr
				// SRA m. 0b1100 1011, 0010 1rrr
				// SRR m. 0b1100 1011, 0011 1rrr
				handle(in.Shift{Direction: in.GetDirection(operand), Source: in.Source(operand), WithCopy: in.GetWithCopyShift(operand)})
			case operand&in.BitMask == in.BitPattern:
				// BIT b r. 0b1100 1011, 01bb brrr
				handle(in.Bit{Source: in.Source(operand), BitNumber: in.BitNumber(operand)})
			case operand&in.SetMask == in.SetPattern:
				// SET b r. 0b1100 1011, 11bb brrr
				handle(in.Set{Source: in.Source(operand), BitNumber: in.BitNumber(operand)})
			case operand&in.ResetMask == in.ResetPattern:
				// RES b r. 0b1100 1011, 10bb brrr
				handle(in.Reset{Source: in.Source(operand), BitNumber: in.BitNumber(operand)})
			default:
				// RLC r. 0b11000 1011, 0001 0rrr
				// RL r. 0b11000 1011, 0001 0rrr
				// RRC r. 0b11000 1011, 0000 1rrr
				// RR r. 0b11000 1011, 0001 1rrr
				handle(in.RotateOperand{Direction: in.GetDirection(operand), WithCopy: in.GetWithCopyRotation(operand), Source: in.Source(operand)})
			}
		case op == in.JumpImmediatePattern:
			// JP nn. 0b1100 0011, L, H
			Immediate := utils.ReverseMergePair(il.Next(), il.Next())
			handle(in.JumpImmediate{Immediate})
		case op&in.JumpConditionalMask == in.JumpImmediateConditionalPattern:
			// JP cc nn. 0b110c c010, L, H
			Immediate := utils.ReverseMergePair(il.Next(), il.Next())
			Condition := in.GetCondition(op)
			handle(in.JumpImmediateConditional{Immediate, Condition})
		case op == in.JumpRelativePattern:
			// JR, n. 0b0001 1000, n
			Immediate := int8(il.Next()) + 2
			handle(in.JumpRelative{Immediate})
		case op&in.JumpConditionalMask == in.JumpRelativeConditionalPattern:
			// JR cc n. 0b001c c000, n
			Immediate := int8(il.Next()) + 2
			Condition := in.GetCondition(op)
			handle(in.JumpRelativeConditional{Immediate, Condition})
		case op == in.JumpMemoryPattern:
			// JP (HL). 0b1110 1001
			handle(in.JumpMemory{})
		case op == in.CallPattern:
			// CALL nn. 0b1100 1101, n, n
			Immediate := utils.ReverseMergePair(il.Next(), il.Next())
			handle(in.Call{Immediate})
		case op&in.CallConditionalMask == in.CallConditionalPattern:
			// CALL cc nn. 0b110c c100
			Immediate := utils.ReverseMergePair(il.Next(), il.Next())
			Condition := in.GetCondition(op)
			handle(in.CallConditional{Immediate, Condition})
		case op == in.ReturnPattern:
			// RET. 0b1100 1001
			handle(in.Return{})
		case op == in.ReturnInterruptPattern:
			// RETI. 0b1101 1001
			handle(in.ReturnInterrupt{})
		case op&in.ReturnConditionalMask == in.ReturnConditionalPattern:
			// RET cc nn. 0b110c c000
			Condition := in.GetCondition(op)
			handle(in.ReturnConditional{Condition})
		case op&in.RSTMask == in.RSTPattern:
			// RST t. 0b11tt t111
			handle(in.RST{Operand: in.GetOperand(op)})
		case op == in.ComplementPattern:
			// CPL. 0b0010 1111
			handle(in.Complement{})
		case op == in.CCFPattern:
			// CCF. 0b0011 1111
			handle(in.CCF{})
		case op == in.SCFPattern:
			// CCF. 0b0011 0111
			handle(in.SCF{})
		case op == in.NopPattern:
			// NOP. 0b0000 0000
			handle(in.Nop{})
		default:
			handle(in.InvalidInstruction{ErrorOpcode: op})
		}
	}
}

// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ebpf

import (
	"fmt"
)

// MemoryInstruction Represents an eBPF load/store operation with an immediate value.
type MemoryInstruction struct {

	// Add all the basic things all instructions have.
	BaseInstruction

	// Size of the operation to make.
	Size uint8

	// Mode of the operation.
	Mode uint8

	// Even if this is an imm operation, it seems that ebpf uses the
	// src register to point at what type of value we are loading from
	// memory. E.G. if srcReg is set to 0x1, the imm will get treated as
	// a map fd. (http://shortn/_cHoySHsuW2)

	// SrcReg is the source register.
	SrcReg *Register

	// Imm value to use.
	Imm int32

	// Offset of memory region to operate.
	Offset int16
}

// GenerateBytecode generates the bytecode associated with this instruction.
func (c *MemoryInstruction) GenerateBytecode() []uint64 {
	bytecode := []uint64{encodeImmediateStOrLdInstruction(c.InstructionClass, c.Size, c.Mode, c.DstReg.RegisterNumber(), c.SrcReg.RegisterNumber(), c.Imm, c.Offset)}

	// It seems that the ld_imm64 instructions need a "pseudo instruction"
	// after them, the documentation is not clear about it but
	// we can find references to insn[1] (which refers to it) in the
	// verifier code: http://shortn/_cHoySHsuW2
	if c.InstructionClass == InsClassLd && c.Mode == StLdModeIMM {
		bytecode = append(bytecode, uint64(0))
	}
	return bytecode
}

// GeneratePoc generates the C macros to repro this program.
func (c *MemoryInstruction) GeneratePoc() []string {
	var macro string
	var insClass string
	if c.InstructionClass == InsClassStx {
		insClass = "BPF_STX"
	} else if c.InstructionClass == InsClassSt {
		insClass = "BPF_ST"
	} else {
		insClass = "BPF_LDX"
	}
	var size string
	switch c.Size {
	case StLdSizeW:
		size = "BPF_W"
	case StLdSizeH:
		size = "BPF_H"
	case StLdSizeB:
		size = "BPF_B"
	case StLdSizeDW:
		size = "BPF_DW"
	default:
		size = "unknown"
	}
	if c.InstructionClass == InsClassLd && c.Mode == StLdModeIMM {
		macro = fmt.Sprintf("BPF_LD_MAP_FD(/*dst=*/%s, map_fd)", c.DstReg.ToString())
	} else if c.InstructionClass == InsClassStx || c.InstructionClass == InsClassLdx {
		macro = fmt.Sprintf("BPF_MEM_OPERATION(%s, %s, /*dst=*/%d, /*src=*/%d, /*offset=*/%d)", insClass, size, c.DstReg.ToString(), c.SrcReg.ToString(), c.Offset)
	} else {
		macro = fmt.Sprintf("BPF_MEM_IMM_OPERATION(%s, %s, /*dst=*/%d, /*src=*/%d, /*offset=*/%d)", insClass, size, c.DstReg.ToString(), c.Imm, c.Offset)
	}

	r := []string{macro}
	return r
}

func newStoreOperation(size uint8, dstReg *Register, src interface{}, offset int16) Instruction {
	var srcReg *Register
	var imm int32
	var insClass uint8
	isInt, srcInt := isIntType(src)
	if isInt {
		imm = int32(srcInt)
		// srcReg will be mostly ignored in this case, we need to specify
		// something so the default nil value doesn't cause trouble.
		srcReg = RegR0
		insClass = InsClassSt
	} else if srcR, ok := src.(*Register); ok {
		srcReg = srcR
		imm = 0
		insClass = InsClassStx
	} else {
		return nil
	}

	return &MemoryInstruction{
		BaseInstruction: BaseInstruction{
			InstructionClass: insClass,
			DstReg:           dstReg,
		},
		Mode: StLdModeMEM,
		Size: size,
		// SrcReg is unused, put it here because otherwise it will be nil
		// and it will cause problems somewhere else.
		SrcReg: srcReg,
		Offset: offset,
		Imm:    imm,
	}
}

// StDW Stores 8 byte data from `src` into `dst`
func StDW(dst *Register, src interface{}, offset int16) Instruction {
	return newStoreOperation(StLdSizeDW, dst, src, offset)
}

// StDW Stores 4 byte data from `src` into `dst`
func StW(dst *Register, src interface{}, offset int16) Instruction {
	return newStoreOperation(StLdSizeW, dst, src, offset)
}

// StH Stores 2 byte (Half word) data from `src` into `dst`
func StH(dst *Register, src interface{}, offset int16) Instruction {
	return newStoreOperation(StLdSizeH, dst, src, offset)
}

// StB Stores 1 byte data from `src` into `dst`
func StB(dst *Register, src interface{}, offset int16) Instruction {
	return newStoreOperation(StLdSizeB, dst, src, offset)
}

func newLoadToRegisterOperation(size uint8, dstReg *Register, src *Register, offset int16) Instruction {
	return &MemoryInstruction{
		BaseInstruction: BaseInstruction{
			InstructionClass: InsClassLdx,
			DstReg:           dstReg,
		},
		Mode: StLdModeMEM,
		Size: size,
		// SrcReg is unused, put it here because otherwise it will be nil
		// and it will cause problems somewhere else.
		SrcReg: src,
		Offset: offset,
	}
}

// LdDW Stores 8 byte data from `src` into `dst`
func LdDW(dst *Register, src *Register, offset int16) Instruction {
	return newLoadToRegisterOperation(StLdSizeDW, dst, src, offset)
}

// LdW Stores 4 byte data from `src` into `dst`
func LdW(dst *Register, src *Register, offset int16) Instruction {
	return newLoadToRegisterOperation(StLdSizeW, dst, src, offset)
}

// LdH Stores 2 byte (Half word) data from `src` into `dst`
func LdH(dst *Register, src *Register, offset int16) Instruction {
	return newLoadToRegisterOperation(StLdSizeH, dst, src, offset)
}

// LdB Stores 1 byte data from `src` into `dst`
func LdB(dst *Register, src *Register, offset int16) Instruction {
	return newLoadToRegisterOperation(StLdSizeB, dst, src, offset)
}

func LdMapByFd(dst *Register, fd int) Instruction {
	return &MemoryInstruction{
		BaseInstruction: BaseInstruction{
			InstructionClass: InsClassLd,
			DstReg:           dst,
		},
		Size: StLdSizeDW,
		Mode: StLdModeIMM,
		// SrcReg is unused, put it here because otherwise it will be nil
		// and it will cause problems somewhere else.
		SrcReg: PseudoMapFD,
		Imm:    int32(fd),
	}
}

func newAtomicInstruction(dst, src *Register, size, class uint8, offset int16, operation int32) Instruction {
	return &MemoryInstruction{
		BaseInstruction: BaseInstruction{
			InstructionClass: class,
			DstReg:           dst,
		},
		Size: size,
		Mode: StLdModeATOMIC,
		// SrcReg is unused, put it here because otherwise it will be nil
		// and it will cause problems somewhere else.
		SrcReg: src,
		Offset: offset,
		Imm:    operation,
	}
}

func MemAdd64(dst, src *Register, offset int16) Instruction {
	return newAtomicInstruction(dst, src, StLdSizeDW, InsClassStx, offset, int32(AluAdd))
}

func MemAdd(dst, src *Register, offset int16) Instruction {
	return newAtomicInstruction(dst, src, StLdSizeW, InsClassStx, offset, int32(AluAdd))
}

func MemOr64(dst, src *Register, offset int16) Instruction {
	return newAtomicInstruction(dst, src, StLdSizeDW, InsClassStx, offset, int32(AluOr))
}

func MemOr(dst, src *Register, offset int16) Instruction {
	return newAtomicInstruction(dst, src, StLdSizeW, InsClassStx, offset, int32(AluOr))
}

func MemAnd64(dst, src *Register, offset int16) Instruction {
	return newAtomicInstruction(dst, src, StLdSizeDW, InsClassStx, offset, int32(AluAnd))
}

func MemAnd(dst, src *Register, offset int16) Instruction {
	return newAtomicInstruction(dst, src, StLdSizeW, InsClassStx, offset, int32(AluAnd))
}

func MemXor64(dst, src *Register, offset int16) Instruction {
	return newAtomicInstruction(dst, src, StLdSizeDW, InsClassStx, offset, int32(AluXor))
}

func MemXor(dst, src *Register, offset int16) Instruction {
	return newAtomicInstruction(dst, src, StLdSizeW, InsClassStx, offset, int32(AluXor))
}

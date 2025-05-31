package run

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unsafe"

	"github.com/paraskun/ess-go/typ"
)

const (
	StackSize = 1024 // stack size in bytes
)

type (
	// Package is a collection of related functions.
	Package struct {
		Img []Image
		Map map[string]int
	}

	// Image is an layout for function data in memory
	// with additional meta information.
	Image struct {
		Name string

		DatSz int
		ArgSz int

		Imm      []byte
		Src      []byte
		CallInfo []int
	}

	// Frame is a representation of function call
	// in memory.
	Frame struct {
		*Image

		Data     []byte
		CallData []*Frame
	}
)

// Virtual machine
type Machine struct {
	Pkg *Package

	rip unsafe.Pointer // instruction pointer
	rsp unsafe.Pointer // stack pointer
	rbp unsafe.Pointer // base pointer
	rcp unsafe.Pointer // constants pointer

	nat []func() // native functions
	vmd []byte   // virtual machine data
	vms []byte   // virtual machine stack

	cf *Frame           // current frame
	sf []*Frame         // saved frames
	sp []unsafe.Pointer // saved instruction pointers
}

// Native functions

// log is a native function intended to
// send arbitrary type argument to given
// channel.
//
// This function should be replaced by
// environment provided one.
func (vm *Machine) log() {
	switch typ.Kind(vm.LoadU08()) {
	case typ.BOOL:
		if vm.LoadU08() == 0 {
			fmt.Printf("false\n")
		} else {
			fmt.Printf("true\n")
		}
	case typ.I64:
		fmt.Printf("%v\n", vm.LoadI64())
	case typ.U64:
		fmt.Printf("%v\n", vm.LoadU64())
	case typ.F64:
		fmt.Printf("%v\n", vm.LoadF64())
	}
}

// loadFunc allocates new Frame for function
// with given offset.
func (vm *Machine) loadFunc(n int) *Frame {
	if n < 0 {
		return nil // native function
	}

	fi := &vm.Pkg.Img[n]
	ff := &Frame{
		Image:    fi,
		Data:     make([]byte, fi.DatSz),
		CallData: make([]*Frame, len(fi.CallInfo)),
	}

	*(*uintptr)(unsafe.Pointer(&ff.Data[0])) = uintptr(unsafe.Pointer(&ff.Data[0]))
	*(*uintptr)(unsafe.Pointer(&ff.Data[8])) = uintptr(unsafe.Pointer(&vm.vmd[0]))

	for i, c := range fi.CallInfo {
		ff.CallData[i] = vm.loadFunc(c)
	}

	return ff
}

func (vm *Machine) Load(pkg *Package, name string) {
	vm.Pkg = pkg

	vm.nat = []func(){vm.log}
	vm.vmd = make([]byte, 1)
	vm.vms = make([]byte, StackSize)
	vm.rsp = unsafe.Pointer(&vm.vms[0])

	idx, ok := pkg.Map[name]

	if !ok {
		panic(fmt.Errorf("function with name %v not found", name))
	}

	vm.cf = vm.loadFunc(idx)
	vm.sf = []*Frame{vm.cf}
	vm.sp = make([]unsafe.Pointer, 0)
}

func (vm *Machine) Exec(arg []byte) {
	vm.cf = vm.sf[0]
	vm.rip = unsafe.Pointer(&vm.cf.Src[0])
	vm.rbp = unsafe.Pointer(&vm.cf.Data[0])

	if len(vm.cf.Imm) > 0 {
		vm.rcp = unsafe.Pointer(&vm.cf.Imm[0])
	}

	vm.vmd[0] = 0
	num := len(arg)
	dst := unsafe.Slice((*byte)(vm.rsp), num)

	copy(dst[:num], arg[:num])

	vm.rsp = unsafe.Add(vm.rsp, num)

	for {
		cmd := Code(vm.nu08())

		switch cmd {
		case JMP:
			vm.rip = unsafe.Add(vm.rip, vm.ni32())
		case JIF:
			off := vm.ni32()

			if vm.LoadU08() == 0 {
				vm.rip = unsafe.Add(vm.rip, off)
			}
		case CALL:
			idx := vm.nu32()
			fun := vm.cf.CallInfo[idx]

			if fun < 0 {
				vm.nat[-fun-1]()
				continue
			}

			vm.cf = vm.cf.CallData[idx]
			vm.sf = append(vm.sf, vm.cf)
			vm.sp = append(vm.sp, vm.rip)

			vm.rip = unsafe.Pointer(&vm.cf.Src[0])
			vm.rbp = unsafe.Pointer(&vm.cf.Data[0])

			if len(vm.cf.Imm) > 0 {
				vm.rcp = unsafe.Pointer(&vm.cf.Imm[0])
			}
		case RET:
			if len(vm.sf) == 1 {
				return
			}

			vm.sf = vm.sf[:len(vm.sf)-1]
			vm.cf = vm.sf[len(vm.sf)-1]

			vm.rip = vm.sp[len(vm.sp)-1]
			vm.rbp = unsafe.Pointer(&vm.cf.Data[0])

			vm.sp = vm.sp[:len(vm.sp)-1]

			if len(vm.cf.Imm) > 0 {
				vm.rcp = unsafe.Pointer(&vm.cf.Imm[0])
			}
		case PUSHB:
			vm.su08(vm.nu08())
		case PUSHW:
			vm.su32(vm.nu32())
		case PUSHD:
			vm.su64(vm.nu64())
		case LEAII:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(uintptr(*(*uint64)(ptr))), vm.nu32())
			vm.su64(uint64(*(*uintptr)(off)))
		case LBII:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())

			vm.su08(*(*uint8)(off))
		case LWII:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())

			vm.su32(*(*uint32)(off))
		case LDII:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(uintptr(*(*uint64)(ptr))), vm.nu32())

			vm.su64(*(*uint64)(off))
		case LAII:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())
			num := int(vm.nu32())

			src := unsafe.Slice((*byte)(off), num)
			dst := unsafe.Slice((*byte)(vm.rsp), num)

			copy(dst[:num], src[:num])

			vm.rsp = unsafe.Add(vm.rsp, num)
		case SBII:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())

			*(*uint8)(off) = vm.LoadU08()
		case SWII:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())

			*(*uint32)(off) = vm.LoadU32()
		case SDII:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())

			*(*uint64)(off) = vm.LoadU64()
		case SAII:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())
			num := int(vm.nu32())

			vm.rsp = unsafe.Add(vm.rsp, -num)

			src := unsafe.Slice((*byte)(vm.rsp), num)
			dst := unsafe.Slice((*byte)(off), num)

			copy(dst[:num], src[:num])
		case LBIS:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())

			vm.su08(*(*uint8)(off))
		case LWIS:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())

			vm.su32(*(*uint32)(off))
		case LDIS:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())

			vm.su64(*(*uint64)(off))
		case LAIS:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())
			num := int(vm.nu32())

			src := unsafe.Slice((*byte)(off), num)
			dst := unsafe.Slice((*byte)(vm.rsp), num)

			copy(dst[:num], src[:num])

			vm.rsp = unsafe.Add(vm.rsp, num)
		case SBIS:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())

			*(*uint8)(off) = vm.LoadU08()
		case SWIS:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())

			*(*uint32)(off) = vm.LoadU32()
		case SDIS:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())

			*(*uint64)(off) = vm.LoadU64()
		case SAIS:
			ptr := unsafe.Add(vm.rbp, vm.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())
			num := int(vm.nu32())

			vm.rsp = unsafe.Add(vm.rsp, -num)

			src := unsafe.Slice((*byte)(vm.rsp), num)
			dst := unsafe.Slice((*byte)(off), num)

			copy(dst[:num], src[:num])
		case LBSI:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())

			vm.su08(*(*uint8)(off))
		case LWSI:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())

			vm.su32(*(*uint32)(off))
		case LDSI:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())

			vm.su64(*(*uint64)(off))
		case LASI:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())
			num := int(vm.nu32())

			src := unsafe.Slice((*byte)(off), num)
			dst := unsafe.Slice((*byte)(vm.rsp), num)

			copy(dst[:num], src[:num])

			vm.rsp = unsafe.Add(vm.rsp, num)
		case SBSI:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())

			*(*uint8)(off) = vm.LoadU08()
		case SWSI:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())

			*(*uint32)(off) = vm.LoadU32()
		case SDSI:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())

			*(*uint64)(off) = vm.LoadU64()
		case SASI:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.nu32())
			num := int(vm.nu32())

			vm.rsp = unsafe.Add(vm.rsp, -num)

			src := unsafe.Slice((*byte)(vm.rsp), num)
			dst := unsafe.Slice((*byte)(off), num)

			copy(dst[:num], src[:num])
		case LBSS:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())

			vm.su08(*(*uint8)(off))
		case LWSS:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())

			vm.su32(*(*uint32)(off))
		case LDSS:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())

			vm.su64(*(*uint64)(off))
		case LASS:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())
			num := int(vm.nu32())

			src := unsafe.Slice((*byte)(off), num)
			dst := unsafe.Slice((*byte)(vm.rsp), num)

			copy(dst[:num], src[:num])

			vm.rsp = unsafe.Add(vm.rsp, num)
		case SBSS:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())

			*(*uint8)(off) = vm.LoadU08()
		case SWSS:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())

			*(*uint32)(off) = vm.LoadU32()
		case SDSS:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())

			*(*uint64)(off) = vm.LoadU64()
		case SASS:
			ptr := unsafe.Add(vm.rbp, vm.LoadU32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), vm.LoadU32())
			num := int(vm.nu32())

			vm.rsp = unsafe.Add(vm.rsp, -num)

			src := unsafe.Slice((*byte)(vm.rsp), num)
			dst := unsafe.Slice((*byte)(off), num)

			copy(dst[:num], src[:num])
		case LBI:
			off := unsafe.Add(vm.rcp, vm.nu32())
			vm.su08(*(*uint8)(off))
		case LWI:
			off := unsafe.Add(vm.rcp, vm.nu32())
			vm.su32(*(*uint32)(off))
		case LDI:
			off := unsafe.Add(vm.rcp, vm.nu32())
			vm.su64(*(*uint64)(off))
		case LAI:
			off := unsafe.Add(vm.rcp, vm.nu32())
			num := int(vm.nu32())

			src := unsafe.Slice((*byte)(off), num)
			dst := unsafe.Slice((*byte)(vm.rsp), num)

			copy(dst[:num], src[:num])

			vm.rsp = unsafe.Add(vm.rsp, num)
		case ADDI:
			vm.si64(vm.LoadI64() + vm.LoadI64())
		case SUBI:
			vm.si64(vm.LoadI64() - vm.LoadI64())
		case MULI:
			vm.si64(vm.LoadI64() * vm.LoadI64())
		case DIVI:
			vm.si64(vm.LoadI64() / vm.LoadI64())
		case POWI:
			vm.si64(int64(math.Pow(float64(vm.LoadI64()), float64(vm.LoadI64()))))
		case SHLI:
			vm.si64(vm.LoadI64() << vm.LoadI64())
		case SHRI:
			vm.si64(vm.LoadI64() >> vm.LoadI64())
		case MODI:
			vm.si64(vm.LoadI64() % vm.LoadI64())
		case XORI:
			vm.si64(vm.LoadI64() ^ vm.LoadI64())
		case ANDI:
			vm.si64(vm.LoadI64() & vm.LoadI64())
		case ORI:
			vm.si64(vm.LoadI64() | vm.LoadI64())
		case BNEGI:
			vm.si64(^vm.LoadI64())
		case UNEGI:
			vm.si64(-vm.LoadI64())
		case LTI:
			if vm.LoadI64() < vm.LoadI64() {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case LEI:
			if vm.LoadI64() <= vm.LoadI64() {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case EQI:
			if vm.LoadI64() == vm.LoadI64() {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case NEI:
			if vm.LoadI64() != vm.LoadI64() {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case ADDU:
			vm.su64(vm.LoadU64() + vm.LoadU64())
		case SUBU:
			vm.su64(vm.LoadU64() - vm.LoadU64())
		case MULU:
			vm.su64(vm.LoadU64() * vm.LoadU64())
		case DIVU:
			vm.su64(vm.LoadU64() / vm.LoadU64())
		case POWU:
			vm.su64(uint64(math.Pow(float64(vm.LoadU64()), float64(vm.LoadU64()))))
		case SHLU:
			vm.su64(vm.LoadU64() << vm.LoadU64())
		case SHRU:
			vm.su64(vm.LoadU64() >> vm.LoadU64())
		case MODU:
			vm.su64(vm.LoadU64() % vm.LoadU64())
		case XORU:
			vm.su64(vm.LoadU64() ^ vm.LoadU64())
		case ANDU:
			vm.su64(vm.LoadU64() & vm.LoadU64())
		case ORU:
			vm.su64(vm.LoadU64() | vm.LoadU64())
		case BNEGU:
			vm.su64(^vm.LoadU64())
		case UNEGU:
			vm.su64(-vm.LoadU64())
		case LTU:
			if vm.LoadU64() < vm.LoadU64() {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case LEU:
			if vm.LoadU64() <= vm.LoadU64() {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case EQU:
			if vm.LoadU64() == vm.LoadU64() {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case NEU:
			if vm.LoadU64() != vm.LoadU64() {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case ADDF:
			vm.sf64(vm.LoadF64() + vm.LoadF64())
		case SUBF:
			vm.sf64(vm.LoadF64() - vm.LoadF64())
		case MULF:
			vm.sf64(vm.LoadF64() * vm.LoadF64())
		case DIVF:
			vm.sf64(vm.LoadF64() / vm.LoadF64())
		case POWF:
			vm.sf64(math.Pow(vm.LoadF64(), vm.LoadF64()))
		case UNEGF:
			vm.sf64(-vm.LoadF64())
		case LTF:
			if vm.LoadF64() < vm.LoadF64() {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case LEF:
			if vm.LoadF64() <= vm.LoadF64() {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case EQF:
			if vm.LoadF64() == vm.LoadF64() {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case NEF:
			if vm.LoadF64() != vm.LoadF64() {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case ANDL:
			if vm.LoadU08() == 0 || vm.LoadU08() == 0 {
				vm.su08(0)
			} else {
				vm.su08(1)
			}
		case ORL:
			if vm.LoadU08() == 0 && vm.LoadU08() == 0 {
				vm.su08(0)
			} else {
				vm.su08(1)
			}
		case NEGL:
			if vm.LoadU08() == 0 {
				vm.su08(1)
			} else {
				vm.su08(0)
			}
		case I2U:
			vm.su64(uint64(vm.LoadI64()))
		case I2F:
			vm.sf64(float64(vm.LoadI64()))
		case U2I:
			vm.si64(int64(vm.LoadU64()))
		case U2F:
			vm.sf64(float64(vm.LoadU64()))
		case F2I:
			vm.si64(int64(vm.LoadF64()))
		case F2U:
			vm.su64(uint64(vm.LoadF64()))
		}
	}
}

func (m *Machine) nu08() uint8 {
	r := *(*uint8)(m.rip)
	m.rip = unsafe.Add(m.rip, 1)

	return r
}

func (m *Machine) nu32() uint32 {
	r := *(*uint32)(m.rip)
	m.rip = unsafe.Add(m.rip, 4)

	return r
}

func (m *Machine) nu64() uint64 {
	r := *(*uint64)(m.rip)
	m.rip = unsafe.Add(m.rip, 8)

	return r
}

func (m *Machine) ni32() int32 {
	r := *(*int32)(m.rip)
	m.rip = unsafe.Add(m.rip, 4)

	return r
}

func (m *Machine) LoadI08() int8 {
	m.rsp = unsafe.Add(m.rsp, -1)
	return *(*int8)(m.rsp)
}

func (m *Machine) LoadI64() int64 {
	m.rsp = unsafe.Add(m.rsp, -8)
	return *(*int64)(m.rsp)
}

func (m *Machine) LoadU08() uint8 {
	m.rsp = unsafe.Add(m.rsp, -1)
	return *(*uint8)(m.rsp)
}

func (m *Machine) LoadU32() uint32 {
	m.rsp = unsafe.Add(m.rsp, -4)
	return *(*uint32)(m.rsp)
}

func (m *Machine) LoadU64() uint64 {
	m.rsp = unsafe.Add(m.rsp, -8)
	return *(*uint64)(m.rsp)
}

func (m *Machine) LoadF64() float64 {
	m.rsp = unsafe.Add(m.rsp, -8)
	return *(*float64)(m.rsp)
}

func (m *Machine) si08(v int8) {
	*(*int8)(m.rsp) = v
	m.rsp = unsafe.Add(m.rsp, 1)
}

func (m *Machine) si64(v int64) {
	*(*int64)(m.rsp) = v
	m.rsp = unsafe.Add(m.rsp, 8)
}

func (m *Machine) su08(v uint8) {
	*(*uint8)(m.rsp) = v
	m.rsp = unsafe.Add(m.rsp, 1)
}

func (m *Machine) su32(v uint32) {
	*(*uint32)(m.rsp) = v
	m.rsp = unsafe.Add(m.rsp, 4)
}

func (m *Machine) su64(v uint64) {
	*(*uint64)(m.rsp) = v
	m.rsp = unsafe.Add(m.rsp, 8)
}

func (m *Machine) sf64(v float64) {
	*(*float64)(m.rsp) = v
	m.rsp = unsafe.Add(m.rsp, 8)
}

func (p *Package) Debug(w io.Writer) {
	for _, img := range p.Img {
		img.Debug(w)
	}
}

func (img *Image) Debug(w io.Writer) {
	fmt.Fprintf(w, ".func\n")
	fmt.Fprintf(w, "\tdat: %d\n", img.DatSz)
	fmt.Fprintf(w, "\targ: %d\n", img.ArgSz)
	fmt.Fprintf(w, ".text\n")

	src := img.Src
	idx := 0

	for off := 0; off < len(src); {
		fmt.Fprintf(w, "\t%d.\t\t%-4d ", idx, off)
		idx += 1

		switch Code(src[off]) {
		case PUSHB:
			fmt.Fprintf(w, "%s %d\n", Code(src[off]).String(), src[off+1])
			off += 2
		case PUSHD:
			fmt.Fprintf(w, "%s %d\n",
				Code(src[off]).String(),
				binary.LittleEndian.Uint64(src[off+1:]),
			)
			off += 9
		case LAIS, SAIS, LASI, SASI,
			LBII, LWII, LDII, SBII,
			SWII, SDII, LEAII, LAI:
			fmt.Fprintf(w, "%s %d %d\n", Code(src[off]).String(),
				binary.LittleEndian.Uint32(src[off+1:]),
				binary.LittleEndian.Uint32(src[off+5:]),
			)
			off += 9
		case LAII, SAII:
			fmt.Fprintf(w, "%s %d %d %d\n", Code(src[off]).String(),
				binary.LittleEndian.Uint32(src[off+1:]),
				binary.LittleEndian.Uint32(src[off+5:]),
				binary.LittleEndian.Uint32(src[off+9:]),
			)
			off += 13
		case JMP, JIF:
			jmp := int32(binary.LittleEndian.Uint32(src[off+1:]))
			fmt.Fprintf(w, "%s %d (%d)\n",
				Code(src[off]).String(),
				jmp,
				int32(off)+5+jmp,
			)
			off += 5
		case CALL:
			fmt.Fprintf(w, "%s %d\n",
				Code(src[off]).String(),
				int32(binary.LittleEndian.Uint32(src[off+1:])),
			)
			off += 5
		case LASS, SASS, PUSHW,
			LBIS, LWIS, LDIS, SBIS,
			SWIS, SDIS, LBSI, LWSI,
			LDSI, SBSI, SWSI, SDSI,
			LBI, LWI, LDI:
			fmt.Fprintf(w, "%s %d\n",
				Code(src[off]).String(),
				binary.LittleEndian.Uint32(src[off+1:]),
			)
			off += 5
		case RET, LBSS, LWSS, LDSS,
			SBSS, SWSS, SDSS, ADDI,
			SUBI, MULI, DIVI, POWI,
			SHLI, SHRI, MODI, XORI,
			ANDI, ORI, BNEGI, UNEGI,
			LTI, LEI, EQI, NEI,
			ADDU, SUBU, MULU, DIVU,
			POWU, SHLU, SHRU, MODU,
			XORU, ANDU, ORU, BNEGU,
			UNEGU, LTU, LEU, EQU,
			NEU, ADDF, SUBF, MULF,
			DIVF, POWF, UNEGF, LTF,
			LEF, EQF, NEF, ANDL,
			ORL, NEGL, I2U, I2F,
			U2I, U2F, F2I, F2U:
			fmt.Fprintf(w, "%s\n", Code(src[off]).String())
			off += 1
		default:
			panic(fmt.Errorf("unkown opcode: %d", src[off]))
		}
	}
}

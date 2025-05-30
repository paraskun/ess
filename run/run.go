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
	}

	// Image is an layout for function data in memory
	// with additional meta information.
	Image struct {
		Name string

		DatSz int
		ArgSz int

		Imm []byte
		Src []byte

		Call []int
	}

	// Frame is a representation of function call
	// in memory.
	Frame struct {
		*Image

		Data []byte
		Call []*Frame
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
	kind := vm.lu08()

	switch typ.Kind(kind & 0b11100000) {
	case typ.COMP:
	case typ.BOOL:
		if vm.lu08() == 0 {
			fmt.Println("false")
		} else {
			fmt.Println("true")
		}
	case typ.I64:
		fmt.Printf("%v\n", vm.li64())
	case typ.U64:
		fmt.Printf("%v\n", vm.lu64())
	case typ.F64:
		fmt.Printf("%v\n", vm.lf64())
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
		Image: fi,
		Data:  make([]byte, fi.DatSz),
		Call:  make([]*Frame, len(fi.Call)),
	}

	*(*uintptr)(unsafe.Pointer(&ff.Data[0])) = uintptr(unsafe.Pointer(&ff.Data[0]))
	*(*uintptr)(unsafe.Pointer(&ff.Data[8])) = uintptr(unsafe.Pointer(&vm.vmd[0]))

	for i, c := range fi.Call {
		ff.Call[i] = vm.loadFunc(c)
	}

	return ff
}

func (vm *Machine) Load(pkg *Package, name string) {
	vm.Pkg = pkg

	vm.nat = []func(){vm.log}
	vm.vmd = make([]byte, 1)
	vm.vms = make([]byte, StackSize)

	vm.sf = make([]*Frame, 1)
	vm.sp = make([]unsafe.Pointer, 0)

	for i, img := range pkg.Img {
		if img.Name == name {

		}
	}
	m.frame = m.rf[0]

	m.rsp = unsafe.Pointer(&m.stack[0])
	m.rbp = unsafe.Pointer(&m.frame.Data[0])

	if len(m.frame.Func.Imm) > 0 {
		m.rcp = unsafe.Pointer(&m.frame.Func.Imm[0])
	}
}

func (m *Machine) Exec(arg []byte) {
	m.frame = m.rf[0]
	m.rip = unsafe.Pointer(&m.rf[0].Func.Src[0])

	m.data[0] = 0

	if m.ini {
		m.data[0] = 1
		m.ini = false
	}

	num := len(arg)
	dst := unsafe.Slice((*byte)(m.rsp), num)

	copy(dst[:num], arg[:num])

	m.rsp = unsafe.Add(m.rsp, num)

	for {
		cmd := Code(m.nu08())

		switch cmd {
		case JMP:
			m.rip = unsafe.Add(m.rip, m.ni32())
		case JIF:
			off := m.ni32()

			if m.lu08() == 0 {
				m.rip = unsafe.Add(m.rip, off)
			}
		case CALL:
			idx := m.nu32()
			fun := m.frame.Func.Call[idx]

			if fun < 0 {
				m.nat[-fun-1]()
				continue
			}

			m.rf = append(m.rf, m.frame.Call[idx])
			m.frame = m.rf[len(m.rf)-1]
			m.rp = append(m.rp, m.rip)
			m.rip = unsafe.Pointer(&m.frame.Func.Src[0])
			m.rbp = unsafe.Pointer(&m.frame.Data[0])
			m.rcp = unsafe.Pointer(&m.frame.Func.Imm[0])
		case RET:
			if len(m.rf) == 1 {
				return
			}

			m.rf = m.rf[:len(m.rf)-1]

			m.frame = m.rf[len(m.rf)-1]
			m.rip = m.rp[len(m.rp)-1]
			m.rp = m.rp[:len(m.rp)-1]
			m.rbp = unsafe.Pointer(&m.frame.Data[0])

			if len(m.frame.Func.Imm) > 0 {
				m.rcp = unsafe.Pointer(&m.frame.Func.Imm[0])
			}
		case PUSHB:
			m.su08(m.nu08())
		case PUSHW:
			m.su32(m.nu32())
		case PUSHD:
			m.su64(m.nu64())
		case LEAII:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())
			m.su64(uint64(*(*uintptr)(off)))
		case LBII:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su08(*(*uint8)(off))
		case LWII:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su32(*(*uint32)(off))
		case LDII:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su64(*(*uint64)(off))
		case LAII:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())
			num := int(m.nu32())

			src := unsafe.Slice((*byte)(off), num)
			dst := unsafe.Slice((*byte)(m.rsp), num)

			copy(dst[:num], src[:num])

			m.rsp = unsafe.Add(m.rsp, num)
		case SBII:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint8)(off) = m.lu08()
		case SWII:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint32)(off) = m.lu32()
		case SDII:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint64)(off) = m.lu64()
		case SAII:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())
			num := int(m.nu32())

			m.rsp = unsafe.Add(m.rsp, -num)

			src := unsafe.Slice((*byte)(m.rsp), num)
			dst := unsafe.Slice((*byte)(off), num)

			copy(dst[:num], src[:num])
		case LBIS:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			m.su08(*(*uint8)(off))
		case LWIS:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			m.su32(*(*uint32)(off))
		case LDIS:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			m.su64(*(*uint64)(off))
		case LAIS:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())
			num := int(m.nu32())

			src := unsafe.Slice((*byte)(off), num)
			dst := unsafe.Slice((*byte)(m.rsp), num)

			copy(dst[:num], src[:num])

			m.rsp = unsafe.Add(m.rsp, num)
		case SBIS:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			*(*uint8)(off) = m.lu08()
		case SWIS:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			*(*uint32)(off) = m.lu32()
		case SDIS:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			*(*uint64)(off) = m.lu64()
		case SAIS:
			ptr := unsafe.Add(m.rbp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())
			num := int(m.nu32())

			m.rsp = unsafe.Add(m.rsp, -num)

			src := unsafe.Slice((*byte)(m.rsp), num)
			dst := unsafe.Slice((*byte)(off), num)

			copy(dst[:num], src[:num])
		case LBSI:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su08(*(*uint8)(off))
		case LWSI:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su32(*(*uint32)(off))
		case LDSI:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su64(*(*uint64)(off))
		case LASI:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())
			num := int(m.nu32())

			src := unsafe.Slice((*byte)(off), num)
			dst := unsafe.Slice((*byte)(m.rsp), num)

			copy(dst[:num], src[:num])

			m.rsp = unsafe.Add(m.rsp, num)
		case SBSI:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint8)(off) = m.lu08()
		case SWSI:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint32)(off) = m.lu32()
		case SDSI:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint64)(off) = m.lu64()
		case SASI:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())
			num := int(m.nu32())

			m.rsp = unsafe.Add(m.rsp, -num)

			src := unsafe.Slice((*byte)(m.rsp), num)
			dst := unsafe.Slice((*byte)(off), num)

			copy(dst[:num], src[:num])
		case LBSS:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			m.su08(*(*uint8)(off))
		case LWSS:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			m.su32(*(*uint32)(off))
		case LDSS:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			m.su64(*(*uint64)(off))
		case LASS:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())
			num := int(m.nu32())

			src := unsafe.Slice((*byte)(off), num)
			dst := unsafe.Slice((*byte)(m.rsp), num)

			copy(dst[:num], src[:num])

			m.rsp = unsafe.Add(m.rsp, num)
		case SBSS:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			*(*uint8)(off) = m.lu08()
		case SWSS:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			*(*uint32)(off) = m.lu32()
		case SDSS:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			*(*uint64)(off) = m.lu64()
		case SASS:
			ptr := unsafe.Add(m.rbp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())
			num := int(m.nu32())

			m.rsp = unsafe.Add(m.rsp, -num)

			src := unsafe.Slice((*byte)(m.rsp), num)
			dst := unsafe.Slice((*byte)(off), num)

			copy(dst[:num], src[:num])
		case LBI:
			off := unsafe.Add(m.rcp, m.nu32())
			m.su08(*(*uint8)(off))
		case LWI:
			off := unsafe.Add(m.rcp, m.nu32())
			m.su32(*(*uint32)(off))
		case LDI:
			off := unsafe.Add(m.rcp, m.nu32())
			m.su64(*(*uint64)(off))
		case LAI:
			off := unsafe.Add(m.rcp, m.nu32())
			num := int(m.nu32())

			src := unsafe.Slice((*byte)(off), num)
			dst := unsafe.Slice((*byte)(m.rsp), num)

			copy(dst[:num], src[:num])

			m.rsp = unsafe.Add(m.rsp, num)
		case ADDI:
			m.si64(m.li64() + m.li64())
		case SUBI:
			m.si64(m.li64() - m.li64())
		case MULI:
			m.si64(m.li64() * m.li64())
		case DIVI:
			m.si64(m.li64() / m.li64())
		case POWI:
			m.si64(int64(math.Pow(float64(m.li64()), float64(m.li64()))))
		case SHLI:
			m.si64(m.li64() << m.li64())
		case SHRI:
			m.si64(m.li64() >> m.li64())
		case MODI:
			m.si64(m.li64() % m.li64())
		case XORI:
			m.si64(m.li64() ^ m.li64())
		case ANDI:
			m.si64(m.li64() & m.li64())
		case ORI:
			m.si64(m.li64() | m.li64())
		case BNEGI:
			m.si64(^m.li64())
		case UNEGI:
			m.si64(-m.li64())
		case LTI:
			if m.li64() < m.li64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case LEI:
			if m.li64() <= m.li64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case EQI:
			if m.li64() == m.li64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case NEI:
			if m.li64() != m.li64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case ADDU:
			m.su64(m.lu64() + m.lu64())
		case SUBU:
			m.su64(m.lu64() - m.lu64())
		case MULU:
			m.su64(m.lu64() * m.lu64())
		case DIVU:
			m.su64(m.lu64() / m.lu64())
		case POWU:
			m.su64(uint64(math.Pow(float64(m.lu64()), float64(m.lu64()))))
		case SHLU:
			m.su64(m.lu64() << m.lu64())
		case SHRU:
			m.su64(m.lu64() >> m.lu64())
		case MODU:
			m.su64(m.lu64() % m.lu64())
		case XORU:
			m.su64(m.lu64() ^ m.lu64())
		case ANDU:
			m.su64(m.lu64() & m.lu64())
		case ORU:
			m.su64(m.lu64() | m.lu64())
		case BNEGU:
			m.su64(^m.lu64())
		case UNEGU:
			m.su64(-m.lu64())
		case LTU:
			if m.lu64() < m.lu64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case LEU:
			if m.lu64() <= m.lu64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case EQU:
			if m.lu64() == m.lu64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case NEU:
			if m.lu64() != m.lu64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case ADDF:
			m.sf64(m.lf64() + m.lf64())
		case SUBF:
			m.sf64(m.lf64() - m.lf64())
		case MULF:
			m.sf64(m.lf64() * m.lf64())
		case DIVF:
			m.sf64(m.lf64() / m.lf64())
		case POWF:
			m.sf64(math.Pow(m.lf64(), m.lf64()))
		case UNEGF:
			m.sf64(-m.lf64())
		case LTF:
			if m.lf64() < m.lf64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case LEF:
			if m.lf64() <= m.lf64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case EQF:
			if m.lf64() == m.lf64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case NEF:
			if m.lf64() != m.lf64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case ANDL:
			if m.lu08() == 0 || m.lu08() == 0 {
				m.su08(0)
			} else {
				m.su08(1)
			}
		case ORL:
			if m.lu08() == 0 && m.lu08() == 0 {
				m.su08(0)
			} else {
				m.su08(1)
			}
		case NEGL:
			if m.lu08() == 0 {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case I2U:
			m.su64(uint64(m.li64()))
		case I2F:
			m.sf64(float64(m.li64()))
		case U2I:
			m.si64(int64(m.lu64()))
		case U2F:
			m.sf64(float64(m.lu64()))
		case F2I:
			m.si64(int64(m.lf64()))
		case F2U:
			m.su64(uint64(m.lf64()))
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

func (m *Machine) li08() int8 {
	m.rsp = unsafe.Add(m.rsp, -1)
	return *(*int8)(m.rsp)
}

func (m *Machine) li64() int64 {
	m.rsp = unsafe.Add(m.rsp, -8)
	return *(*int64)(m.rsp)
}

func (m *Machine) lu08() uint8 {
	m.rsp = unsafe.Add(m.rsp, -1)
	return *(*uint8)(m.rsp)
}

func (m *Machine) lu32() uint32 {
	m.rsp = unsafe.Add(m.rsp, -4)
	return *(*uint32)(m.rsp)
}

func (m *Machine) lu64() uint64 {
	m.rsp = unsafe.Add(m.rsp, -8)
	return *(*uint64)(m.rsp)
}

func (m *Machine) lf64() float64 {
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

func (p *Pragma) Debug(w io.Writer) {
	for _, img := range p.Img {
		img.Debug(w)
	}
}

func (img *FuncImage) Debug(w io.Writer) {
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

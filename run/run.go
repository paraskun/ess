package run

import (
	"math"
	"unsafe"
)

type (
	Pragma struct {
		Img []FuncImage
	}

	FuncImage struct {
		DataSize int
		ArgsSize int

		Imm []byte
		Src []byte

		Call []int
	}

	FuncFrame struct {
		Func *FuncImage
		Data []byte
		Call []*FuncFrame
	}
)

type Machine struct {
	*Pragma

	stack []byte
	frame []*FuncFrame

	ip unsafe.Pointer
	ff *FuncFrame

	rp []unsafe.Pointer
	bp unsafe.Pointer
	sp unsafe.Pointer
}

func (m *Machine) loadFunc(n int) *FuncFrame {
	fi := &m.Pragma.Img[n]
	ff := &FuncFrame{
		Func: fi,
		Data: make([]byte, fi.DataSize),
		Call: make([]*FuncFrame, len(fi.Call)),
	}

	for i, c := range fi.Call {
		ff.Call[i] = m.loadFunc(c)
	}

	return ff
}

func (m *Machine) Load(p *Pragma) {
	m.Pragma = p

	m.frame = []*FuncFrame{m.loadFunc(0)}
	m.stack = make([]byte, 100)

	m.ff = m.frame[0]
	m.ip = unsafe.Pointer(&m.frame[0].Func.Src[0])
	m.bp = unsafe.Pointer(&m.frame[0].Data[0])
	m.sp = unsafe.Pointer(&m.stack[0])
}

func (m *Machine) call(idx uint32) {
	m.frame = append(m.frame, m.ff.Call[idx])
	m.ff = m.frame[len(m.frame)-1]
	m.rp = append(m.rp, m.ip)
	m.ip = unsafe.Pointer(&m.ff.Func.Src[0])
	m.bp = unsafe.Pointer(&m.ff.Data[0])
}

func (m *Machine) Exec() {
	for {
		cmd := Code(m.nu08())

		switch cmd {
		case JMP:
			m.ip = unsafe.Pointer(&m.ff.Func.Src[m.nu32()])
		case JIF:
			if m.lu08() == 0 {
				m.ip = unsafe.Pointer(&m.ff.Func.Src[m.nu32()])
			}
		case CALLI:
			m.call(m.nu32())
		case CALLS:
			m.call(m.lu32())
		case RET:
			m.frame = m.frame[:len(m.frame)-1]
			m.ff = m.frame[len(m.frame)-1]
			m.ip = m.rp[len(m.rp)-1]
			m.bp = unsafe.Pointer(&m.ff.Func.Src[0])
			m.rp = m.rp[:len(m.rp)-1]
		case PUSHB:
			m.su08(m.nu08())
		case PUSHW:
			m.su32(m.nu32())
		case PUSHD:
			m.su64(m.nu64())
		case ADDRI:
			ptr := unsafe.Add(m.bp, m.nu32())
			m.su64(uint64(*(*uintptr)(ptr)))
		case ADDRS:
			ptr := unsafe.Add(m.bp, m.lu32())
			m.su64(uint64(*(*uintptr)(ptr)))
		case LBII:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su08(*(*uint8)(off))
		case LWII:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su32(*(*uint32)(off))
		case LDII:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su64(*(*uint64)(off))
		case LAII:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())
			num := int(m.nu32())

			src := *(*[]byte)(off)
			dst := *(*[]byte)(m.sp)

			copy(dst[:num], src[:num])
		case SBII:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint8)(off) = m.lu08()
		case SWII:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint32)(off) = m.lu32()
		case SDII:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint64)(off) = m.lu64()
		case SAII:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())
			num := int(m.nu32())

			src := *(*[]byte)(unsafe.Add(m.sp, -num))
			dst := *(*[]byte)(off)

			copy(dst[:num], src[:num])
		case LBIS:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			m.su08(*(*uint8)(off))
		case LWIS:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			m.su32(*(*uint32)(off))
		case LDIS:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			m.su64(*(*uint64)(off))
		case LAIS:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())
			num := int(m.nu32())

			src := *(*[]byte)(off)
			dst := *(*[]byte)(m.sp)

			copy(dst[:num], src[:num])
		case SBIS:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			*(*uint8)(off) = m.lu08()
		case SWIS:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			*(*uint32)(off) = m.lu32()
		case SDIS:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			*(*uint64)(off) = m.lu64()
		case SAIS:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())
			num := int(m.nu32())

			src := *(*[]byte)(unsafe.Add(m.sp, -num))
			dst := *(*[]byte)(off)

			copy(dst[:num], src[:num])
		case LBSI:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su08(*(*uint8)(off))
		case LWSI:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su32(*(*uint32)(off))
		case LDSI:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su64(*(*uint64)(off))
		case LASI:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())
			num := int(m.nu32())

			src := *(*[]byte)(off)
			dst := *(*[]byte)(m.sp)

			copy(dst[:num], src[:num])
		case SBSI:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint8)(off) = m.lu08()
		case SWSI:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint32)(off) = m.lu32()
		case SDSI:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint64)(off) = m.lu64()
		case SASI:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())
			num := int(m.nu32())

			src := *(*[]byte)(unsafe.Add(m.sp, -num))
			dst := *(*[]byte)(off)

			copy(dst[:num], src[:num])
		case LBSS:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			m.su08(*(*uint8)(off))
		case LWSS:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			m.su32(*(*uint32)(off))
		case LDSS:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			m.su64(*(*uint64)(off))
		case LASS:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())
			num := int(m.nu32())

			src := *(*[]byte)(off)
			dst := *(*[]byte)(m.sp)

			copy(dst[:num], src[:num])
		case SBSS:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			*(*uint8)(off) = m.lu08()
		case SWSS:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			*(*uint32)(off) = m.lu32()
		case SDSS:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())

			*(*uint64)(off) = m.lu64()
		case SASS:
			ptr := unsafe.Add(m.bp, m.lu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.lu32())
			num := int(m.nu32())

			src := *(*[]byte)(unsafe.Add(m.sp, -num))
			dst := *(*[]byte)(off)

			copy(dst[:num], src[:num])
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
		}
	}
}

func (m *Machine) nu08() uint8 {
	r := *(*uint8)(m.ip)
	m.ip = unsafe.Add(m.ip, 1)

	return r
}

func (m *Machine) nu32() uint32 {
	r := *(*uint32)(m.ip)
	m.ip = unsafe.Add(m.ip, 4)

	return r
}

func (m *Machine) nu64() uint64 {
	r := *(*uint64)(m.ip)
	m.ip = unsafe.Add(m.ip, 8)

	return r
}

func (m *Machine) li08() int8 {
	m.sp = unsafe.Add(m.sp, -1)
	return *(*int8)(m.sp)
}

func (m *Machine) li64() int64 {
	m.sp = unsafe.Add(m.sp, -8)
	return *(*int64)(m.sp)
}

func (m *Machine) lu08() uint8 {
	m.sp = unsafe.Add(m.sp, -1)
	return *(*uint8)(m.sp)
}

func (m *Machine) lu32() uint32 {
	m.sp = unsafe.Add(m.sp, -4)
	return *(*uint32)(m.sp)
}

func (m *Machine) lu64() uint64 {
	m.sp = unsafe.Add(m.sp, -8)
	return *(*uint64)(m.sp)
}

func (m *Machine) lf64() float64 {
	m.sp = unsafe.Add(m.sp, -8)
	return *(*float64)(m.sp)
}

func (m *Machine) si08(v int8) {
	*(*int8)(m.sp) = v
	m.sp = unsafe.Add(m.sp, 1)
}

func (m *Machine) si64(v int64) {
	*(*int64)(m.sp) = v
	m.sp = unsafe.Add(m.sp, 8)
}

func (m *Machine) su08(v uint8) {
	*(*uint8)(m.sp) = v
	m.sp = unsafe.Add(m.sp, 1)
}

func (m *Machine) su32(v uint32) {
	*(*uint32)(m.sp) = v
	m.sp = unsafe.Add(m.sp, 4)
}

func (m *Machine) su64(v uint64) {
	*(*uint64)(m.sp) = v
	m.sp = unsafe.Add(m.sp, 8)
}

func (m *Machine) sf64(v float64) {
	*(*float64)(m.sp) = v
	m.sp = unsafe.Add(m.sp, 8)
}

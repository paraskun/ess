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
		Call []FuncFrame
	}
)

type Machine struct {
	src []byte
	mem []byte
	sck []byte

	ip unsafe.Pointer
	bp unsafe.Pointer
	sp unsafe.Pointer

	sip unsafe.Pointer
	sbp unsafe.Pointer
	ssp unsafe.Pointer
}

func (m *Machine) Load(src []byte) {
	m.src = src
	m.mem = make([]byte, 0)
	m.sck = make([]byte, 0)

	m.ip = unsafe.Pointer(&m.src[0])
	m.bp = unsafe.Pointer(&m.mem[0])
	m.sp = unsafe.Pointer(&m.sck[0])
}

func (m *Machine) Exec() {
	for {
		cmd := Code(m.nu08())

		switch cmd {
		case JMP:
			m.ip = unsafe.Pointer(&m.src[m.nu32()])
		case JIF:
			if m.lu08() == 0 {
				m.ip = unsafe.Pointer(&m.src[m.nu32()])
			}
		case LB:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su08(*(*uint8)(off))
		case LD:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			m.su64(*(*uint64)(off))
		case SB:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint8)(off) = m.lu08()
		case SD:
			ptr := unsafe.Add(m.bp, m.nu32())
			off := unsafe.Add(unsafe.Pointer(*(*uintptr)(ptr)), m.nu32())

			*(*uint64)(off) = m.lu64()
		case ADDR:
		case ADD:
			m.si64(m.li64() + m.li64())
		case SUB:
			m.si64(m.li64() - m.li64())
		case MUL:
			m.si64(m.li64() * m.li64())
		case DIV:
			m.si64(m.li64() / m.li64())
		case POW:
			m.si64(int64(math.Pow(float64(m.li64()), float64(m.li64()))))
		case SHL:
			m.si64(m.li64() << m.li64())
		case SHR:
			m.si64(m.li64() >> m.li64())
		case MOD:
			m.si64(m.li64() % m.li64())
		case XOR:
			m.si64(m.li64() ^ m.li64())
		case AND:
			m.si64(m.li64() & m.li64())
		case OR:
			m.si64(m.li64() | m.li64())
		case BNEG:
			m.si64(^m.li64())
		case UNEG:
			m.si64(-m.li64())
		case LT:
			if m.li64() < m.li64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case LE:
			if m.li64() <= m.li64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case EQ:
			if m.li64() == m.li64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case NE:
			if m.li64() != m.li64() {
				m.su08(1)
			} else {
				m.su08(0)
			}
		case LAND:
			if m.lu08() == 0 || m.lu08() == 0 {
				m.su08(0)
			} else {
				m.su08(1)
			}
		case LOR:
			if m.lu08() == 0 && m.lu08() == 0 {
				m.su08(0)
			} else {
				m.su08(1)
			}
		case LNEG:
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

func (m *Machine) lu08() uint8 {
	m.sp = unsafe.Add(m.sp, -1)
	return *(*uint8)(m.sp)
}

func (m *Machine) lu64() uint64 {
	m.sp = unsafe.Add(m.sp, -8)
	return *(*uint64)(m.sp)
}

func (m *Machine) li08() int8 {
	m.sp = unsafe.Add(m.sp, -1)
	return *(*int8)(m.sp)
}

func (m *Machine) li64() int64 {
	m.sp = unsafe.Add(m.sp, -8)
	return *(*int64)(m.sp)
}

func (m *Machine) su08(v uint8) {
	*(*uint8)(m.sp) = v
	m.sp = unsafe.Add(m.sp, 1)
}

func (m *Machine) su64(v uint64) {
	*(*uint64)(m.sp) = v
	m.sp = unsafe.Add(m.sp, 8)
}

func (m *Machine) si08(v int8) {
	*(*int8)(m.sp) = v
	m.sp = unsafe.Add(m.sp, 1)
}

func (m *Machine) si64(v int64) {
	*(*int64)(m.sp) = v
	m.sp = unsafe.Add(m.sp, 8)
}

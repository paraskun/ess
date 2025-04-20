package run

import (
	"unsafe"
)

type Heap struct {
	Var []byte
	Str []string

	Call []Heap
}

type Func struct {
	Code []Command
}

type Machine struct {
	Func map[string]Func
	Heap Heap

	ds []byte
	ss []string

	pc []Command
	cc []Command
	ph *Heap
	ch *Heap
	ip int
	rp int
}

func (*Machine) Load(name string)

func (m *Machine) Exec(args []byte) {
	m.ch = &m.Heap
	m.ip = 0

	for {
		cmd := m.cc[m.ip]

		switch cmd {
		case JMP:
			m.ip += 1
			m.ip = m.u32()

			break
		case JIF:
			m.ip += 1
			p := m.u32()

			if m.sGetBol() {
				m.ip = p
			}

			break
		case VADDR:
			m.ip += 1
			idx := m.u32()

			m.sPutPtr((uintptr)(unsafe.Pointer(&m.ch.Var[idx])))

			break
		case GET:
			m.ip += 1

			ix := m.u32()
			sz := m.u8()

			m.ds = append(m.ds, m.ch.Var[ix:(ix+sz)]...)

			break
		case PUT:
			m.ip += 1

			ix := m.u32()
			sz := m.u8()

			copy(m.ds[len(m.ds)-(int)(sz):], m.ch.Var[ix:(ix+sz)])

			break
		case PGET:
			break
		case PPUT:
			break
		case I2F:
			break
		case F2I:
			break
		case IADD:
			break
		case ISUB:
			break
		case IMUL:
			break
		case IDIV:
			break
		case IPOW:
			break
		case ISHL:
			break
		case ISHR:
			break
		case IMOD:
			break
		case IBAND:
			break
		case IBOR:
			break
		case IBXOR:
			break
		case IBNEG:
			break
		case IUNEG:
			break
		case ILT:
			break
		case ILE:
			break
		case INE:
			break
		}
	}
}

func (m *Machine) sGetInt() int64
func (m *Machine) sGetFlt() float64
func (m *Machine) sGetBol() bool
func (m *Machine) sGetPtr() uintptr

func (m *Machine) sPutInt(i int64)
func (m *Machine) sPutFlt(f float64)
func (m *Machine) sPutBol(b bool)
func (m *Machine) sPutPtr(p uintptr)

func (m *Machine) u8() int {
	return (int)(m.cc[m.ip])
}

func (m *Machine) u32() int {
	ptr := unsafe.Pointer(&m.cc[m.ip])
	m.ip += 4
	return (int)(*(*uint)(ptr))
}

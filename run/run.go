package run

import (
	"unsafe"
)

type FuncImage struct {
	Name  string
	ArgSz int
	Call  []string
}

type callInfo struct {
	frame
}

type Frame struct {
	Var  []byte
	Str  []string
	Call []CallInfo
}

type Machine struct {
	Func []FuncImage

	fs []*Frame
	ip []int64
	os []byte

	fp unsafe.Pointer
	op unsafe.Pointer
}

func (m *Machine) Load(name string) {
	// allocate frame
	// fill addresses
}

func (m *Machine) Exec(args []byte) {
	for {
		cmd := m.cc[m.ip]

		switch cmd {
		case JMP:
			m.ip = m.cu32()
			break
		case JIF:
			if *m.sgu08() != 0 {
				m.ip = m.cu32()
			}

			break
		case ADDR:
			m.spptr(&m.Data.Stack[m.cu32()])
			break
		case GET:
			ix := m.cu32()
			sz := m.cu08()

			break
		case PUT:
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

func (m *Machine) sgu08() *uint8
func (m *Machine) sgi64() *int64
func (m *Machine) sgf64() *float64

func (m *Machine) spptr(ptr *byte)

func (m *Machine) cu08() int8
func (m *Machine) cu32() int32

package typ

import "github.com/paraskun/o2/tty"

type Error struct {
	Span tty.Span
	Full string
	Help string
}

func (err *Error) Error() string {
	return err.Full
}

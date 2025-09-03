package typ

import (
	"fmt"
	"io"

	"github.com/paraskun/o2/tty"
)

type Error struct {
	Full string
	Help string
	Snip []*tty.Snippet
}

func (err *Error) Error() string {
	return err.Full
}

func (err *Error) Note(w io.Writer) {
	fmt.Fprintf(w, "o2: %s\n\n", err.Full)

	for _, s := range err.Snip {
		s.Print(w)
	}

	if err.Help != "" {
		fmt.Fprintf(w, "hint: %s\n", err.Help)
	}
}

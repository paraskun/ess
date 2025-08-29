package typ

import (
	"fmt"
	"io"

	"github.com/paraskun/o2/tty"
)

type Error struct {
	Full string
	Help string
	Snip []Location
}

func (err *Error) Error() string {
	return err.Full
}

func (err *Error) Note(w io.Writer) {
	fmt.Fprintf(w, "o2: %s\n\n", err.Full)

	for _, s := range err.Snip {
		s.Span.Reset()

		tty.Print(w, &tty.Frame{
			Name: fmt.Sprintf("%s/%s%s", s.File.Pkg.Mod.Mod.Name, s.File.Pkg.Path, s.File.Name),
			Span: tty.BoxOf(s.Span),
		})
	}

	if err.Help != "" {
		fmt.Fprintf(w, "hint: %s\n", err.Help)
	}
}

package main

import "github.com/paraskun/o2/tty"

func main() {
	tty.Print(&tty.Frame{
		Text: "test/main.o2:1:19",
		Span: &tty.Text{
			Text: "1 + true;",
			Hint: "unsupported operation",
		},
	})
}

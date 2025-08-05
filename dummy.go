package main

import (
	"github.com/fatih/color"
	"github.com/paraskun/o2/tty"
)

func main() {
	tty.Print(&tty.Frame{
		Name: "test/main.o2:1:3",
		Sub: &tty.Box{
			Sub: []tty.Span{
				&tty.Box{
					Hint: &tty.Hint{
						Lit: "in that block",
						Col: color.New(color.FgYellow),
					},
					Sub: []tty.Span{
						&tty.Row{
							Ind: 1,
							Sub: []tty.Mono{
								&tty.Tok{
									Lit: "a",
								},
								&tty.Tok{
									Lit: "=",
									Ind: 1,
								},
								&tty.Row{
									Hint: &tty.Hint{
										Lit: "unsupported operation",
										Col: color.New(color.FgRed),
									},
									Ind: 1,
									Sub: []tty.Mono{
										&tty.Tok{
											Lit: "1",
										},
										&tty.Tok{
											Lit: "+",
											Ind: 1,
										},
										&tty.Tok{
											Lit: "true",
											Ind: 1,
										},
									},
								},
								&tty.Tok{
									Lit: ";",
								},
							},
						},
					},
				},
			},
		},
	})
}

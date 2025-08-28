package ast

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/paraskun/o2/lex"
	"github.com/paraskun/o2/tty"
	"github.com/paraskun/o2/typ"
	"github.com/paraskun/o2/typ/mod"
)

const (
	Indent = 4
)

type Visitor interface {
	VisitStmt(Stmt) *typ.Error
	VisitDecl(Decl) *typ.Error
	VisitExpr(Expr) *typ.Error
}

type Node interface {
	Span() tty.Span
	Accept(Visitor) *typ.Error
}

// Statements

type (
	Stmt interface {
		Node

		stmt()
	}

	BlockStmt struct {
		Box tty.Group
		Sub []Stmt

		Env *typ.Env
	}

	VarStmt struct {
		Box tty.Group
		Var *tty.Tok
		Ini Expr

		Obj *typ.Object
	}

	LetStmt struct {
		Box tty.Group
		Var *tty.Tok
		Ini Expr

		Obj *typ.Object
	}

	AssignStmt struct {
		Box tty.Group
		Var Expr
		Val Expr

		Obj *typ.Object
	}

	LoopStmt struct {
		Box tty.Group
		Con Expr
		Sub *BlockStmt
	}

	CondStmt struct {
		Box tty.Group
		Con Expr
		Pos *BlockStmt
		Neg *BlockStmt
	}

	CallStmt struct {
		*CallExpr
	}

	ReturnStmt struct {
		Box tty.Group
		Ret Expr
	}
)

// Expressions

type (
	Expr interface {
		Node

		Object() *typ.Object

		expr()
	}

	BasicExpr struct {
		Lex *lex.Lexeme

		Obj *typ.Object
	}

	StructFieldExpr struct {
		Box tty.Group
		Mem *tty.Tok
		Val Expr
	}

	StructExpr struct {
		Box tty.Group
		Sym *tty.Tok
		Mem []*StructFieldExpr

		Obj *typ.Object
	}

	IdenExpr struct {
		Tok *tty.Tok

		Obj *typ.Object
	}

	DotExpr struct {
		Box tty.Group
		Mem *tty.Tok
		Ctx Expr

		Obj *typ.Object
	}

	InfExpr struct {
		Box tty.Group
		Lex *lex.Lexeme
		X   Expr
		Y   Expr

		Res *typ.Object
	}

	PfxExpr struct {
		Box tty.Group
		Lex *lex.Lexeme
		X   Expr

		Res *typ.Object
	}

	CallExpr struct {
		Box tty.Group
		Sym *tty.Tok
		Arg []Expr

		Fun *typ.Object
		Ret *typ.Object
	}

	ToI64Expr struct {
		Box tty.Group
		X   Expr

		Res *typ.Object
	}

	ToU64Expr struct {
		Box tty.Group
		X   Expr

		Res *typ.Object
	}

	ToF64Expr struct {
		Box tty.Group
		X   Expr

		Res *typ.Object
	}

	GroupExpr struct {
		Box tty.Group
		Sub Expr
	}
)

// Declarations

type (
	Decl interface {
		Node

		decl()
	}

	UseDecl struct {
		Box tty.Group
		Pkg *tty.Tok

		Obj *typ.Object
	}

	VarDecl struct {
		*VarStmt
	}

	TypeSpec struct {
		Lex *lex.Lexeme
		Typ *typ.Type
	}

	NamedField struct {
		Box tty.Group
		Sym *tty.Tok
		Typ *TypeSpec
	}

	FuncDecl struct {
		Box tty.Group
		Sig *tty.Row
		Sym *tty.Tok
		Arg []*NamedField
		Ret *TypeSpec
		Sub *BlockStmt

		Env *typ.Env
		Obj *typ.Object
	}

	StructDecl struct {
		Box tty.Group
		Sym *tty.Tok
		Mem []*NamedField

		Obj *typ.Object
	}

	EnumDecl struct {
		Box tty.Group
		Sym *tty.Tok
		Mem []*tty.Tok

		Obj *typ.Object
	}
)

type File struct {
	Box *tty.Box
	Dec []Decl
}

type options struct {
	fmt bool
}

type Option func(*options)

func Format() Option {
	return func(o *options) {
		o.fmt = true
	}
}

type parser struct {
	ops *options

	lex lex.Scanner
	buf bool
	prv *lex.Lexeme
	cur *lex.Lexeme
	err *typ.Error

	pkg *mod.Package
	src *mod.File

	box tty.Span
}

func (p *parser) peek() (*lex.Lexeme, *typ.Error) {
	if !p.buf {
		p.buf = true
		p.prv = p.cur
		p.cur, p.err = p.lex.Next()
	}

	return p.cur, p.err
}

func (p *parser) peekIn(g tty.Group, r, b int) (*lex.Lexeme, *typ.Error) {
	tok, err := p.peek()

	if err != nil {
		g.Add(tok.Tok, r, b)
	}

	return tok, err
}

func (p *parser) next() (*lex.Lexeme, *typ.Error) {
	if !p.buf {
		p.prv = p.cur
		p.cur, p.err = p.lex.Next()
	}

	p.buf = false

	return p.cur, p.err
}

func (p *parser) nextIn(g tty.Group, r, b int) (*lex.Lexeme, *typ.Error) {
	tok, err := p.next()
	g.Add(tok.Tok, r, b)

	return tok, err
}

func (p *parser) must(t lex.Type) (*lex.Lexeme, *typ.Error) {
	if !p.buf {
		p.prv = p.cur
		p.cur, p.err = p.lex.Next()
	}

	p.buf = false

	if p.err == nil && p.cur.Typ != t {
		row, col := p.cur.Tok.Pos.Row, p.cur.Tok.Pos.Col

		p.cur.Tok.Hint().Text = fmt.Sprintf("%d:%d - %s expected here", row, col, t)
		p.cur.Tok.Hint().Attr.Color.Add(color.FgRed)

		p.err = &typ.Error{
			Span: p.cur.Tok,
			Full: fmt.Sprintf("%s expected, but %s were given", t, p.cur.Typ),
			Help: "consider following language structure rules",
		}
	}

	return p.cur, p.err
}

func (p *parser) mustIn(t lex.Type, g tty.Group, r, b int) (*lex.Lexeme, *typ.Error) {
	tok, err := p.must(t)
	g.Add(tok.Tok, r, b)

	return tok, err
}

func (p *parser) note(err *typ.Error) {
	fmt.Printf("o2: %v\n\n", err.Full)
	tty.Print(os.Stdout, &tty.Frame{
		Name: p.src.Name,
		Span: &tty.Box{
			Sub: []tty.Span{err.Span},
		},
	})
	fmt.Printf("\nhint: %s\n", err.Help)
}

func Parse(pkg *mod.Package, opts ...Option) {
	ops := &options{}

	for _, opt := range opts {
		opt(ops)
	}

	parse(pkg, ops)
}

func parse(pkg *mod.Package, ops *options) {
	p := &parser{pkg: pkg, ops: ops}

	for _, src := range pkg.Src {
		f, err := os.Open(src.Path)

		if err != nil {
			fmt.Printf("o2: %v\n", err)
			return
		} else {
			defer f.Close()
		}

		buf, err := io.ReadAll(f)

		if err != nil {
			fmt.Printf("o2: %v\n", err)
			return
		}

		dec := &File{
			Dec: make([]Decl, 0),
			Box: &tty.Box{
				Ctl: true,
			},
		}

		p.src = src
		p.lex.Load([]rune(string(buf)))

		for {
			tok, err := p.peek()

			if err != nil {
				p.note(err)
				return
			}

			if tok.Typ == lex.EOF {
				break
			}

			sub, err := p.parseDecl()

			if err != nil {
				p.note(err)
				return
			}

			ind := 1

			if len(dec.Dec) == 0 {
				ind = 0
			}

			dec.Dec = append(dec.Dec, sub)
			dec.Box.Add(sub.Span(), 0, ind)
		}

		src.Dec = dec

		if p.ops.fmt {
			f.Close()

			if err := os.Truncate(src.Path, 0); err != nil {
				fmt.Printf("o2: format: %v\n", err)
				break
			}

			out, err := os.OpenFile(src.Path, os.O_WRONLY, 0644)

			if err != nil {
				fmt.Printf("o2: format: %v\n", err)
				break
			} else {
				defer out.Close()
			}

			tty.Print(out, dec.Box)
		}
	}
}

func (p *parser) parseDecl() (Decl, *typ.Error) {
	tok, err := p.peek()

	if err != nil {
		return nil, err
	}

	switch tok.Typ {
	case lex.USE:
		use, err := p.parseUseDecl()

		if err != nil {
			return nil, err
		}

		pkg := p.pkg.Mod.Lookup(use.Pkg.Lit)

		if pkg == nil {
			use.Pkg.Hint().Text = "could not find this package"
			use.Pkg.Hint().Attr.Color.Add(color.FgRed)

			return nil, &typ.Error{
				Span: use.Span(),
				Full: "package not found",
				Help: "verify package name and module dependencies",
			}
		}

		use.Obj = &typ.Object{
			Seg: typ.Abstract,
			Typ: &typ.Type{
				Kind:  typ.PKG,
				Extra: pkg,
			},
		}

		parse(pkg, p.ops)

		return use, nil
	case lex.VAR:
		dec, err := p.parseVarStmt()

		if err != nil {
			return nil, err
		}

		return &VarDecl{VarStmt: dec}, nil
	case lex.FUNC:
		return p.parseFuncDecl()
	case lex.TYPE:
		return p.parseStructDecl()
	case lex.ENUM:
		return p.parseEnumDecl()
	default:
		row, col := tok.Tok.Pos.Row, tok.Tok.Pos.Col
		tok.Tok.Hint().Text = fmt.Sprintf("%d:%d - declaration expected here", row, col)
		tok.Tok.Hint().Attr.Color.Add(color.FgRed)

		return nil, &typ.Error{
			Span: tok.Tok,
			Full: "unexpected control sequence",
			Help: "consider introducing additional logical scope",
		}
	}
}

func (p *parser) parseUseDecl() (*UseDecl, *typ.Error) {
	dec := &UseDecl{Box: &tty.Row{}}

	if _, err := p.mustIn(lex.USE, dec.Box, 0, 0); err != nil {
		err.Span = dec.Box
		return nil, err
	}

	if tok, err := p.mustIn(lex.ISTR, dec.Box, 1, 0); err != nil {
		err.Span = dec.Box
		return nil, err
	} else {
		dec.Pkg = tok.Tok
	}

	return dec, nil
}

func (p *parser) parseFuncDecl() (*FuncDecl, *typ.Error) {
	dec := &FuncDecl{Sig: &tty.Row{}}
	top := &tty.Row{}

	if _, err := p.mustIn(lex.FUNC, top, 0, 0); err != nil {
		err.Span = top
		return nil, err
	}

	top.Add(dec.Sig, 1, 0)

	if tok, err := p.mustIn(lex.IDEN, dec.Sig, 0, 0); err != nil {
		err.Span = top
		return nil, err
	} else {
		dec.Sym = tok.Tok
	}

	if _, err := p.mustIn(lex.LP, dec.Sig, 0, 0); err != nil {
		err.Span = top
		return nil, err
	}

	for {
		tok, err := p.peekIn(dec.Sig, 0, 0)

		if err != nil {
			err.Span = top
			return nil, err
		}

		if tok.Typ == lex.RP {
			break
		}

		ind := 0

		if len(dec.Arg) != 0 {
			ind = 1
		}

		arg, err := p.parseNamedField(dec.Sig, ind, 0)

		if err != nil {
			err.Span = top
			return nil, err
		}

		dec.Arg = append(dec.Arg, arg)
		tok, err = p.peekIn(dec.Sig, 0, 0)

		if err != nil {
			err.Span = top
			return nil, err
		}

		if tok.Typ != lex.RP {
			if _, err = p.mustIn(lex.COM, dec.Sig, 0, 0); err != nil {
				err.Span = top
				return nil, err
			}
		}
	}

	if _, err := p.mustIn(lex.RP, dec.Sig, 0, 0); err != nil {
		err.Span = top
		return nil, err
	}

	tok, err := p.peekIn(dec.Sig, 1, 0)

	if err != nil {
		err.Span = top
		return nil, err
	}

	if tok.Typ == lex.LP {
		p.nextIn(dec.Sig, 1, 0)

		if dec.Ret, err = p.parseTypeSpec(dec.Sig, 0, 0); err != nil {
			err.Span = top
			return nil, err
		}

		if _, err := p.mustIn(lex.RP, dec.Sig, 0, 0); err != nil {
			err.Span = top
			return nil, err
		}
	}

	tok, err = p.peekIn(top, 1, 0)

	if err != nil {
		err.Span = top
		return nil, err
	}

	if tok.Typ == lex.LB {
		p.nextIn(top, 1, 0)

		dec.Box = &tty.Box{}
		dec.Box.Add(top, 0, 0)
		dec.Sub, err = p.parseBlockStmt()

		if err != nil {
			dec.Box.Add(err.Span, Indent, 0)
			err.Span = dec.Box
			return nil, err
		}

		dec.Box.Add(dec.Sub.Span(), Indent, 0)

		if _, err := p.mustIn(lex.RB, dec.Box, 0, 0); err != nil {
			err.Span = dec.Box
			return nil, err
		}
	} else {
		dec.Box = &tty.Row{}
		dec.Box.Add(top, 0, 0)
	}

	return dec, nil
}

func (p *parser) parseStructDecl() (*StructDecl, *typ.Error) {
	dec := &StructDecl{}
	top := &tty.Row{}

	if _, err := p.mustIn(lex.TYPE, top, 0, 0); err != nil {
		err.Span = top
		return nil, err
	}

	tok, err := p.mustIn(lex.IDEN, top, 1, 0)

	if err != nil {
		err.Span = top
		return nil, err
	}

	dec.Sym = tok.Tok

	if _, err := p.mustIn(lex.LB, top, 1, 0); err != nil {
		err.Span = top
		return nil, err
	}

	tok, err = p.peekIn(top, 1, 0)

	if err != nil {
		err.Span = top
		return nil, err
	}

	if tok.Typ != lex.RB {
		dec.Box = &tty.Box{}
		dec.Box.Add(top, 0, 0)

		for {
			tok, err := p.peekIn(dec.Box, 4, 0)

			if err != nil {
				err.Span = dec.Box
				return nil, err
			}

			if tok.Typ == lex.RB {
				break
			}

			mem, err := p.parseNamedField(dec.Box, 4, 0)

			if err != nil {
				err.Span = dec.Box
				return nil, err
			}

			dec.Mem = append(dec.Mem, mem)
		}
	} else {
		dec.Box = &tty.Row{}
		dec.Box.Add(top, 0, 0)
	}

	if _, err := p.mustIn(lex.RB, dec.Box, 0, 0); err != nil {
		err.Span = dec.Box
		return nil, err
	}

	return dec, nil
}

func (p *parser) parseEnumDecl() (*EnumDecl, *typ.Error) {
	dec := &EnumDecl{}
	top := &tty.Row{}

	if _, err := p.mustIn(lex.ENUM, top, 0, 0); err != nil {
		err.Span = top
		return nil, err
	}

	tok, err := p.mustIn(lex.IDEN, top, 1, 0)

	if err != nil {
		err.Span = top
		return nil, err
	}

	dec.Sym = tok.Tok

	if _, err := p.mustIn(lex.LB, top, 1, 0); err != nil {
		err.Span = top
		return nil, err
	}

	tok, err = p.peekIn(top, 1, 0)

	if err != nil {
		err.Span = top
		return nil, err
	}

	if tok.Typ != lex.RB {
		dec.Box = &tty.Box{}
		dec.Box.Add(top, 0, 0)

		for {
			tok, err := p.peekIn(dec.Box, 4, 0)

			if err != nil {
				err.Span = dec.Box
				return nil, err
			}

			if tok.Typ == lex.RB {
				break
			}

			mem, err := p.mustIn(lex.IDEN, dec.Box, 4, 0)

			if err != nil {
				err.Span = dec.Box
				return nil, err
			}

			dec.Mem = append(dec.Mem, mem.Tok)
		}
	} else {
		dec.Box = &tty.Row{}
		dec.Box.Add(top, 0, 0)
	}

	if _, err := p.mustIn(lex.RB, dec.Box, 0, 0); err != nil {
		err.Span = dec.Box
		return nil, err
	}

	return dec, nil
}

func (p *parser) parseNamedField(g tty.Group, r, b int) (*NamedField, *typ.Error) {
	res := &NamedField{Box: &tty.Row{}}
	g.Add(res.Box, r, b)

	tok, err := p.mustIn(lex.IDEN, res.Box, 0, 0)

	if err != nil {
		err.Span = res.Box
		return nil, err
	}

	res.Sym = tok.Tok
	tsp, err := p.parseTypeSpec(res.Box, 1, 0)

	if err != nil {
		err.Span = res.Box
		return nil, err
	}

	res.Typ = tsp

	return res, nil
}

func (p *parser) parseTypeSpec(g tty.Group, r, b int) (*TypeSpec, *typ.Error) {
	tok, err := p.nextIn(g, r, b)

	if err != nil {
		return nil, err
	}

	switch tok.Typ {
	case lex.BOOL, lex.I64, lex.U64, lex.F64, lex.IDEN:
		return &TypeSpec{Lex: tok}, nil
	}

	row, col := tok.Tok.Pos.Row, tok.Tok.Pos.Col
	tok.Tok.Hint().Text = fmt.Sprintf("%d:%d - type specification expected here", row, col)
	tok.Tok.Hint().Attr.Color.Add(color.FgRed)

	return nil, &typ.Error{
		Span: tok.Tok,
		Full: "unexpected control sequence",
		Help: "consider specifying existing type",
	}
}

func (p *parser) parseBlockStmt() (*BlockStmt, *typ.Error) {
	res := &BlockStmt{}
	tok, err := p.peek()

	if err != nil {
		return nil, err
	}

	if tok.Typ != lex.RB {
		res.Box = &tty.Box{}

		for {
			tok, err := p.peekIn(res.Box, 0, 0)

			if err != nil {
				err.Span = res.Box
				return nil, err
			}

			if tok.Typ == lex.RB {
				break
			}

			sub, err := p.parseStmt()

			if err != nil {
				res.Box.Add(err.Span, 0, 0)
				err.Span = res.Box
				return nil, err
			}

			res.Box.Add(sub.Span(), 0, 0)
			res.Sub = append(res.Sub, sub)
		}
	} else {
		res.Box = &tty.Row{}
	}

	return res, nil
}

func (p *parser) parseStmt() (Stmt, *typ.Error) {
	tok, err := p.peek()

	if err != nil {
		return nil, err
	}

	switch tok.Typ {
	case lex.IDEN:
		group := &tty.Row{}
		iden, err := p.nextIn(group, 0, 0)

		if err != nil {
			err.Span = group
			return nil, err
		}

		tok, err = p.peekIn(group, 1, 0)

		if err != nil {
			err.Span = group
			return nil, err
		}

		if tok.Typ == lex.LP {
			res, err := p.parseCall(iden.Tok)

			if err != nil {
				return nil, err
			}

			tok, err := p.must(lex.SEM)

			if err != nil {
				res.Box.InsertEnd(err.Span.(tty.Mono), 0)
				err.Span = res.Box
				return nil, err
			}

			res.Box.InsertEnd(tok.Tok, 0)

			return &CallStmt{res}, nil
		}

		var cur Expr = &IdenExpr{Tok: iden.Tok}

		for {
			tok, err := p.peekIn(group, 1, 0)

			if err != nil {
				err.Span = group
				return nil, err
			}

			switch tok.Typ {
			case lex.DOT:
				dot := &DotExpr{
					Box: &tty.Row{},
					Ctx: cur,
				}

				dot.Box.Add(dot.Ctx.Span(), 0, 0)

				if _, err := p.nextIn(dot.Box, 0, 0); err != nil {
					err.Span = dot.Box
					return nil, err
				}

				tok, err := p.mustIn(lex.IDEN, dot.Box, 0, 0)

				if err != nil {
					err.Span = dot.Box
					return nil, err
				}

				dot.Mem = tok.Tok
				cur = dot

				continue
			}

			break
		}

		res := &AssignStmt{Var: cur}
		top := &tty.Row{}

		top.Add(cur.Span(), 0, 0)

		if _, err := p.mustIn(lex.EQ, top, 1, 0); err != nil {
			err.Span = top
			return nil, err
		}

		res.Val, err = p.parseExpr0()

		if err != nil {
			switch err.Span.(type) {
			case *tty.Box:
				res.Box = &tty.Box{}
				res.Box.Add(top, 0, 0)
				res.Box.Add(err.Span, Indent, 0)
			case tty.Mono:
				res.Box = &tty.Row{}
				res.Box.Add(top, 0, 0)
				res.Box.Add(err.Span, 1, 0)
			}

			err.Span = res.Box

			return nil, err
		}

		switch b := res.Val.Span().(type) {
		case *tty.Box:
			res.Box = &tty.Box{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(b, Indent, 0)

			tok, err := p.must(lex.SEM)
			b.InsertEnd(tok.Tok, 0)

			if err != nil {
				err.Span = res.Box
				return nil, err
			}
		case tty.Mono:
			res.Box = &tty.Row{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(b, 1, 0)

			if _, err := p.mustIn(lex.SEM, res.Box, 0, 0); err != nil {
				err.Span = res.Box
				return nil, err
			}
		}

		return res, nil
	case lex.VAR:
		return p.parseVarStmt()
	case lex.LET:
		return p.parseLetStmt()
	case lex.FOR:
		return p.parseLoopStmt()
	case lex.IF:
		return p.parseCondStmt()
	case lex.RET:
		return p.parseReturnStmt()
	}

	row, col := tok.Tok.Pos.Row, tok.Tok.Pos.Col
	tok.Tok.Hint().Text = fmt.Sprintf("%d:%d - statement expected here", row, col)
	tok.Tok.Hint().Attr.Color.Add(color.FgRed)

	return nil, &typ.Error{
		Span: tok.Tok,
		Full: "unexpected control sequence",
		Help: "consider introducing new statement",
	}
}

func (p *parser) parseCall(sym *tty.Tok) (*CallExpr, *typ.Error) {
	res := &CallExpr{}
	top := &tty.Row{}

	top.Add(sym, 0, 0)

	if _, err := p.nextIn(top, 0, 0); err != nil {
		err.Span = top
		return nil, err
	}

	box := &tty.Box{}
	row := top

	for {
		tok, err := p.peekIn(top, 1, 0)

		if err != nil {
			err.Span = top
			return nil, err
		}

		if tok.Typ == lex.RP {
			break
		}

		arg, err := p.parseExpr0()

		if err != nil {
			switch b := err.Span.(type) {
			case *tty.Box:
				if len(row.Sub) != 0 {
					box.Add(row, Indent, 0)
				}

				box.Add(b, Indent, 0)
				res.Box = box
			case tty.Mono:
				if len(row.Sub) != 0 {
					row.Add(b, 1, 0)
				} else {
					row.Add(b, 0, 0)
				}

				if len(box.Sub) != 0 {
					box.Add(row, Indent, 0)
					res.Box = box
				} else {
					res.Box = row
				}
			}

			err.Span = res.Box

			return nil, err
		}

		switch b := arg.Span().(type) {
		case *tty.Box:
			if len(row.Sub) != 0 {
				box.Add(row, 4, 0)
			}

			box.Add(b, 4, 0)
			tok, err := p.peek()

			if err != nil {
				box.InsertEnd(tok.Tok, 0)
				err.Span = box

				return nil, err
			}

			if tok.Typ != lex.RP {
				tok, err := p.must(lex.COM)
				box.InsertEnd(tok.Tok, 0)

				if err != nil {
					err.Span = box
					return nil, err
				}

				row = &tty.Row{}
			} else {
				tok, err := p.must(lex.RP)
				box.InsertEnd(tok.Tok, 0)

				if err != nil {
					err.Span = box
					return nil, err
				}
			}
		case tty.Mono:
			if len(row.Sub) != 0 {
				row.Add(b, 1, 0)
			} else {
				row.Add(b, 0, 0)
			}

			tok, err := p.peekIn(row, 1, 0)

			if err != nil {
				return nil, err
			}

			if tok.Typ != lex.RP {
				if _, err := p.mustIn(lex.COM, row, 0, 0); err != nil {
					return nil, err
				}
			} else {
				if _, err := p.mustIn(lex.RP, row, 0, 0); err != nil {
					return nil, err
				}
			}
		}

		res.Arg = append(res.Arg, arg)
	}

	if len(box.Sub) != 0 {
		res.Box = box
	} else {
		res.Box = row
	}

	return res, nil
}

func (p *parser) parseVarStmt() (*VarStmt, *typ.Error) {
	top := &tty.Row{}

	if _, err := p.mustIn(lex.VAR, top, 0, 0); err != nil {
		err.Span = top
		return nil, err
	}

	tok, err := p.mustIn(lex.IDEN, top, 1, 0)

	if err != nil {
		err.Span = top
		return nil, err
	}

	res := &VarStmt{Var: tok.Tok}

	if _, err := p.mustIn(lex.EQ, top, 1, 0); err != nil {
		err.Span = top
		return nil, err
	}

	res.Ini, err = p.parseExpr0()

	if err != nil {
		switch b := err.Span.(type) {
		case *tty.Box:
			res.Box = &tty.Box{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(b, Indent, 0)
		case tty.Mono:
			res.Box = &tty.Row{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(b, 1, 0)
		}

		err.Span = res.Box

		return nil, err
	}

	switch b := res.Ini.Span().(type) {
	case *tty.Box:
		res.Box = &tty.Box{}
		res.Box.Add(top, 0, 0)
		res.Box.Add(b, 4, 0)
	case tty.Mono:
		res.Box = &tty.Row{}
		res.Box.Add(top, 0, 0)
		res.Box.Add(b, 1, 0)
	}

	tok, err = p.must(lex.SEM)
	res.Box.InsertEnd(tok.Tok, 0)

	if err != nil {
		err.Span = res.Box
		return nil, err
	}

	return res, nil
}

func (p *parser) parseLetStmt() (*LetStmt, *typ.Error) {
	top := &tty.Row{}

	if _, err := p.mustIn(lex.LET, top, 0, 0); err != nil {
		err.Span = top
		return nil, err
	}

	tok, err := p.mustIn(lex.IDEN, top, 1, 0)

	if err != nil {
		err.Span = top
		return nil, err
	}

	res := &LetStmt{Var: tok.Tok}

	if _, err := p.mustIn(lex.EQ, top, 1, 0); err != nil {
		err.Span = top
		return nil, err
	}

	res.Ini, err = p.parseExpr0()

	if err != nil {
		switch b := err.Span.(type) {
		case *tty.Box:
			res.Box = &tty.Box{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(b, Indent, 0)
		case tty.Mono:
			res.Box = &tty.Row{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(b, 1, 0)
		}

		err.Span = res.Box

		return nil, err
	}

	switch b := res.Ini.Span().(type) {
	case *tty.Box:
		res.Box = &tty.Box{}
		res.Box.Add(top, 0, 0)
		res.Box.Add(b, 4, 0)
	case tty.Mono:
		res.Box = &tty.Row{}
		res.Box.Add(top, 0, 0)
		res.Box.Add(b, 1, 0)
	}

	tok, err = p.must(lex.SEM)
	res.Box.InsertEnd(tok.Tok, 0)

	if err != nil {
		err.Span = res.Box
		return nil, err
	}

	return res, nil
}

func (p *parser) parseLoopStmt() (*LoopStmt, *typ.Error) {
	res := &LoopStmt{Box: &tty.Box{}}
	tok, err := p.must(lex.FOR)

	if err != nil {
		return nil, err
	}

	res.Con, err = p.parseExpr0()

	if err != nil {
		switch b := err.Span.(type) {
		case *tty.Box:
			res.Box.Add(tok.Tok, 0, 0)
			res.Box.Add(b, Indent, 0)
		case tty.Mono:
			row := &tty.Row{}
			row.Add(tok.Tok, 0, 0)
			row.Add(b, 1, 0)
			res.Box.Add(row, 0, 0)
		}

		err.Span = res.Box

		return nil, err
	}

	var log tty.Group

	switch res.Con.Span().(type) {
	case *tty.Box:
		log = &tty.Box{}
		log.Add(tok.Tok, 0, 0)
		log.Add(res.Con.Span(), 4, 0)
		res.Box.Add(log, 0, 0)

		if _, err := p.mustIn(lex.LB, log, 0, 0); err != nil {
			err.Span = res.Box
			return nil, err
		}
	case tty.Mono:
		log = &tty.Row{}
		log.Add(tok.Tok, 0, 0)
		log.Add(res.Con.Span(), 1, 0)
		res.Box.Add(log, 0, 0)

		if _, err := p.mustIn(lex.LB, log, 1, 0); err != nil {
			err.Span = res.Box
			return nil, err
		}
	}

	res.Sub, err = p.parseBlockStmt()

	if err != nil {
		res.Box.Add(err.Span, Indent, 0)
		err.Span = res.Box
		return nil, err
	}

	res.Box.Add(res.Sub.Span(), Indent, 0)

	if _, err := p.mustIn(lex.RB, res.Box, 0, 0); err != nil {
		err.Span = res.Box
		return nil, err
	}

	return res, nil
}

func (p *parser) parseCondStmt() (*CondStmt, *typ.Error) {
	res := &CondStmt{Box: &tty.Box{}}
	tok, err := p.must(lex.IF)

	if err != nil {
		return nil, err
	}

	res.Con, err = p.parseExpr0()

	if err != nil {
		switch b := err.Span.(type) {
		case *tty.Box:
			res.Box.Add(tok.Tok, 0, 0)
			res.Box.Add(b, Indent, 0)
		case tty.Mono:
			row := &tty.Row{}
			row.Add(tok.Tok, 0, 0)
			row.Add(b, 1, 0)
			res.Box.Add(row, 0, 0)
		}

		err.Span = res.Box

		return nil, err
	}

	var log tty.Group

	switch res.Con.Span().(type) {
	case *tty.Box:
		log = &tty.Box{}
		log.Add(tok.Tok, 0, 0)
		log.Add(res.Con.Span(), 4, 0)
		res.Box.Add(log, 0, 0)

		if _, err := p.mustIn(lex.LB, log, 0, 0); err != nil {
			err.Span = res.Box
			return nil, err
		}
	case tty.Mono:
		log = &tty.Row{}
		log.Add(tok.Tok, 0, 0)
		log.Add(res.Con.Span(), 1, 0)
		res.Box.Add(log, 0, 0)

		if _, err := p.mustIn(lex.LB, log, 1, 0); err != nil {
			err.Span = res.Box
			return nil, err
		}
	}

	res.Pos, err = p.parseBlockStmt()

	if err != nil {
		res.Box.Add(err.Span, Indent, 0)
		err.Span = res.Box
		return nil, err
	}

	res.Box.Add(res.Pos.Span(), Indent, 0)

	if _, err := p.mustIn(lex.RB, res.Box, 0, 0); err != nil {
		err.Span = res.Box
		return nil, err
	}

	tok, err = p.peek()

	if err != nil {
		res.Box.InsertEnd(tok.Tok, 1)
		err.Span = res.Box

		return nil, err
	}

	if tok.Typ == lex.ELSE {
		tok, _ = p.next()
		res.Box.InsertEnd(tok.Tok, 1)

		tok, err = p.must(lex.LB)
		res.Box.InsertEnd(tok.Tok, 1)

		if err != nil {
			err.Span = res.Box
			return nil, err
		}

		res.Neg, err = p.parseBlockStmt()

		if err != nil {
			res.Box.Add(err.Span, Indent, 0)
			err.Span = res.Box
			return nil, err
		}

		res.Box.Add(res.Neg.Span(), Indent, 0)

		if _, err := p.mustIn(lex.RB, res.Box, 0, 0); err != nil {
			err.Span = res.Box
			return nil, err
		}
	}

	return res, nil
}

func (p *parser) parseReturnStmt() (*ReturnStmt, *typ.Error) {
	res := &ReturnStmt{}
	top := &tty.Row{}

	if _, err := p.mustIn(lex.RET, top, 0, 0); err != nil {
		return nil, err
	}

	tok, err := p.peekIn(top, 1, 0)

	if err != nil {
		err.Span = top
		return nil, err
	}

	if tok.Typ == lex.SEM {
		res.Box = top
	} else {
		res.Ret, err = p.parseExpr0()

		if err != nil {
			switch b := err.Span.(type) {
			case *tty.Box:
				box := &tty.Box{}
				box.Add(top, 0, 0)
				box.Add(b, Indent, 0)
				res.Box = box
			case tty.Mono:
				top.Add(b, 1, 0)
				res.Box = top
			}

			err.Span = res.Box

			return nil, err
		}

		switch rb := res.Ret.Span().(type) {
		case *tty.Box:
			box := &tty.Box{}
			box.Add(top, 0, 0)
			box.Add(rb, Indent, 0)
			res.Box = box
		case tty.Mono:
			top.Add(rb, 1, 0)
			res.Box = top
		}

	}

	tok, err = p.must(lex.SEM)
	res.Box.InsertEnd(tok.Tok, 0)

	if err != nil {
		err.Span = res.Box
		return nil, err
	}

	return res, nil
}

func (p *parser) parseExpr0() (Expr, *typ.Error) {
	cur, err := p.parseExpr1()

	if err != nil {
		return nil, err
	}

	for {
		tok, err := p.peek()

		if err != nil {
			box := &tty.Box{}
			box.Add(cur.Span(), 0, 0)
			box.InsertEnd(err.Span.(tty.Mono), 1)

			err.Span = box

			return nil, err
		}

		switch tok.Typ {
		case lex.LAND, lex.LOR:
			tok, _ := p.next()
			res := &InfExpr{Lex: tok, X: cur}

			res.Y, err = p.parseExpr1()

			if err != nil {
				switch yb := err.Span.(type) {
				case *tty.Box:
					yb.InsertBeg(tok.Tok, 1)

					res.Box = &tty.Box{}
					res.Box.Add(cur.Span(), 0, 0)
					res.Box.Add(yb, Indent, 0)
				case tty.Mono:
					row := &tty.Row{}
					row.Add(tok.Tok, 0, 0)
					row.Add(yb, 1, 0)

					switch xb := cur.Span().(type) {
					case *tty.Box:
						xb.InsertEnd(row, 1)
						res.Box = xb
					case tty.Mono:
						res.Box = &tty.Row{}
						res.Box.Add(xb, 0, 0)
						res.Box.Add(row, 1, 0)
					}
				}

				err.Span = res.Box

				return nil, err
			}

			switch yb := res.Y.Span().(type) {
			case *tty.Box:
				yb.InsertBeg(tok.Tok, 1)

				res.Box = &tty.Box{}
				res.Box.Add(cur.Span(), 0, 0)
				res.Box.Add(yb, 4, 0)
			case tty.Mono:
				row := &tty.Row{}
				row.Add(tok.Tok, 0, 0)
				row.Add(yb, 1, 0)

				switch xb := cur.Span().(type) {
				case *tty.Box:
					xb.InsertEnd(row, 1)

					res.Box = xb
				case tty.Mono:
					res.Box = &tty.Row{}
					res.Box.Add(xb, 0, 0)
					res.Box.Add(row, 1, 0)
				}
			}

			cur = res

			continue
		}

		break
	}

	return cur, nil
}

func (p *parser) parseExpr1() (Expr, *typ.Error) {
	cur, err := p.parseExpr2()

	if err != nil {
		return nil, err
	}

	for {
		tok, err := p.peek()

		if err != nil {
			box := &tty.Box{}
			box.Add(cur.Span(), 0, 0)
			box.InsertEnd(err.Span.(tty.Mono), 1)

			err.Span = box

			return nil, err
		}

		switch tok.Typ {
		case lex.LT, lex.GT, lex.LE, lex.GE, lex.NE, lex.EEQ:
			tok, _ := p.next()
			res := &InfExpr{Lex: tok, X: cur}

			res.Y, err = p.parseExpr2()

			if err != nil {
				switch yb := err.Span.(type) {
				case *tty.Box:
					yb.InsertBeg(tok.Tok, 1)

					res.Box = &tty.Box{}
					res.Box.Add(cur.Span(), 0, 0)
					res.Box.Add(yb, Indent, 0)
				case tty.Mono:
					row := &tty.Row{}
					row.Add(tok.Tok, 0, 0)
					row.Add(yb, 1, 0)

					switch xb := cur.Span().(type) {
					case *tty.Box:
						xb.InsertEnd(row, 1)
						res.Box = xb
					case tty.Mono:
						res.Box = &tty.Row{}
						res.Box.Add(xb, 0, 0)
						res.Box.Add(row, 1, 0)
					}
				}

				err.Span = res.Box

				return nil, err
			}

			switch yb := res.Y.Span().(type) {
			case *tty.Box:
				yb.InsertBeg(tok.Tok, 1)

				res.Box = &tty.Box{}
				res.Box.Add(cur.Span(), 0, 0)
				res.Box.Add(yb, 4, 0)
			case tty.Mono:
				row := &tty.Row{}
				row.Add(tok.Tok, 0, 0)
				row.Add(yb, 1, 0)

				switch xb := cur.Span().(type) {
				case *tty.Box:
					xb.InsertEnd(row, 1)

					res.Box = xb
				case tty.Mono:
					res.Box = &tty.Row{}
					res.Box.Add(xb, 0, 0)
					res.Box.Add(row, 1, 0)
				}
			}

			cur = res

			continue
		}

		break
	}

	return cur, nil

}

func (p *parser) parseExpr2() (Expr, *typ.Error) {
	cur, err := p.parseExpr3()

	if err != nil {
		return nil, err
	}

	for {
		tok, err := p.peek()

		if err != nil {
			box := &tty.Box{}
			box.Add(cur.Span(), 0, 0)
			box.InsertEnd(err.Span.(tty.Mono), 1)

			err.Span = box

			return nil, err
		}

		switch tok.Typ {
		case lex.SHL, lex.SHR, lex.BAND, lex.BOR, lex.BXOR:
			tok, _ := p.next()
			res := &InfExpr{Lex: tok, X: cur}

			res.Y, err = p.parseExpr3()

			if err != nil {
				switch yb := err.Span.(type) {
				case *tty.Box:
					yb.InsertBeg(tok.Tok, 1)

					res.Box = &tty.Box{}
					res.Box.Add(cur.Span(), 0, 0)
					res.Box.Add(yb, Indent, 0)
				case tty.Mono:
					row := &tty.Row{}
					row.Add(tok.Tok, 0, 0)
					row.Add(yb, 1, 0)

					switch xb := cur.Span().(type) {
					case *tty.Box:
						xb.InsertEnd(row, 1)
						res.Box = xb
					case tty.Mono:
						res.Box = &tty.Row{}
						res.Box.Add(xb, 0, 0)
						res.Box.Add(row, 1, 0)
					}
				}

				err.Span = res.Box

				return nil, err
			}

			switch yb := res.Y.Span().(type) {
			case *tty.Box:
				yb.InsertBeg(tok.Tok, 1)

				res.Box = &tty.Box{}
				res.Box.Add(cur.Span(), 0, 0)
				res.Box.Add(yb, 4, 0)
			case tty.Mono:
				row := &tty.Row{}
				row.Add(tok.Tok, 0, 0)
				row.Add(yb, 1, 0)

				switch xb := cur.Span().(type) {
				case *tty.Box:
					xb.InsertEnd(row, 1)

					res.Box = xb
				case tty.Mono:
					res.Box = &tty.Row{}
					res.Box.Add(xb, 0, 0)
					res.Box.Add(row, 1, 0)
				}
			}

			cur = res

			continue
		}

		break
	}

	return cur, nil
}

func (p *parser) parseExpr3() (Expr, *typ.Error) {
	cur, err := p.parseExpr4()

	if err != nil {
		return nil, err
	}

	for {
		tok, err := p.peek()

		if err != nil {
			box := &tty.Box{}
			box.Add(cur.Span(), 0, 0)
			box.InsertEnd(err.Span.(tty.Mono), 1)

			err.Span = box

			return nil, err
		}

		switch tok.Typ {
		case lex.ADD, lex.SUB:
			tok, _ := p.next()
			res := &InfExpr{Lex: tok, X: cur}

			res.Y, err = p.parseExpr4()

			if err != nil {
				switch yb := err.Span.(type) {
				case *tty.Box:
					yb.InsertBeg(tok.Tok, 1)

					res.Box = &tty.Box{}
					res.Box.Add(cur.Span(), 0, 0)
					res.Box.Add(yb, Indent, 0)
				case tty.Mono:
					row := &tty.Row{}
					row.Add(tok.Tok, 0, 0)
					row.Add(yb, 1, 0)

					switch xb := cur.Span().(type) {
					case *tty.Box:
						xb.InsertEnd(row, 1)
						res.Box = xb
					case tty.Mono:
						res.Box = &tty.Row{}
						res.Box.Add(xb, 0, 0)
						res.Box.Add(row, 1, 0)
					}
				}

				err.Span = res.Box

				return nil, err
			}

			switch yb := res.Y.Span().(type) {
			case *tty.Box:
				yb.InsertBeg(tok.Tok, 1)

				res.Box = &tty.Box{}
				res.Box.Add(cur.Span(), 0, 0)
				res.Box.Add(yb, 4, 0)
			case tty.Mono:
				row := &tty.Row{}
				row.Add(tok.Tok, 0, 0)
				row.Add(yb, 1, 0)

				switch xb := cur.Span().(type) {
				case *tty.Box:
					xb.InsertEnd(row, 1)

					res.Box = xb
				case tty.Mono:
					res.Box = &tty.Row{}
					res.Box.Add(xb, 0, 0)
					res.Box.Add(row, 1, 0)
				}
			}

			cur = res

			continue
		}

		break
	}

	return cur, nil
}

func (p *parser) parseExpr4() (Expr, *typ.Error) {
	cur, err := p.parseExpr5()

	if err != nil {
		return nil, err
	}

	for {
		tok, err := p.peek()

		if err != nil {
			box := &tty.Box{}
			box.Add(cur.Span(), 0, 0)
			box.InsertEnd(err.Span.(tty.Mono), 1)

			err.Span = box

			return nil, err
		}

		switch tok.Typ {
		case lex.MUL, lex.DIV, lex.MOD, lex.POW:
			tok, _ := p.next()
			res := &InfExpr{Lex: tok, X: cur}

			res.Y, err = p.parseExpr5()

			if err != nil {
				switch yb := err.Span.(type) {
				case *tty.Box:
					yb.InsertBeg(tok.Tok, 1)

					res.Box = &tty.Box{}
					res.Box.Add(cur.Span(), 0, 0)
					res.Box.Add(yb, Indent, 0)
				case tty.Mono:
					row := &tty.Row{}
					row.Add(tok.Tok, 0, 0)
					row.Add(yb, 1, 0)

					switch xb := cur.Span().(type) {
					case *tty.Box:
						xb.InsertEnd(row, 1)
						res.Box = xb
					case tty.Mono:
						res.Box = &tty.Row{}
						res.Box.Add(xb, 0, 0)
						res.Box.Add(row, 1, 0)
					}
				}

				err.Span = res.Box

				return nil, err
			}

			switch yb := res.Y.Span().(type) {
			case *tty.Box:
				yb.InsertBeg(tok.Tok, 1)

				res.Box = &tty.Box{}
				res.Box.Add(cur.Span(), 0, 0)
				res.Box.Add(yb, 4, 0)
			case tty.Mono:
				row := &tty.Row{}
				row.Add(tok.Tok, 0, 0)
				row.Add(yb, 1, 0)

				switch xb := cur.Span().(type) {
				case *tty.Box:
					xb.InsertEnd(row, 1)

					res.Box = xb
				case tty.Mono:
					res.Box = &tty.Row{}
					res.Box.Add(xb, 0, 0)
					res.Box.Add(row, 1, 0)
				}
			}

			cur = res

			continue
		}

		break
	}

	return cur, nil
}

func (p *parser) parseExpr5() (Expr, *typ.Error) {
	tok, err := p.peek()

	if err != nil {
		return nil, err
	}

	switch tok.Typ {
	case lex.LNEG, lex.BNEG, lex.UNEG:
		tok, _ := p.next()
		res := &PfxExpr{Lex: tok}

		res.X, err = p.parseExpr5()

		if err != nil {
			switch b := err.Span.(type) {
			case *tty.Box:
				b.InsertBeg(tok.Tok, 0)
				res.Box = b
			case tty.Mono:
				res.Box = &tty.Row{}
				res.Box.Add(tok.Tok, 0, 0)
				res.Box.Add(b, 0, 0)
			}

			err.Span = res.Box

			return nil, err
		}

		switch sb := res.X.Span().(type) {
		case *tty.Box:
			sb.InsertBeg(tok.Tok, 0)
			res.Box = sb
		case tty.Mono:
			res.Box = &tty.Row{}
			res.Box.Add(tok.Tok, 0, 0)
			res.Box.Add(sb, 0, 0)
		}

		return res, nil
	}

	return p.parseExpr6()
}

func (p *parser) parseExpr6() (Expr, *typ.Error) {
	cur, err := p.parseExpr7()

	if err != nil {
		return nil, err
	}

	for {
		tok, err := p.peek()

		if err != nil {
			box := &tty.Box{}
			box.Add(cur.Span(), 0, 0)
			box.InsertEnd(err.Span.(tty.Mono), 1)

			err.Span = box

			return nil, err
		}

		switch tok.Typ {
		case lex.DOT:
			row := &tty.Row{}
			p.nextIn(row, 0, 0)

			mem, err := p.mustIn(lex.IDEN, row, 0, 0)

			if err != nil {
				switch cb := cur.Span().(type) {
				case *tty.Box:
					cb.InsertEnd(row, 0)
					err.Span = cb
				case tty.Mono:
					box := &tty.Row{}
					box.Add(cb, 0, 0)
					box.Add(row, 0, 0)
					err.Span = box
				}

				return nil, err
			}

			res := &DotExpr{
				Ctx: cur,
				Mem: mem.Tok,
			}

			switch cb := cur.Span().(type) {
			case *tty.Box:
				cb.InsertEnd(row, 0)
				res.Box = cb
			case tty.Mono:
				res.Box = &tty.Row{}
				res.Box.Add(cb, 0, 0)
				res.Box.Add(row, 0, 0)
			}

			cur = res

			continue
		}

		break
	}

	return cur, nil
}

func (p *parser) parseExpr7() (Expr, *typ.Error) {
	tok, err := p.peek()

	if err != nil {
		return nil, err
	}

	switch tok.Typ {
	case lex.IDEN:
		top := &tty.Row{}
		iden, _ := p.nextIn(top, 0, 0)
		tok, err = p.peekIn(top, 1, 0)

		if err != nil {
			err.Span = top
			return nil, err
		}

		switch tok.Typ {
		case lex.LP:
			return p.parseCall(iden.Tok)
		case lex.LB:
			res := &StructExpr{Sym: iden.Tok}

			_, _ = p.nextIn(top, 0, 0)
			tok, err = p.peekIn(top, 1, 0)

			if err != nil {
				err.Span = top
				return nil, err
			}

			if tok.Typ != lex.RB {
				res.Box = &tty.Box{}
				res.Box.Add(top, 0, 0)

				// res.Box.Hint().Text = "right here"
				// res.Box.Hint().Attr.Color.Add(color.FgRed)

				for tok.Typ != lex.RB {
					mem, err := p.parseStructFieldExpr()

					if err != nil {
						res.Box.Add(err.Span, Indent, 0)
						err.Span = res.Box
						return nil, err
					}

					res.Box.Add(mem.Box, Indent, 0)
					res.Mem = append(res.Mem, mem)

					tok, err = p.peekIn(res.Box, 0, 0)

					if err != nil {
						err.Span = res.Box
						return nil, err
					}

					if tok.Typ != lex.RB {
						com, err := p.mustIn(lex.COM, nil, 0, 0)

						if err != nil {
							mem.Box.InsertEnd(err.Span.(tty.Mono), 1)
							err.Span = res.Box
							return nil, err
						}

						mem.Box.InsertEnd(com.Tok, 0)
					}
				}
			} else {
				res.Box = top
			}

			if _, err := p.mustIn(lex.RB, res.Box, 0, 0); err != nil {
				err.Span = res.Box
				return nil, err
			}

			return res, nil
		}

		return &IdenExpr{Tok: iden.Tok}, nil
	case lex.II64, lex.IU64, lex.IF64, lex.TRUE, lex.FALSE:
		tok, _ := p.next()
		return &BasicExpr{Lex: tok}, nil
	case lex.I64:
		res := &ToI64Expr{}
		top := &tty.Row{}

		if _, err := p.nextIn(top, 0, 0); err != nil {
			err.Span = top
			return nil, err
		}

		if _, err := p.mustIn(lex.LP, top, 0, 0); err != nil {
			err.Span = top
			return nil, err
		}

		res.X, err = p.parseExpr0()

		if err != nil {
			switch b := err.Span.(type) {
			case *tty.Box:
				res.Box = &tty.Box{}
				res.Box.Add(top, 0, 0)
				res.Box.Add(b, Indent, 0)
			case tty.Mono:
				res.Box = &tty.Row{}
				res.Box.Add(top, 0, 0)
				res.Box.Add(b, 0, 0)
			}

			err.Span = res.Box

			return nil, err
		}

		switch xb := res.X.Span().(type) {
		case *tty.Box:
			res.Box = &tty.Box{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(xb, 4, 0)
		case tty.Mono:
			res.Box = &tty.Row{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(xb, 0, 0)
		}

		if _, err := p.mustIn(lex.RP, res.Box, 0, 0); err != nil {
			err.Span = res.Box
			return nil, err
		}

		return res, nil
	case lex.U64:
		res := &ToU64Expr{}
		top := &tty.Row{}

		if _, err := p.nextIn(top, 0, 0); err != nil {
			err.Span = top
			return nil, err
		}

		if _, err := p.mustIn(lex.LP, top, 0, 0); err != nil {
			err.Span = top
			return nil, err
		}

		res.X, err = p.parseExpr0()

		if err != nil {
			switch b := err.Span.(type) {
			case *tty.Box:
				res.Box = &tty.Box{}
				res.Box.Add(top, 0, 0)
				res.Box.Add(b, Indent, 0)
			case tty.Mono:
				res.Box = &tty.Row{}
				res.Box.Add(top, 0, 0)
				res.Box.Add(b, 0, 0)
			}

			err.Span = res.Box

			return nil, err
		}

		switch xb := res.X.Span().(type) {
		case *tty.Box:
			res.Box = &tty.Box{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(xb, 4, 0)
		case tty.Mono:
			res.Box = &tty.Row{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(xb, 0, 0)
		}

		if _, err := p.mustIn(lex.RP, res.Box, 0, 0); err != nil {
			err.Span = res.Box
			return nil, err
		}

		return res, nil
	case lex.F64:
		res := &ToF64Expr{}
		top := &tty.Row{}

		if _, err := p.nextIn(top, 0, 0); err != nil {
			err.Span = top
			return nil, err
		}

		if _, err := p.mustIn(lex.LP, top, 0, 0); err != nil {
			err.Span = top
			return nil, err
		}

		res.X, err = p.parseExpr0()

		if err != nil {
			switch b := err.Span.(type) {
			case *tty.Box:
				res.Box = &tty.Box{}
				res.Box.Add(top, 0, 0)
				res.Box.Add(b, Indent, 0)
			case tty.Mono:
				res.Box = &tty.Row{}
				res.Box.Add(top, 0, 0)
				res.Box.Add(b, 0, 0)
			}

			err.Span = res.Box

			return nil, err
		}

		switch xb := res.X.Span().(type) {
		case *tty.Box:
			res.Box = &tty.Box{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(xb, 4, 0)
		case tty.Mono:
			res.Box = &tty.Row{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(xb, 0, 0)
		}

		if _, err := p.mustIn(lex.RP, res.Box, 0, 0); err != nil {
			err.Span = res.Box
			return nil, err
		}

		return res, nil
	case lex.LP:
		res := &GroupExpr{}
		tok, _ := p.next()
		res.Sub, err = p.parseExpr0()

		if err != nil {
			switch b := err.Span.(type) {
			case *tty.Box:
				res.Box = &tty.Box{}
				res.Box.Add(tok.Tok, 0, 0)
				res.Box.Add(b, Indent, 0)
			case tty.Mono:
				res.Box = &tty.Row{}
				res.Box.Add(tok.Tok, 0, 0)
				res.Box.Add(b, 0, 0)
			}

			err.Span = res.Box

			return nil, err
		}

		switch sb := res.Sub.Span().(type) {
		case *tty.Box:
			res.Box = &tty.Box{}
			res.Box.Add(tok.Tok, 0, 0)
			res.Box.Add(sb, Indent, 0)
		case tty.Mono:
			res.Box = &tty.Row{}
			res.Box.Add(tok.Tok, 0, 0)
			res.Box.Add(sb, 0, 0)
		}

		if _, err := p.mustIn(lex.RP, res.Box, 0, 0); err != nil {
			err.Span = res.Box
			return nil, err
		}

		return res, nil
	}

	row, col := tok.Tok.Pos.Row, tok.Tok.Pos.Col
	tok.Tok.Hint().Text = fmt.Sprintf("%d:%d - expression expected here", row, col)
	tok.Tok.Hint().Attr.Color.Add(color.FgRed)

	return nil, &typ.Error{
		Span: tok.Tok,
		Full: "unexpected control sequence",
		Help: "verify language structuring rules",
	}
}

func (p *parser) parseStructFieldExpr() (*StructFieldExpr, *typ.Error) {
	res := &StructFieldExpr{}
	top := &tty.Row{}

	mem, err := p.mustIn(lex.IDEN, top, 0, 0)

	if err != nil {
		err.Span = top
		return nil, err
	}

	res.Mem = mem.Tok

	if _, err := p.mustIn(lex.COL, top, 0, 0); err != nil {
		err.Span = top
		return nil, err
	}

	res.Val, err = p.parseExpr0()

	if err != nil {
		switch b := err.Span.(type) {
		case *tty.Box:
			res.Box = &tty.Box{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(b, Indent, 0)
		case tty.Mono:
			res.Box = &tty.Row{}
			res.Box.Add(top, 0, 0)
			res.Box.Add(b, 1, 0)
		}

		err.Span = res.Box

		return nil, err
	}

	switch vb := res.Val.Span().(type) {
	case *tty.Box:
		res.Box = &tty.Box{}
		res.Box.Add(top, 0, 0)
		res.Box.Add(vb, Indent, 0)
	case tty.Mono:
		res.Box = &tty.Row{}
		res.Box.Add(top, 0, 0)
		res.Box.Add(vb, 1, 0)
	}

	return res, nil
}

func (s *ReturnStmt) Span() tty.Span { return s.Box }
func (s *VarStmt) Span() tty.Span    { return s.Box }
func (s *LetStmt) Span() tty.Span    { return s.Box }
func (s *AssignStmt) Span() tty.Span { return s.Box }
func (s *BlockStmt) Span() tty.Span  { return s.Box }
func (s *LoopStmt) Span() tty.Span   { return s.Box }
func (s *CondStmt) Span() tty.Span   { return s.Box }
func (s *CallStmt) Span() tty.Span   { return s.Box }

func (s *ReturnStmt) Accept(v Visitor) *typ.Error { return v.VisitStmt(s) }
func (s *VarStmt) Accept(v Visitor) *typ.Error    { return v.VisitStmt(s) }
func (s *LetStmt) Accept(v Visitor) *typ.Error    { return v.VisitStmt(s) }
func (s *AssignStmt) Accept(v Visitor) *typ.Error { return v.VisitStmt(s) }
func (s *BlockStmt) Accept(v Visitor) *typ.Error  { return v.VisitStmt(s) }
func (s *LoopStmt) Accept(v Visitor) *typ.Error   { return v.VisitStmt(s) }
func (s *CondStmt) Accept(v Visitor) *typ.Error   { return v.VisitStmt(s) }
func (s *CallStmt) Accept(v Visitor) *typ.Error   { return v.VisitStmt(s) }

func (*ReturnStmt) stmt() {}
func (*VarStmt) stmt()    {}
func (*LetStmt) stmt()    {}
func (*AssignStmt) stmt() {}
func (*BlockStmt) stmt()  {}
func (*LoopStmt) stmt()   {}
func (*CondStmt) stmt()   {}
func (*CallStmt) stmt()   {}

func (e *BasicExpr) Span() tty.Span  { return e.Lex.Tok }
func (e *StructExpr) Span() tty.Span { return e.Box }
func (e *IdenExpr) Span() tty.Span   { return e.Tok }
func (e *DotExpr) Span() tty.Span    { return e.Box }
func (e *InfExpr) Span() tty.Span    { return e.Box }
func (e *PfxExpr) Span() tty.Span    { return e.Box }
func (e *CallExpr) Span() tty.Span   { return e.Box }
func (e *ToI64Expr) Span() tty.Span  { return e.Box }
func (e *ToU64Expr) Span() tty.Span  { return e.Box }
func (e *ToF64Expr) Span() tty.Span  { return e.Box }
func (e *GroupExpr) Span() tty.Span  { return e.Box }

func (e *BasicExpr) Accept(v Visitor) *typ.Error  { return v.VisitExpr(e) }
func (e *StructExpr) Accept(v Visitor) *typ.Error { return v.VisitExpr(e) }
func (e *IdenExpr) Accept(v Visitor) *typ.Error   { return v.VisitExpr(e) }
func (e *DotExpr) Accept(v Visitor) *typ.Error    { return v.VisitExpr(e) }
func (e *InfExpr) Accept(v Visitor) *typ.Error    { return v.VisitExpr(e) }
func (e *PfxExpr) Accept(v Visitor) *typ.Error    { return v.VisitExpr(e) }
func (e *CallExpr) Accept(v Visitor) *typ.Error   { return v.VisitExpr(e) }
func (e *ToI64Expr) Accept(v Visitor) *typ.Error  { return v.VisitExpr(e) }
func (e *ToU64Expr) Accept(v Visitor) *typ.Error  { return v.VisitExpr(e) }
func (e *ToF64Expr) Accept(v Visitor) *typ.Error  { return v.VisitExpr(e) }
func (e *GroupExpr) Accept(v Visitor) *typ.Error  { return v.VisitExpr(e) }

func (e *BasicExpr) Object() *typ.Object  { return e.Obj }
func (e *StructExpr) Object() *typ.Object { return e.Obj }
func (e *IdenExpr) Object() *typ.Object   { return e.Obj }
func (e *DotExpr) Object() *typ.Object    { return e.Obj }
func (e *InfExpr) Object() *typ.Object    { return e.Res }
func (e *PfxExpr) Object() *typ.Object    { return e.Res }
func (e *CallExpr) Object() *typ.Object   { return e.Ret }
func (e *ToI64Expr) Object() *typ.Object  { return e.Res }
func (e *ToU64Expr) Object() *typ.Object  { return e.Res }
func (e *ToF64Expr) Object() *typ.Object  { return e.Res }
func (e *GroupExpr) Object() *typ.Object  { return e.Sub.Object() }

func (*BasicExpr) expr()  {}
func (*StructExpr) expr() {}
func (*IdenExpr) expr()   {}
func (*DotExpr) expr()    {}
func (*InfExpr) expr()    {}
func (*PfxExpr) expr()    {}
func (*CallExpr) expr()   {}
func (*ToI64Expr) expr()  {}
func (*ToU64Expr) expr()  {}
func (*ToF64Expr) expr()  {}
func (*GroupExpr) expr()  {}

func (d *UseDecl) Span() tty.Span    { return d.Box }
func (d *VarDecl) Span() tty.Span    { return d.Box }
func (d *FuncDecl) Span() tty.Span   { return d.Box }
func (d *StructDecl) Span() tty.Span { return d.Box }
func (d *EnumDecl) Span() tty.Span   { return d.Box }

func (d *UseDecl) Accept(v Visitor) *typ.Error    { return v.VisitDecl(d) }
func (d *VarDecl) Accept(v Visitor) *typ.Error    { return v.VisitDecl(d) }
func (d *FuncDecl) Accept(v Visitor) *typ.Error   { return v.VisitDecl(d) }
func (d *StructDecl) Accept(v Visitor) *typ.Error { return v.VisitDecl(d) }
func (d *EnumDecl) Accept(v Visitor) *typ.Error   { return v.VisitDecl(d) }

func (*UseDecl) decl()    {}
func (*VarDecl) decl()    {}
func (*FuncDecl) decl()   {}
func (*StructDecl) decl() {}
func (*EnumDecl) decl()   {}

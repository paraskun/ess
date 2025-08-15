package lex

type Type byte

const (
	EOF Type = iota
	ERR

	// Literals

	IDEN // identifier

	II64 // 102
	IU64 // 102u
	IF64 // 102.0
	ISTR // "hi!"

	// Punctuation

	LP  // (
	RP  // )
	LB  // {
	RB  // }
	LSB // [
	RSB // ]
	COL // :
	SEM // ;
	COM // ,
	DOT // .

	// Binary operators

	ADD  // +
	SUB  // -
	MUL  // *
	DIV  // /
	POW  // **
	SHL  // <<
	SHR  // >>
	MOD  // %
	BAND // &
	BOR  // |
	BXOR // ^

	// Prefix operators

	BNEG // ~
	UNEG // -
	LNEG // !

	// Conditional operators

	LT  // <
	LE  // <=
	GT  // >
	GE  // >=
	EEQ // ==
	NE  // !=
	EQ  // =

	// Logical operators

	LAND // &&
	LOR  // ||

	// Keywords

	USE
	VAR
	LET
	I64
	U64
	F64
	STR
	BOOL
	TYPE
	ENUM
	FUNC
	FOR
	IF
	ELSE
	RET
	BREAK
	TRUE
	FALSE

	// Other

	DOG
)

func (t Type) String() string {
	return str[t]
}

func AsKeyword(lit string) (Type, bool) {
	tt, ok := kwd[lit]
	return tt, ok
}

var kwd = map[string]Type{
	"use":    USE,
	"var":    VAR,
	"let":    LET,
	"i64":    I64,
	"u64":    U64,
	"f64":    F64,
	"str":    STR,
	"bool":   BOOL,
	"type":   TYPE,
	"enum":   ENUM,
	"func":   FUNC,
	"for":    FOR,
	"if":     IF,
	"else":   ELSE,
	"return": RET,
	"break":  BREAK,
	"true":   TRUE,
	"false":  FALSE,
}

var str = map[Type]string{
	IDEN: "identifier",
	II64: "signed numeric literal",
	IU64: "unsigned numeric literal",
	IF64: "floating-point literal",
	ISTR: "string literal",

	LP:  "(",
	RP:  ")",
	LB:  "{",
	RB:  "}",
	LSB: "[",
	RSB: "]",
	COL: ":",
	SEM: ";",
	COM: ",",
	DOT: ".",

	ADD:  "+",
	SUB:  "-",
	MUL:  "*",
	DIV:  "/",
	POW:  "**",
	SHL:  "<<",
	SHR:  ">>",
	MOD:  "%",
	BAND: "&",
	BOR:  "|",
	BXOR: "^",

	BNEG: "~",
	UNEG: "-",
	LNEG: "!",

	LT:  "<",
	LE:  "<=",
	GT:  ">",
	GE:  ">=",
	EEQ: "==",
	NE:  "!=",
	EQ:  "=",

	LAND: "&&",
	LOR:  "||",

	USE:   "use",
	VAR:   "var",
	LET:   "let",
	I64:   "i64",
	U64:   "u64",
	F64:   "f64",
	STR:   "str",
	BOOL:  "bool",
	TYPE:  "type",
	ENUM:  "enum",
	FUNC:  "func",
	FOR:   "for",
	IF:    "if",
	ELSE:  "else",
	RET:   "ret",
	BREAK: "break",
	TRUE:  "true",
	FALSE: "false",

	DOG: "@",
}

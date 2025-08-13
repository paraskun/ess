package lex

type Type byte

const (
	EOF Type = iota

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
	INI // :=

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
)

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

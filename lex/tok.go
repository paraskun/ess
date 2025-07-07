package lex

type TokenType byte

const (
	EOF TokenType = iota

	// Literals

	IDF // identifier

	II64 // 102
	IU64 // 102u
	IF64 // 102.0
	ISTR // "hi!"
	IMEM // enum member (uppercase identifier)

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

type Token struct {
	TokenType

	Row int
	Col int
	Lit string
}

var kwd = map[string]TokenType{
	"use":    USE,
	"var":    VAR,
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

func AsKeyword(lit string) (TokenType, bool) {
	tt, ok := kwd[lit]
	return tt, ok
}

var dbg = map[TokenType]string{
	EOF:   "end of file",
	IDF:   "identifier",
	II64:  "i64 literal",
	IU64:  "u64 literal",
	IF64:  "f64 literal",
	LP:    "\"(\"",
	RP:    "\")\"",
	LB:    "\"{\"",
	RB:    "\"}\"",
	LSB:   "\"[\"",
	RSB:   "\"]\"",
	COL:   "\":\"",
	SEM:   "\";\"",
	COM:   "\",\"",
	DOT:   "\".\"",
	ADD:   "\"+\"",
	SUB:   "\"-\"",
	MUL:   "\"*\"",
	DIV:   "\"/\"",
	POW:   "\"**\"",
	SHL:   "\"<<\"",
	SHR:   "\">>\"",
	MOD:   "\"%\"",
	BAND:  "\"&\"",
	BOR:   "\"|\"",
	BXOR:  "\"^\"",
	BNEG:  "\"~\"",
	UNEG:  "\"-\"",
	LNEG:  "\"!\"",
	LT:    "\"<\"",
	LE:    "\"<=\"",
	GT:    "\">\"",
	GE:    "\">=\"",
	EEQ:   "\"==\"",
	NE:    "\"!=\"",
	EQ:    "\"=\"",
	LAND:  "\"&&\"",
	LOR:   "\"||\"",
	VAR:   "\"var\"",
	I64:   "\"i64\"",
	U64:   "\"u64\"",
	F64:   "\"f64\"",
	BOOL:  "\"bool\"",
	TYPE:  "\"type\"",
	ENUM:  "\"enum\"",
	FUNC:  "\"func\"",
	FOR:   "\"for\"",
	IF:    "\"if\"",
	ELSE:  "\"else\"",
	RET:   "\"return\"",
	BREAK: "\"break\"",
	TRUE:  "\"true\"",
	FALSE: "\"false\"",
}

func (t TokenType) String() string {
	return dbg[t]
}

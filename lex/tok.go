package lex

type TokenType byte

const (
	EOF TokenType = iota

	// Literals

	IDF  // identifier
	II64 // 102
	IU64 // 102u
	IF64 // 102.0

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

	VAR
	I64
	U64
	F64
	BOOL
	TYPE
	FUNC
	FOR
	IF
	ELSE
	RET
	BREAK
	TRUE
	FALSE
	PREV
)

type Token struct {
	TokenType

	Row int
	Col int
	Lit string
}

var kwd = map[string]TokenType{
	"var":    VAR,
	"i64":    I64,
	"u64":    U64,
	"f64":    F64,
	"bool":   BOOL,
	"type":   TYPE,
	"func":   FUNC,
	"for":    FOR,
	"if":     IF,
	"else":   ELSE,
	"return": RET,
	"break":  BREAK,
	"true":   TRUE,
	"false":  FALSE,
	"prev":   PREV,
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
	FUNC:  "\"func\"",
	FOR:   "\"for\"",
	IF:    "\"if\"",
	ELSE:  "\"else\"",
	RET:   "\"return\"",
	BREAK: "\"break\"",
	TRUE:  "\"true\"",
	FALSE: "\"false\"",
	PREV:  "\"prev\"",
}

func (t TokenType) String() string {
	return dbg[t]
}

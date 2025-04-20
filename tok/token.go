package tok

type TokenType int

const (
	ERR TokenType = iota
	EOF

	// Literals

	ID    // identifier
	INT   // 102
	FLT   // 102.23
	STR   // "hi"
	TRUE  // true
	FALSE // false

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

	// Unary operators

	BINV // ~
	UNEG // -

	// Conditional operators

	LT  // <
	LE  // <=
	GT  // >
	GE  // >=
	EQ  // =
	EEQ // ==
	NE  // !=

	// Logical operators

	LAND // &&
	LOR  // ||
	LNEG // !

	// Keywords

	FUNC  // func
	FOR   // for
	IF    // if
	ELSE  // else
	RET   // return
	BREAK // break

	// Predefined functions

	PRINT // print
)

type Token struct {
	TokenType

	Row int
	Col int
	Lit string
	Err error
}

var kwd = map[string]TokenType{
	"func":   FUNC,
	"for":    FOR,
	"if":     IF,
	"else":   ELSE,
	"return": RET,
	"break":  BREAK,
	"print":  PRINT,
}

func AsKeyword(lit string) (TokenType, bool) {
	tt, ok := kwd[lit]
	return tt, ok
}

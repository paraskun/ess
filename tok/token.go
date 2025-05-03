//go:generate stringer -type=TokenType

package tok

type TokenType int

const (
	EOF TokenType = iota

	// Literals

	IDF  // identifier
	LI64 // 102

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

	// Prefix operators

	BNEG // ~
	UNEG // -
	LNEG // !

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

	// Keywords

	I64
	BOOL
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
	"i64":    I64,
	"bool":   BOOL,
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

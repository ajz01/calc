package token

import "strconv"

type Token int

const (
	ILLEGAL Token = iota
	EOF

	literal_beg
	IDENT // Identifier
	FRML  // =
	BOOL  // TRUE|FALSE
	CELL  // Cell reference
	ERR   // #ERROR!
	ERREF // #REF!
	FUNC  // Built-in functions
	RNG
	REF
	SHEET  // Sheet name
	INT    // 12345
	FLOAT  // 123.45
	IMAG   // 123.45i
	STRING // "abc"
	literal_end

	operator_beg
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	EXP // ^

	LAND // And
	LOR  // Or

	EQL // =
	LSS // <
	GTR // >
	NOT // <>

	LEQ // <=
	GEQ // >=

	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :
	operator_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	IDENT:  "IDENT",
	FRML:   "FRML",
	BOOL:   "BOOL",
	CELL:   "CELL",
	ERR:    "ERR",
	ERREF:  "ERREF",
	FUNC:   "FUNC",
	REF:    "REF",
	RNG:    "RNG",
	SHEET:  "SHEET",
	INT:    "INT",
	FLOAT:  "FLOAT",
	IMAG:   "IMAG",
	STRING: "STRING",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	EXP: "^",

	LAND: "AND",
	LOR:  "OR",

	EQL: "=",
	LSS: "<",
	GTR: ">",
	NOT: "<>",

	LEQ: "<=",
	GEQ: ">=",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",
	COMMA:  ",",
	PERIOD: ".",

	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",
}

var ops = map[rune]Token{
	'+': ADD,
	'-': SUB,
	'*': MUL,
	'/': QUO,
	'^': EXP,

	'(': LPAREN,
	'[': LBRACK,
	'{': LBRACE,
	',': COMMA,
	'.': PERIOD,

	')': RPAREN,
	']': RBRACK,
	'}': RBRACE,
	';': SEMICOLON,
	':': COLON,
}

func SingleRune(r rune) (t Token) {
	t = ops[r]
	return
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

const (
	LowestPrec  = 0
	UnaryPrec   = 6
	HighestPrec = 7
)

func (op Token) Precedence() int {
	switch op {
	case LOR:
		return 1
	case LAND:
		return 2
	case EQL, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB:
		return 4
	case MUL, QUO:
		return 5
	}
	return LowestPrec
}

func (tok Token) IsLiteral() bool { return literal_beg < tok && tok < literal_end }

func (tok Token) IsOperator() bool { return operator_beg < tok && tok < operator_end }

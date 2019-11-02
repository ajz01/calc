package scanner

import (
	"fmt"
	"github.com/ajz01/calc/token"
	"sort"
	"unicode"
	"unicode/utf8"
)

type ErrorHandler func(pos token.Position, msg string)

type Scanner struct {
	src        []byte
	err        ErrorHandler
	ch         rune
	offset     int
	rdOffset   int
	lineOffset int
	colOffset  int
	prev       token.Token
}

func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		s.colOffset++
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.colOffset = 0
		}
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal UTF-8 encoding")
			}
		}
		s.rdOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
		}
		s.ch = -1 // eof
	}
}

func (s *Scanner) peek() byte {
	if s.rdOffset < len(s.src) {
		return s.src[s.rdOffset]
	}
	return 0
}

func (s *Scanner) error(offs int, msg string) {
	if s.err != nil {
		s.err(token.Position{Offset: offs}, msg)
	}
}

func (s *Scanner) errorf(offs int, format string, args ...interface{}) {
	s.error(offs, fmt.Sprintf(format, args...))
}

func (s *Scanner) Init(src []byte, err ErrorHandler) {
	s.src = src
	s.err = err
	s.next()
}

func isLetter(ch rune) bool {
	return 'a' <= lower(ch) && lower(ch) <= 'z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return isDecimal(ch) || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

func (s *Scanner) scanIdentifier() (string, bool) {
	ref := false
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	// Col or Row only Ref
	if s.offset-offs == 1 {
		if isLetter(rune(s.src[offs : offs+1][0])) {
			ref = true
		}
	} else if s.offset-offs == 2 {
		if isLetter(rune(s.src[offs : offs+1][0])) && isDigit(rune(s.src[offs+1 : offs+2][0])) {
			ref = true
		}
	}
	return string(s.src[offs:s.offset]), ref
}

func lower(ch rune) rune     { return ('a' - 'A') | ch } // returns lower-case ch if ch is ASCII
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }
func isHex(ch rune) bool     { return '0' <= ch && ch <= '9' || 'a' <= lower(ch) && lower(ch) <= 'f' }

func (s *Scanner) digits(base int, invalid *int) (digsep int) {
	if base <= 10 {
		max := rune('0' + base)
		for isDecimal(s.ch) || s.ch == '_' {
			ds := 1
			if s.ch == '_' {
				ds = 2
			} else if s.ch >= max && *invalid < 0 {
				*invalid = int(s.offset) // record invalid rune offset
			}
			digsep |= ds
			s.next()
		}
	} else {
		for isHex(s.ch) || s.ch == '_' {
			ds := 1
			if s.ch == '_' {
				ds = 2
			}
			digsep |= ds
			s.next()
		}
	}
	return
}

func (s *Scanner) scanNumber() (token.Token, string) {
	offs := s.offset
	tok := token.ILLEGAL

	base := 10
	prefix := rune(0)
	digsep := 0
	invalid := -1

	// integer part
	if s.ch != '.' {
		tok = token.INT
		if s.ch == '0' {
			s.next()
			switch lower(s.ch) {
			case 'x':
				s.next()
				base, prefix = 16, 'x'
			case 'o':
				s.next()
				base, prefix = 8, 'o'
			case 'b':
				s.next()
				base, prefix = 2, 'b'
			default:
				base, prefix = 8, '0'
				digsep = 1 // leading 0
			}
		}
		digsep |= s.digits(base, &invalid)
	}

	// fractional part
	if s.ch == '.' {
		tok = token.FLOAT
		if prefix == 'o' || prefix == 'b' {
			s.error(s.offset, "invalid radix point in "+litname(prefix))
		}
		s.next()
		digsep |= s.digits(base, &invalid)
	}

	if digsep&1 == 0 {
		s.error(s.offset, litname(prefix)+" has no digits")
	}

	// exponent
	if e := lower(s.ch); e == 'e' || e == 'p' {
		switch {
		case e == 'e' && prefix != 0 && prefix != '0':
			s.errorf(s.offset, "%q exponent requires decimal mantissa", s.ch)
		case e == 'p' && prefix != 'x':
			s.errorf(s.offset, "%q exponent requires hexidecimal mantissa", s.ch)
		}
		s.next()
		tok = token.FLOAT
		if s.ch == '+' || s.ch == '-' {
			s.next()
		}
		ds := s.digits(10, nil)
		digsep |= ds
		if ds&1 == 0 {
			s.error(s.offset, "exponent has no digits")
		}
	} else if prefix == 'x' && tok == token.FLOAT {
		s.error(s.offset, "hexadecimal mantissa requires a 'p' exponent")
	}

	// suffix i
	if s.ch == 'i' {
		tok = token.IMAG
		s.next()
	}

	lit := string(s.src[offs:s.offset])
	if tok == token.INT && invalid >= 0 {
		s.errorf(invalid, "invalid digit %q in %s", lit[invalid-offs], litname(prefix))
	}
	if digsep&2 != 0 {
		if i := invalidSep(lit); i >= 0 {
			s.error(offs+i, "'_' must seperate successive digits")
		}
	}

	return tok, lit
}

func litname(prefix rune) string {
	switch prefix {
	case 'x':
		return "hexadecimal literal"
	case 'o':
		return "octal literal"
	case 'b':
		return "binary literal"
	}
	return "decimal literal"
}

func invalidSep(x string) int {
	x1 := ' '
	d := '.'
	i := 0

	if len(x) >= 2 && x[0] == '0' {
		x1 = lower(rune(x[1]))
		if x1 == 'x' || x1 == 'o' || x1 == 'b' {
			d = '0'
			i = 2
		}
	}

	// mantissa and exponent
	for ; i < len(x); i++ {
		p := d // previous digit
		d = rune(x[i])
		switch {
		case d == '_':
			if p != '0' {
				return i
			}
		case isDecimal(d) || x1 == 'x' && isHex(d):
			d = '0'
		default:
			if p == '_' {
				return i - 1
			}
			d = '.'
		}
	}
	if d == '_' {
		return len(x) - 1
	}

	return -1
}

func (s *Scanner) scanString() string {
	// '"' opening already consumed
	offs := s.offset - 1

	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {
			s.error(offs, "string literal not terminated")
			break
		}
		s.next()
		if ch == '"' {
			break
		}
		if ch == '\\' {
			// skip for now
			//s.scanEscape('"')
		}
	}

	return string(s.src[offs:s.offset])
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

func (s *Scanner) switch2(tok0, tok1 token.Token) token.Token {
	if s.ch == '=' {
		s.next()
		return tok1
	}
	return tok0
}

func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {
	//scanAgain:
	s.skipWhitespace()

	// current token start
	pos = token.Pos(s.offset + 1)

	switch ch := s.ch; {
	case isLetter(ch):
		var ref bool
		lit, ref = s.scanIdentifier()
		if s.ch == '(' {
			tok = token.IDENT//.FUNC
		} else if ref {
			if s.ch == ':' {
				s.next()
				lit2, ref2 := s.scanIdentifier()
				if ref2 {
					tok = token.RNG
					lit = lit + ":" + lit2
				}
			} else {
				tok = token.REF
			}
		} else {
			tok = token.IDENT
			fmt.Printf("prev %v\n", s.prev)
		}
	case isDecimal(ch) || ch == '.' && isDecimal(rune(s.peek())):
		tok, lit = s.scanNumber()
	default:
		s.next()
		switch ch {
		case '=':
			if s.colOffset == 0 {
				tok = token.FRML
			} else {
				tok = token.EQL
			}
		case '"':
			tok = token.STRING
			lit = s.scanString()
		case '<':
			tok = s.switch2(token.LSS, token.LEQ)
		case '>':
			tok = s.switch2(token.GTR, token.GEQ)
		default:
			tok = token.SingleRune(ch)
		}
	}

	s.prev = tok

	return
}

type Error struct {
	Pos token.Position
	Msg string
}

func (e Error) Error() string {
	return e.Pos.String() + e.Msg
}

type ErrorList []*Error

func (p *ErrorList) Add(pos token.Position, msg string) {
	*p = append(*p, &Error{pos, msg})
}

func (p ErrorList) Len() int      { return len(p) }
func (p ErrorList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p ErrorList) Less(i, j int) bool {
	e := &p[i].Pos
	f := &p[j].Pos
	return e.Offset < f.Offset
}

func (p ErrorList) Sort() {
	sort.Sort(p)
}

func (p ErrorList) Error() string {
	switch len(p) {
	case 0:
		return "no errors"
	case 1:
		return p[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", p[0], len(p)-1)
}

func (p ErrorList) Err() error {
	if len(p) == 0 {
		return nil
	}
	return p
}

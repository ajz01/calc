package parser

import (
	"fmt"
	"github.com/ajz01/calc/ast"
	"github.com/ajz01/calc/scanner"
	"github.com/ajz01/calc/token"
)

type parser struct {
	errors  scanner.ErrorList
	scanner scanner.Scanner

	trace  bool
	indent int

	// Next token
	pos token.Pos   // token position
	tok token.Token // one token look-ahead
	lit string      // token literal

	exprLev int
	inRhs   bool

	targetStack [][]*ast.Ident
}

func (p *parser) init(src []byte) {
	eh := func(pos token.Position, msg string) { p.errors.Add(pos, msg) }
	p.scanner.Init(src, eh)
	p.trace = Trace
	p.next()
}

func (p *parser) printTrace(a ...interface{}) {
	const dots = ".............................."
	const n = len(dots)
	i := 2 * p.indent
	for i > n {
		fmt.Print(dots)
		i -= n
	}
	fmt.Print(dots[0:i])
	fmt.Println(a...)
}

func trace(p *parser, msg string) *parser {
	p.printTrace(msg, "(")
	p.indent++
	return p
}

func un(p *parser) {
	p.indent--
	p.printTrace(")")
}

func (p *parser) next0() {
	if p.trace && p.pos.IsValid() {
		s := p.tok.String()
		switch {
		case p.tok.IsLiteral():
			p.printTrace(s, p.lit)
		case p.tok.IsOperator():
			p.printTrace("\"" + s + "\"")
		default:
			p.printTrace(s)
		}
	}

	p.pos, p.tok, p.lit = p.scanner.Scan()
}

func (p *parser) next() {
	p.next0()
}

func (p *parser) error(pos token.Pos, msg string) {
	epos := token.Position{Offset: int(pos)}
	p.errors.Add(epos, msg)
}

func (p *parser) parseIdent() *ast.Ident {
	pos := p.pos
	name := "_"
	if p.tok == token.IDENT {
		name = p.lit
		p.next()
	}
	return &ast.Ident{NamePos: pos, Name: name}
}

func (p *parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}
	p.next()
	return pos
}

func (p *parser) expectClosing(tok token.Token, context string) token.Pos {
	if p.tok != tok {
		p.error(p.pos, "missing ','"+context)
		p.next()
	}
	return p.expect(tok)
}

func (p *parser) errorExpected(pos token.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		switch {
		case p.tok == token.SEMICOLON && p.lit == "\n":
			msg += ", found newline"
		case p.tok.IsLiteral():
			msg += ", found " + p.lit
		default:
			msg += ", found '" + p.tok.String() + "'"
		}
	}
	p.error(pos, msg)
}

func (p *parser) parseIdentList() (list []*ast.Ident) {
	if p.trace {
		defer un(trace(p, "IdentList"))
	}

	list = append(list, p.parseIdent())
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.parseIdent())
	}

	return
}

func (p *parser) parseExprList(lhs bool) (list []ast.Expr) {
	if p.trace {
		defer un(trace(p, "ExpressionList"))
	}

	list = append(list, p.checkExpr(p.parseExpr(lhs)))
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.checkExpr(p.parseExpr(lhs)))
	}

	return
}

func (p *parser) parseType() ast.Expr {
	if p.trace {
		defer un(trace(p, "Type"))
	}

	typ := p.tryIdentOrType()//p.tryType()

	if typ == nil {
		pos := p.pos
		p.errorExpected(pos, "type")
		return &ast.BadExpr{From: pos, To: p.pos}
	}

	return typ
}

func (p *parser) parseParameterList(isParam bool) (params []*ast.Field) {
	if p.trace {
		defer un(trace(p, "ParameterList"))
	}

	var list []ast.Expr
	for {
		list = append(list, p.parseExpr(true))
		if p.tok != token.COMMA {
			break
		}
		p.next()
		if p.tok == token.RPAREN {
			break
		}
	}

	params = make([]*ast.Field, len(list))
	for i, typ := range list {
		params[i] = &ast.Field{Type: typ}
	}
	return
}

func (p *parser) parseParameters(isParam bool) *ast.FieldList {
	if p.trace {
		defer un(trace(p, "Parameters"))
	}

	var params []*ast.Field
	lparen := p.expect(token.LPAREN)
	if p.tok != token.RPAREN {
		params = p.parseParameterList(isParam)
	}
	rparen := p.expect(token.RPAREN)

	return &ast.FieldList{Opening: lparen, List: params, Closing: rparen}
}

func (p *parser) parseSignature() (params *ast.FieldList) {
	if p.trace {
		defer un(trace(p, "Signature"))
	}

	params = p.parseParameters(true)

	return
}

/*func (p *parser) parseFuncType() *ast.FuncType {
	if p.trace {
		defer un(trace(p, "FuncType"))
	}

	pos := p.expect(token.FUNC)
	params := p.parseSignature()

	return &ast.FuncType{Func: pos, Params: params}
}*/

func (p *parser) tryIdentOrType() ast.Expr {
	switch p.tok {
	case token.IDENT, token.REF, token.RNG:
		return p.parseIdent()
	case token.LPAREN:
		lparen := p.pos
		p.next()
		typ := p.parseType()
		rparen := p.expect(token.RPAREN)
		return &ast.ParenExpr{Lparen: lparen, X: typ, Rparen: rparen}
	}

	return nil
}

func (p *parser) parseOperand(lhs bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "Operand"))
	}

	switch p.tok {
	case token.IDENT:
		x := p.parseIdent()
		return x

	case token.INT, token.FLOAT, token.IMAG, token.STRING, token.REF, token.RNG:
		x := &ast.BasicLit{ValuePos: p.pos, Kind: p.tok, Value: p.lit}
		p.next()
		return x

	case token.LPAREN:
		lparen := p.pos
		p.next()
		x := p.parseRhs()
		rparen := p.expect(token.RPAREN)
		return &ast.ParenExpr{Lparen: lparen, X: x, Rparen: rparen}

	//case token.FUNC:
		//return p.parseFuncType()
	}

	// we have an error
	pos := p.pos
	p.errorExpected(pos, "operand")
	//p.advance(stmtStart)
	return &ast.BadExpr{From: pos, To: p.pos}
}

func (p *parser) checkExpr(x ast.Expr) ast.Expr {
	switch unparen(x).(type) {
	case *ast.BadExpr:
	case *ast.Ident:
	case *ast.BasicLit:
	case *ast.ParenExpr:
		panic("unreachable")
	case *ast.UnaryExpr:
	case *ast.BinaryExpr:
	default:
		p.errorExpected(x.Pos(), "expression")
		x = &ast.BadExpr{From: x.Pos(), To: x.End()}
	}
	return x
}

func unparen(x ast.Expr) ast.Expr {
	if p, isParen := x.(*ast.ParenExpr); isParen {
		x = unparen(p.X)
	}
	return x
}

func (p *parser) parseCall(fun ast.Expr) *ast.CallExpr {
	if p.trace {
		defer un(trace(p, "Call"))
	}

	lparen := p.expect(token.LPAREN)
	var list []ast.Expr
	for p.tok != token.RPAREN && p.tok != token.EOF {
		list = append(list, p.parseRhs())
		if !p.atComma("argument list", token.RPAREN) {
			break
		}
		p.next()
	}
	rparen := p.expectClosing(token.RPAREN, "argument list")

	return &ast.CallExpr{Fun: fun, Lparen: lparen, Args: list, Rparen: rparen}
}

func (p *parser) atComma(context string, follow token.Token) bool {
	if p.tok == token.COMMA {
		return true
	}
	if p.tok != follow {
		msg := "missing ','"
		p.error(p.pos, msg+" in "+context)
		return true
	}
	return false
}

func (p *parser) parsePrimaryExpr(lhs bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "PrimaryExpr"))
	}

	x := p.parseOperand(lhs)
L:
	for {
		switch p.tok {
		case token.LPAREN:
			x = p.parseCall(p.checkExpr(x))
		default:
			break L
		}
		lhs = false
	}

	return x
}

func (p *parser) parseUnaryExpr(lhs bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "UnaryExpr"))
	}

	switch p.tok {
	case token.ADD, token.SUB, token.NOT, token.MUL:
		pos, op := p.pos, p.tok
		p.next()
		x := p.parseUnaryExpr(false)
		return &ast.UnaryExpr{OpPos: pos, Op: op, X: p.checkExpr(x)}

	}

	return p.parsePrimaryExpr(lhs)
}

func (p *parser) tokPrec() (token.Token, int) {
	tok := p.tok
	return tok, tok.Precedence()
}

func (p *parser) parseBinaryExpr(lhs bool, prec1 int) ast.Expr {
	if p.trace {
		defer un(trace(p, "BinaryExpr"))
	}

	x := p.parseUnaryExpr(lhs)
	for {
		op, oprec := p.tokPrec()
		if oprec < prec1 {
			return x
		}
		pos := p.expect(op)
		if lhs {
			lhs = false
		}
		y := p.parseBinaryExpr(false, oprec+1)
		x = &ast.BinaryExpr{X: p.checkExpr(x), OpPos: pos, Op: op, Y: p.checkExpr(y)}
	}
}

func (p *parser) parseExpr(lhs bool) ast.Expr {
	if p.trace {
		defer un(trace(p, "Expression"))
	}

	return p.parseBinaryExpr(lhs, token.LowestPrec+1)
}

func (p *parser) parseRhs() ast.Expr {
	old := p.inRhs
	p.inRhs = true
	x := p.checkExpr(p.parseExpr(false))
	p.inRhs = old
	return x
}

func (p *parser) parseBytes() ast.Expr {
	if p.trace {
		defer un(trace(p, "Bytes"))
	}

	return p.parseExpr(true)
}

const Trace = true

func ParseBytes(src []byte) (f ast.Expr, err error) {
	var p parser
	defer func() {
		if f == nil {
			var e ast.Expr
			f = e
		}
		p.errors.Sort()
		err = p.errors.Err()
	}()

	p.init(src)
	f = p.parseBytes()

	return
}

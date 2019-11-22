package ast

import (
	"github.com/ajz01/calc/token"
)

type Node interface {
	Pos() token.Pos // position of first character belonging to the node
	End() token.Pos // position of first character immediately after the node
}

type Expr interface {
	Node
	exprNode()
}

type Field struct {
	Names []*Ident
	Type  Expr
	Tag   *BasicLit
}

func (f *Field) Pos() token.Pos {
	if len(f.Names) > 0 {
		return f.Names[0].Pos()
	}
	return f.Type.Pos()
}

func (f *Field) End() token.Pos {
	if f.Tag != nil {
		return f.Tag.End()
	}
	return f.Type.End()
}

type FieldList struct {
	Opening token.Pos
	List    []*Field
	Closing token.Pos
}

func (f *FieldList) Pos() token.Pos {
	if f.Opening.IsValid() {
		return f.Opening
	}
	if len(f.List) > 0 {
		return f.List[0].Pos()
	}
	return token.NoPos
}

func (f *FieldList) End() token.Pos {
	if f.Closing.IsValid() {
		return f.Closing + 1
	}
	if n := len(f.List); n > 0 {
		return f.List[n-1].End()
	}
	return token.NoPos
}

func (f *FieldList) NumFields() int {
	n := 0
	if f != nil {
		for _, g := range f.List {
			m := len(g.Names)
			if m == 0 {
				m = 1
			}
			n += m
		}
	}
	return n
}

type (
	BadExpr struct {
		From, To token.Pos
	}

	Ident struct {
		NamePos token.Pos
		Name    string
	}

	BasicLit struct {
		ValuePos token.Pos
		Kind     token.Token
		Value    string
	}

	ParenExpr struct {
		Lparen token.Pos
		X      Expr
		Rparen token.Pos
	}

	CallExpr struct {
		Fun    Expr
		Lparen token.Pos
		Args   []Expr
		Rparen token.Pos
	}

	UnaryExpr struct {
		OpPos token.Pos
		Op    token.Token
		X     Expr
	}

	BinaryExpr struct {
		X     Expr
		OpPos token.Pos
		Op    token.Token
		Y     Expr
	}

	/*FuncType struct {
		Func   token.Pos
		Params *FieldList
	}*/
)

func (x *BadExpr) Pos() token.Pos    { return x.From }
func (x *Ident) Pos() token.Pos      { return x.NamePos }
func (x *BasicLit) Pos() token.Pos   { return x.ValuePos }
func (x *ParenExpr) Pos() token.Pos  { return x.Lparen }
func (x *CallExpr) Pos() token.Pos   { return x.Fun.Pos() }
func (x *UnaryExpr) Pos() token.Pos  { return x.OpPos }
func (x *BinaryExpr) Pos() token.Pos { return x.X.Pos() }
/*func (x *FuncType) Pos() token.Pos {
	if x.Func.IsValid() || x.Params == nil {
		return x.Func
	}
	return x.Params.Pos()
}*/

func (x *BadExpr) End() token.Pos    { return x.To }
func (x *Ident) End() token.Pos      { return token.Pos(int(x.NamePos) + len(x.Name)) }
func (x *BasicLit) End() token.Pos   { return token.Pos(int(x.ValuePos) + len(x.Value)) }
func (x *ParenExpr) End() token.Pos  { return x.Rparen + 1 }
func (x *CallExpr) End() token.Pos   { return x.Rparen + 1 }
func (x *UnaryExpr) End() token.Pos  { return x.X.End() }
func (x *BinaryExpr) End() token.Pos { return x.Y.End() }
//func (x *FuncType) End() token.Pos   { return x.Params.End() }

func (*BadExpr) exprNode()    {}
func (*Ident) exprNode()      {}
func (*BasicLit) exprNode()   {}
func (*ParenExpr) exprNode()  {}
func (*CallExpr) exprNode()   {}
func (*UnaryExpr) exprNode()  {}
func (*BinaryExpr) exprNode() {}
//func (*FuncType) exprNode()   {}

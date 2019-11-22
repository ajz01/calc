package ast

import "fmt"

type Visitor interface {
	Visit(node Node) (w Visitor)
}

func walkIdentList(v Visitor, list []*Ident) {
	for _, x := range list {
		Walk(v, x)
	}
}

func walkExprList(v Visitor, list []Expr) {
	for _, x := range list {
		Walk(v, x)
	}
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *Field:
		walkIdentList(v, n.Names)
		Walk(v, n.Type)

	case *FieldList:
		for _, f := range n.List {
			Walk(v, f)
		}

	case *BadExpr, *Ident, *BasicLit:

	case *ParenExpr:
		Walk(v, n.X)

	case *CallExpr:
		Walk(v, n.Fun)
		walkExprList(v, n.Args)

	case *UnaryExpr:
		Walk(v, n.X)

	case *BinaryExpr:
		Walk(v, n.X)
		Walk(v, n.Y)

	/*case *FuncType:
		if n.Params != nil {
			Walk(v, n.Params)
		}*/

	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}

	v.Visit(nil)
}

type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f(node) {
		return f
	}
	return nil
}

func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}

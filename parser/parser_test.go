package parser

import (
	"github.com/ajz01/calc/ast"
	"testing"
)

func parse(src string) (ast.Expr, error) {
	return ParseBytes([]byte(src))
}

func TestParseBinary(t *testing.T) {
	src := "1+2"
	e, err := parse(src)
	if err != nil {
		t.Errorf("ParseExpr(%q) %v", src, err)
	}
	if _, ok := e.(*ast.BinaryExpr); !ok {
		t.Errorf("ParseExpr(%q): got %T, want *ast.BinaryExpr", src, e)
	}
}

func TestParseFunc(t *testing.T) {
	src := "FUNC(1+2+3)"
	e, err := parse(src)
	if err != nil {
		t.Errorf("ParseExpr(%q) %v", src, err)
	}
	if _, ok := e.(*ast.FuncType); !ok {
		t.Errorf("ParseExpr(%q): got %T, want *ast.FuncType", src, e)
	}
}

func TestParseRef(t *testing.T) {
	src := "A1"
	e, err := parse(src)
	if err != nil {
		t.Errorf("ParseExpr(%q) %v", src, err)
	}
	if n, ok := e.(*ast.BasicLit); !ok {
		t.Errorf("ParseExpr(%q): got %T, want *ast.BasicLit", src, e)
	} else {
		if n.Value != src {
			t.Errorf("ParseExpr(%q): unexpected reference value %s", src, n.Value)
		}
	}
}

func TestParseRange(t *testing.T) {
	src := "A1:D3"
	e, err := parse(src)
	if err != nil {
		t.Errorf("ParseExpr(%q) %v", src, err)
	}
	if n, ok := e.(*ast.BasicLit); !ok {
		t.Errorf("ParseExpr(%q): got %T, want *ast.BasicLit", src, e)
	} else {
		if n.Value != src {
			t.Errorf("ParseExpr(%q): unexpected range value %s", src, n.Value)
		}
	}
}

func TestParseFuncParamRange(t *testing.T) {
	src := "FUNC(A1:D3)"
	e, err := parse(src)
	if err != nil {
		t.Errorf("ParseExpr(%q) %v", src, err)
	}
	if n, ok := e.(*ast.FuncType); !ok {
		t.Errorf("ParseExpr(%q): got %T, want *ast.FuncType", src, e)
	} else {
		if r, ok := n.Params.List[0].Type.(*ast.BasicLit); !ok {
			t.Errorf("ParseExpr(%q): unexpected param type %T", src, n.Params.List[0].Type)
		} else {
			if r.Value != "A1:D3" {
				t.Errorf("ParseExpr(%q): unexpected range value: %s", src, r.Value)
			}
		}
	}
}

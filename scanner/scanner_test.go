package scanner

import (
	"fmt"
	"github.com/ajz01/calc/token"
	"testing"
)

func setupScanner(formula string) Scanner {
	var s Scanner
	src := []byte(formula)
	err := func(pos token.Position, msg string) {
		fmt.Printf("%d %s\n", pos, msg)
	}
	s.Init(src, err)
	return s
}

func TestScanInt(t *testing.T) {
	s := setupScanner("123")
	_, tok, lit := s.Scan()
	if tok != token.INT || lit != "123" {
		t.Errorf("Scan Integer = %q %q want INT 123", tok, lit)
	}
}

func TestScanFloat(t *testing.T) {
	s := setupScanner("123.45678")
	_, tok, lit := s.Scan()
	if tok != token.FLOAT || lit != "123.45678" {
		t.Errorf("Scan Float = %q %q want FLOAT 123.45678", tok, lit)
	}
}

func TestScanIdent(t *testing.T) {
	s := setupScanner("FUNC")
	_, tok, lit := s.Scan()
	if tok != token.IDENT || lit != "FUNC" {
		t.Errorf("Scan Func = %q %q want FUNC FUNC", tok, lit)
	}
}

func TestScanRange(t *testing.T) {
	s := setupScanner("A1:D3")
	_, tok, lit := s.Scan()
	if tok != token.RNG || lit != "A1:D3" {
		t.Errorf("Scan Range = %q %q want RNG A1:D3", tok, lit)
	}
}
func TestScanOps(t *testing.T) {
	sym := "+-*/^"
	toks := []token.Token{token.ADD, token.SUB, token.MUL, token.QUO, token.EXP}
	s := setupScanner(sym)
	for i, r := range sym {
		_, tok, lit := s.Scan()
		if tok != toks[i] || lit != "" {
			t.Errorf("Scan Ops = %q %q want %q ''", tok, lit, r)
		}
	}
}

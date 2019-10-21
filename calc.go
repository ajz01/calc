package main

import (
	"bufio"
	"fmt"
	"github.com/ajz01/calc/ast"
	"github.com/ajz01/calc/parser"
	"os"
)

// Run scanner and parser against symbol string sym.
func Calc(sym string) (ast.Expr, error) {
	return parser.ParseBytes([]byte(sym))
}

func main() {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Enter formula or (quit)\n")
		arg, err := r.ReadString('\n')
		if err != nil {
			fmt.Printf("Invalid string")
			continue
		}
		if arg == "quit\n" {
			break
		}
		_, err = Calc(arg)
		if err != nil {
			fmt.Printf("Error ParseBytes(%s)\n", arg)
		}
	}
}

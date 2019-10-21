package token

import "fmt"

type Position struct {
	Offset int
	Line   int
}

type Pos int

const NoPos Pos = 0

func (p Pos) IsValid() bool {
	return p != NoPos
}

func (pos Position) String() string {
	return fmt.Sprintf("%d %d", pos.Line, pos.Offset)
}

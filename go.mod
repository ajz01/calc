module github.com/ajz01/calc

replace github.com/ajz01/calc/ast => ./ast

replace github.com/ajz01/calc/parser => ./parser

replace github.com/ajz01/calc/scanner => ./scanner

replace github.com/ajz01/calc/token => ./token

go 1.12

require (
	github.com/ajz01/calc/ast v0.0.0
	github.com/ajz01/calc/parser v0.0.0-00010101000000-000000000000
)

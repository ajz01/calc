module github.com/ajz01/calc/parser

go 1.13

replace github.com/ajz01/calc/ast => ../ast

replace github.com/ajz01/calc/scanner => ../scanner

replace github.com/ajz01/calc/token => ../token

require (
	github.com/ajz01/calc/ast v0.0.0
	github.com/ajz01/calc/scanner v0.0.0-00010101000000-000000000000
	github.com/ajz01/calc/token v0.0.0-00010101000000-000000000000
)

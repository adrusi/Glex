package main

import (
	"fmt"
	"github.com/adrusi/glex"
	"os"
	"reflect"
	"strings"
)

var lexer *glex.Lexer

type binary int

const (
	plus binary = iota
	minus
	times
	div
)

type number float64

type lf struct{}

func init() {
	lexer = glex.NewLexer()

	lexer.Rule(`[\t ]+`, func() {}) // Skip whitespace.

	type lineCounter int
	var lc lineCounter
	lexer.Var(lc)

	lexer.Rule(`\n+`, func(lc *lineCounter) lf {
		(*lc)++
		fmt.Printf("\nLine #%d\n", *lc)
		return lf{}
	})

	lexer.Rule(`-?\d+(\.?\d+)?`, func(matches glex.Matches) number {
		text := matches[0]
		var x float64
		fmt.Sscanf(text, "%f", &x)
		return number(x)
	})

	lexer.Rule(`\+|-|\*|/`, func(matches glex.Matches) binary {
		text := matches[0]
		if text == "+" {
			return plus
		} else if text == "-" {
			return minus
		} else if text == "*" {
			return times
		} else {
			return div
		}
	})
}

func main() {
	in := strings.NewReader("1 + 1\n2 * 3")
	scanner := lexer.Lex(in)
	tok, err := scanner.Scan()
	for ; err == nil; tok, err = scanner.Scan() {
		fmt.Printf("%s\n", reflect.TypeOf(tok).Name())
	}
	fmt.Println()
	fmt.Fprintln(os.Stderr, err)
}

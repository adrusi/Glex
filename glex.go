package glex

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"regexp"
)

type Action interface{}

type rule struct {
	pattern *regexp.Regexp
	action  Action
}

type Lexer struct {
	rules []rule
	vars  []reflect.Type
}

func NewLexer() *Lexer {
	return &Lexer{
		rules: []rule{},
		vars:  []reflect.Type{},
	}
}

func (l *Lexer) Rule(p string, action Action) (err error) {
	pattern, err := regexp.Compile(fmt.Sprintf("\\A(?m:%s)", p))
	if err != nil {
		return
	}
	l.rules = append(l.rules, rule{pattern, action})
	return
}

// Takes a value for its type. Only one var may exist per type, although this
// rule is not enforced (currently), it simply doesn't work. The value passed
// is irrelevant, idiomatic use would look like this:
//     type depthCounter int
//     var dc depthCounter
//     lexer.Var(dc)
func (l *Lexer) Var(val interface{}) {
	l.vars = append(l.vars, reflect.TypeOf(val))
}

func (l Lexer) Lex(in io.Reader) (scanner *Scanner) {
	source := newFilePosReader(newRuneScanner(bufio.NewReader(in)))
	scanner = &Scanner{
		lexer:  l,
		source: source,
		line:   0,
		vars:   make([]reflect.Value, len(l.vars)),
		future: []interface{}{},
	}
	source.onNewLine = func(line int) {
		scanner.line = line
	}
	for i, t := range l.vars {
		scanner.vars[i] = reflect.New(t)
	}
	return
}

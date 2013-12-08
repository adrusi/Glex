package glex

import (
	"bufio"
	"strings"
	"testing"
)

const (
	lorem = `
		Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus
		quis rutrum nisl, vel congue dolor. Donec tincidunt massa id
		condimentum tristique. Vestibulum sed velit nec ligula convallis
		viverra eu vel mi. Sed aliquam ornare lorem, sit amet mattis ipsum
		porta eget. Ut urna justo, convallis nec vehicula sit amet, viverra
		quis lorem. Phasellus sit amet tempor erat. Vivamus tempus hendrerit
		leo, mattis aliquam nunc vehicula nec. Pellentesque feugiat fringilla
		gravida. Mauris condimentum elit ut consequat scelerisque. Nulla
		molestie tempor est vel consequat.
	`
	loremLineCount = 11
	http           = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n"
	httpLineCount  = 3
	utf            = "foo\u000Bbar\u000Cbaz\u0085qux\u2028\u2029"
	utfLineCount   = 6
)

func TestScannerTransactionTransparency(t *testing.T) {
	a := newScannerTransaction(bufio.NewReader(strings.NewReader(lorem)))
	b := bufio.NewReader(strings.NewReader(lorem))
	var err error
	for err == nil {
		r1, _, err := a.ReadRune()
		r2, _, err := b.ReadRune()
		if r1 != r2 {
			t.Error(
				"Rune from scannerTransaction did not match rune from bufio.")
		}
	}
	a.Commit()
}

func TestRevert(t *testing.T) {
	a := bufio.NewReader(strings.NewReader(lorem))
	const depth = 20
	s1, s2 = make([]rune, depth), make([]rune, depth)
	for _, s := range [][]rune{s1, s2} {
		b := newScannerTransaction(a)
		for i := 0; i < depth; i++ {
			s[i], _, _ = b.ReadRune()
		}
		b.revert()
	}
	for i := 0; i < depth; i++ {
		if s1[i] != s2[i] {
			t.Errorf("Runes after revert did not match after position %d", i)
		}
	}
}

func TestFilePosReaderTransparency(t *testing.T) {
	a := newFilePosReader(bufio.NewReader(strings.NewReader(lorem)))
	b := bufio.NewReader(strings.NewReader(lorem))
	var err error
	for err == nil {
		r1, _, err := a.ReadRune()
		r2, _, err := b.ReadRune()
		if r1 != r2 {
			t.Error(
				"Rune from filePosReader did not match rune from bufio.")
		}
	}
}

func TestLineCount(t *testing.T) {
	func test(name, s string, c int) {
		a := newFilePosReader(bufio.NewReader(strings.NewReader(s)))
		var err error
		for err == nil {
			_, _, err := a.ReadRune()
		}
		if a.line != c {
			t.Errorf("filePosReader failed to save line number of %s.", name)
		}
	}
	test("lorem", lorem, loremLineCount)
	test("http",  http,  httpLineCount)
	test("utf",   utf,   utfLineCount)
}

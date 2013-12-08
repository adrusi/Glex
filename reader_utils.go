package glex

import (
	"bufio"
	"errors"
	"io"
)

// scannerTransaction allows runes read from a RuneScanner to be reverted
// later. This is useful when tokenizing to allow the regexp package to
// attempt to match from a RuneReader and then revert the changes if the token
// doesn't match, so that the next possible token can be attempted.
type scannerTransaction struct {
	scanner   io.RuneScanner
	readCount int
}

func newScannerTransaction(scanner io.RuneScanner) scannerTransaction {
	return scannerTransaction{
		scanner:   scanner,
		readCount: 0,
	}
}

func (s scannerTransaction) ReadRune() (r rune, size int, err error) {
	r, size, err = s.scanner.ReadRune()
	if err == nil {
		readCount++
	}
	return
}

func (s scannerTransaction) revert() {
	for i := 0; i < s.readCount; i++ {
		s.scanner.UnreadRune()
	}
}

func (s scannerTransaction) commit() {
	// pass
}

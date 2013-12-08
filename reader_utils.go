package glex

import (
	"bufio"
	"errors"
	"io"
	"unicode"
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

// filePosReader allows tracking of line number while reading runes from an
// io.RuneReader.
type filePosReader struct {
	source     io.RuneReader
	line       int
	col        int
	justReadCR bool // allows the reader to recognize the multi-rune CRLF
	onNewLine  func(int)
}

func newFilePosReader(source io.RuneReader) *filePosReader {
	return &filePosReader{
		source:     source,
		line:       0,
		justReadCR: false,
		onNewLine:  func() {},
	}
}

func (f *filePosReader) ReadRune() (r rune, size int, err error) {
	r, size, err = f.source.ReadRune()
	if err != nil {
		return
	}
	// recognize line breaks
	switch r {
	case '\u000D':
		f.justReadCR = true
		f.line++
		f.onNewLine(f.line)
	case '\u000A':
		if !f.justReadCR {
			fallthrough
		}
	case '\u000B', '\u000C', '\u0085', '\u2028', '\u2029':
		f.line++
		f.onNewLine(f.line)
	}
	return
}

// Even if there is an error, UnreadRune cannot guarantee that the run was not
// unread. I may well have been read and the error come from a different
// source.
func (f *filePosReader) UnreadRune() (err error) {
	err = f.source.UnreadRune()
	if err != nil {
		return
	}
	r, _, err := f.source.ReadRune()
	if err != nil {
		return
	}
	err = f.source.UnreadRune()
	if err != nil {
		return
	}
	switch r {
	case '\u000D':
		f.justReadCR = true
		f.line--
		f.onNewLine(f.line)
	case '\u000A':
		if !f.justReadCR {
			fallthrough
		}
	case '\u000B', '\u000C', '\u0085', '\u2028', '\u2029':
		f.line--
		f.onNewLine(f.line)
	}
	return
}

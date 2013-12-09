package glex

import (
	"errors"
	"io"
)

type runeScanner struct {
	source io.RuneReader
	cache  []rune
	pos    int
}

func newRuneScanner(in io.RuneReader) *runeScanner {
	return &runeScanner{
		source: in,
		cache:  []rune{},
		pos:    0,
	}
}

func (s *runeScanner) ReadRune() (r rune, size int, err error) {
	if s.pos < len(s.cache) {
		r = s.cache[s.pos]
		s.pos++
		return
	}
	r, size, err = s.source.ReadRune()
	if err == nil {
		s.pos++
	}
	s.cache = append(s.cache, r)
	return
}

func (s *runeScanner) UnreadRune() error {
	if len(s.cache) == 0 {
		return errors.New("No runes left to unread.")
	}
	s.pos--
	return nil
}

// scannerTransaction allows runes read from a RuneScanner to be reverted
// later. This is useful when tokenizing to allow the regexp package to
// attempt to match from a RuneReader and then revert the changes if the token
// doesn't match, so that the next possible token can be attempted.
type scannerTransaction struct {
	scanner   io.RuneScanner
	readCount int
}

func newScannerTransaction(scanner io.RuneScanner) *scannerTransaction {
	return &scannerTransaction{
		scanner:   scanner,
		readCount: 0,
	}
}

func (s *scannerTransaction) ReadRune() (r rune, size int, err error) {
	r, size, err = s.scanner.ReadRune()
	if err == nil {
		s.readCount++
	}
	return
}

func (s *scannerTransaction) UnreadRune() (err error) {
	err = s.scanner.UnreadRune()
	if err != nil {
		return
	}
	s.readCount--
	return
}

func (s *scannerTransaction) revert() {
	for i := 0; i < s.readCount; i++ {
		s.scanner.UnreadRune()
	}
}

func (s *scannerTransaction) commit() {
	// pass
}

// filePosReader allows tracking of line number while reading runes from an
// io.RuneReader.
type filePosReader struct {
	source     io.RuneScanner
	line       int
	col        int
	justReadCR bool // allows the reader to recognize the multi-rune CRLF
	onNewLine  func(int)
}

func newFilePosReader(source io.RuneScanner) *filePosReader {
	return &filePosReader{
		source:     source,
		line:       1,
		justReadCR: false,
		onNewLine:  func(_ int) {},
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
			f.line++
			f.onNewLine(f.line)
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
			f.line--
			f.onNewLine(f.line)
		}
	case '\u000B', '\u000C', '\u0085', '\u2028', '\u2029':
		f.line--
		f.onNewLine(f.line)
	}
	return
}

// recallReader stores all the runes that it reads to be recalled later.
type recallReader struct {
	source       io.RuneScanner
	recollection []rune
}

func newRecallReader(in io.RuneScanner) *recallReader {
	return &recallReader{
		source:       in,
		recollection: []rune{},
	}
}

func (s *recallReader) ReadRune() (r rune, size int, err error) {
	r, size, err = s.source.ReadRune()
	if err != nil {
		return
	}
	s.recollection = append(s.recollection, r)
	return
}

func (s *recallReader) UnreadRune() (err error) {
	err = s.source.UnreadRune()
	if err != nil {
		return
	}
	s.recollection = s.recollection[:len(s.recollection)-1]
	return
}

func (s recallReader) recall() string {
	return string(s.recollection)
}

package glex

import (
	"errors"
	// "fmt"
	"io"
	"reflect"
)

type Scanner struct {
	source io.RuneScanner
	lexer  Lexer
	line   int
	vars   []reflect.Value // variables used by actions such as counters
	future []interface{}   // tokens to be returned on the next call to Scan
	// stored in REVERSE order.
}

type Matches []string

// Return a single token from the stream
func (s Scanner) Scan() (interface{}, error) {
	if len(s.future) > 0 {
		t := s.future[len(s.future)-1]
		s.future = s.future[:len(s.future)-1]
		return t, nil
	}
	for _, rule := range s.lexer.rules {
		recaller := newRecallReader(s.source)
		transaction := newScannerTransaction(recaller)
		rule.pattern.FindReaderSubmatchIndex(transaction)
		// weird thing where the above returns empty slice but below returns
		// matches.
		matches := rule.pattern.FindStringSubmatch(recaller.recall())
		// fmt.Printf("%s\t\t%q\t\t%s\n", rule.pattern, recaller.recall(), matches)
		if len(matches) > 0 {
			// Regexp might read more than the match from the RuneReader.
			// To remedy this we unread until only the match has been read.
			for len(matches[0]) < len(recaller.recall()) {
				recaller.UnreadRune() // TODO do something with this error
			}
			return s.handleAction(rule, matches)
		} else {
			transaction.revert()
		}
	}
	return nil, errors.New(
		"Failed to match any tokens. Maybe the stream ended?")
}

func (s Scanner) handleAction(r rule, matches Matches) (interface{}, error) {
	providers := make([]reflect.Value, len(s.vars)+1)
	for i, provider := range s.vars {
		providers[i] = provider
	}
	providers[len(providers)-1] = reflect.ValueOf(&matches)
	injector := newInjector(providers...)
	results, err := injector.call(r.action)
	if err != nil {
		return nil, err
	}

	// TODO handle >1 return values
	if len(results) == 0 {
		return s.Scan()
	}

	result := results[0]

	t := reflect.TypeOf(result)
	// Check if result is an anonymous struct, in which case we treat it as
	// multiple tokens.
	if t.Kind() == reflect.Slice && t.Name() == "" {
		v := reflect.ValueOf(result)
		if v.Len() == 0 {
			return s.Scan()
		}
		rest := v.Slice(1, v.Len())
		for i := 0; i < v.Len(); i++ {
			s.future = append(s.future, rest.Index(i).Interface())
		}
		return v.Index(0).Interface(), nil
	}

	return result, nil
}

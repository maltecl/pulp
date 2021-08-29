package pulp

import (
	"strings"
	"unicode/utf8"
)

type tokenTyp uint8

const (
	tokEof tokenTyp = iota
	tokGoSource
	tokOtherSource
)

type token struct {
	typ   tokenTyp
	value string
}

type lexer struct {
	input      string
	pos, start int
	width      int
	tokens     chan *token
	state      lexerFunc
}

const (
	eof = rune(-1)
)

type lexerFunc func(*lexer) lexerFunc

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		// l.width = 0
		l.emit(tokEof)
		return eof
	}

	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func lexUntilLBrace(l *lexer) lexerFunc {
	return lexUntil([2]rune{'{', '{'}, tokOtherSource, lexUntilRBrace)
}

func lexUntilRBrace(l *lexer) lexerFunc {
	return lexUntil([2]rune{'}', '}'}, tokGoSource, lexUntilLBrace)
}

func lexUntil(pattern [2]rune, tokenTyp tokenTyp, continueWith lexerFunc) lexerFunc {
	return func(l *lexer) lexerFunc {
		for {
			next := l.next()
			if next == eof {
				return nil
			}

			if next == pattern[0] && l.next() == pattern[1] {
				break
			}
		}

		l.backup()
		l.backup()

		l.emit(tokenTyp)

		l.next()
		l.next()
		l.ignore()

		return continueWith
	}
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) emit(t tokenTyp) {
	val := l.input[l.start:l.pos]
	tok := &token{t, strings.ReplaceAll(strings.ReplaceAll(val, "\n", ""), "\t", "")}
	l.tokens <- tok
	l.start = l.pos
}

func (l *lexer) run() {
	for state := l.state; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

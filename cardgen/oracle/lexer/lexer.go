// Package lexer tokenizes Oracle source text without interpreting its words.
package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

const eof = rune(-1)

// Lexer is a synchronous pull scanner for Scryfall Oracle text.
type Lexer struct {
	source string
	offset int
	line   int
	column int
}

// NewLexer returns a lexer positioned at the start of source.
func NewLexer(source string) *Lexer {
	return &Lexer{source: source, line: 1, column: 1}
}

// Next returns the next token. It always advances unless it returns EOF.
func (l *Lexer) Next() shared.Token {
	l.skipHorizontalSpace()
	start := l.position()
	if l.offset == len(l.source) {
		return shared.Token{Kind: shared.EOF, Span: shared.Span{Start: start, End: start}}
	}

	r, width := l.peek()
	if r == utf8.RuneError && width == 1 {
		l.advance()
		return l.token(shared.Invalid, start)
	}
	if r == 0 || r == '\uFEFF' {
		l.advance()
		return l.token(shared.Invalid, start)
	}

	switch {
	case isWordStart(r):
		return l.scanWord(start)
	case unicode.IsDigit(r):
		return l.scanInteger(start)
	}

	switch r {
	case '{':
		return l.scanSymbol(start)
	case '\r', '\n':
		l.scanNewline()
		return l.token(shared.Newline, start)
	case ',':
		l.advance()
		return l.token(shared.Comma, start)
	case '.':
		l.advance()
		return l.token(shared.Period, start)
	case ':':
		l.advance()
		return l.token(shared.Colon, start)
	case ';':
		l.advance()
		return l.token(shared.Semicolon, start)
	case '(':
		l.advance()
		return l.token(shared.LeftParen, start)
	case ')':
		l.advance()
		return l.token(shared.RightParen, start)
	case '"', '\u201C', '\u201D':
		l.advance()
		return l.token(shared.Quote, start)
	case '\u2022':
		l.advance()
		return l.token(shared.Bullet, start)
	case '\u2014':
		l.advance()
		return l.token(shared.EmDash, start)
	case '+':
		l.advance()
		return l.token(shared.Plus, start)
	case '-', '\u2212':
		l.advance()
		return l.token(shared.Minus, start)
	case '/':
		l.advance()
		return l.token(shared.Slash, start)
	case '*':
		l.advance()
		return l.token(shared.Asterisk, start)
	case '!':
		l.advance()
		return l.token(shared.Exclamation, start)
	case '?':
		l.advance()
		return l.token(shared.Question, start)
	case '\'', '\u2019':
		l.advance()
		return l.token(shared.Apostrophe, start)
	case '[':
		l.advance()
		return l.token(shared.LeftBracket, start)
	case ']':
		l.advance()
		return l.token(shared.RightBracket, start)
	case '&':
		l.advance()
		return l.token(shared.Ampersand, start)
	case '\u2013':
		l.advance()
		return l.token(shared.EnDash, start)
	default:
		l.advance()
		return l.token(shared.Glyph, start)
	}
}

func (l *Lexer) scanWord(start shared.Position) shared.Token {
	l.advance()
	for {
		r, _ := l.peek()
		if isWordContinue(r) {
			l.advance()
			continue
		}
		if isWordJoiner(r) && l.joinerContinuesWord() {
			l.advance()
			continue
		}
		break
	}
	return l.token(shared.Word, start)
}

func (l *Lexer) scanInteger(start shared.Position) shared.Token {
	for {
		r, _ := l.peek()
		if !unicode.IsDigit(r) {
			break
		}
		l.advance()
	}
	return l.token(shared.Integer, start)
}

func (l *Lexer) scanSymbol(start shared.Position) shared.Token {
	l.advance()
	for {
		r, width := l.peek()
		switch {
		case r == '}':
			l.advance()
			return l.token(shared.Symbol, start)
		case r == eof || r == '\r' || r == '\n':
			return l.token(shared.Invalid, start)
		case r == utf8.RuneError && width == 1:
			l.advance()
			return l.token(shared.Invalid, start)
		default:
			l.advance()
		}
	}
}

func (l *Lexer) scanNewline() {
	r, _ := l.peek()
	l.advance()
	if r == '\r' {
		next, _ := l.peek()
		if next == '\n' {
			l.advance()
		}
	}
}

func (l *Lexer) skipHorizontalSpace() {
	if l.offset == 0 && len(l.source) >= 3 &&
		l.source[0] == 0xEF && l.source[1] == 0xBB && l.source[2] == 0xBF {
		l.offset = 3
	}
	for {
		r, _ := l.peek()
		if r == eof || r == '\r' || r == '\n' || !unicode.IsSpace(r) {
			return
		}
		l.advance()
	}
}

func (l *Lexer) joinerContinuesWord() bool {
	_, width := l.peek()
	nextOffset := l.offset + width
	if nextOffset >= len(l.source) {
		return false
	}
	next, _ := utf8.DecodeRuneInString(l.source[nextOffset:])
	return isWordContinue(next)
}

func (l *Lexer) peek() (r rune, width int) {
	if l.offset >= len(l.source) {
		return eof, 0
	}
	return utf8.DecodeRuneInString(l.source[l.offset:])
}

func (l *Lexer) advance() rune {
	r, width := l.peek()
	if r == eof {
		return r
	}
	l.offset += width
	if r == '\n' || (r == '\r' && (l.offset >= len(l.source) || l.source[l.offset] != '\n')) {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return r
}

func (l *Lexer) position() shared.Position {
	return shared.Position{Offset: l.offset, Line: l.line, Column: l.column}
}

func (l *Lexer) token(kind shared.Kind, start shared.Position) shared.Token {
	end := l.position()
	return shared.Token{
		Kind: kind,
		Text: l.source[start.Offset:end.Offset],
		Span: shared.Span{Start: start, End: end},
	}
}

func isWordStart(r rune) bool {
	return unicode.IsLetter(r)
}

func isWordContinue(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsMark(r)
}

func isWordJoiner(r rune) bool {
	return r == '\'' || r == '\u2019' || r == '-'
}

// InvalidReason explains why an Invalid token could not be lexed. It returns
// an empty string for all other token kinds.
func InvalidReason(token shared.Token) string {
	if token.Kind != shared.Invalid {
		return ""
	}
	switch {
	case !utf8.ValidString(token.Text):
		return "invalid UTF-8 encoding"
	case strings.ContainsRune(token.Text, 0):
		return "NUL is not valid in Oracle text"
	case strings.ContainsRune(token.Text, '\uFEFF'):
		return "a UTF-8 BOM is only valid at the start of Oracle text"
	case strings.HasPrefix(token.Text, "{"):
		return "unclosed braced symbol"
	default:
		return "invalid Oracle text"
	}
}

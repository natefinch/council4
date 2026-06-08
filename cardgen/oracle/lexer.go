package oracle

import (
	"unicode"
	"unicode/utf8"
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
func (l *Lexer) Next() Token {
	l.skipHorizontalSpace()
	start := l.position()
	if l.offset == len(l.source) {
		return Token{Kind: EOF, Span: Span{Start: start, End: start}}
	}

	r, width := l.peek()
	if r == utf8.RuneError && width == 1 {
		l.advance()
		return l.token(Invalid, start)
	}
	if r == 0 || r == '\uFEFF' {
		l.advance()
		return l.token(Invalid, start)
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
		return l.token(Newline, start)
	case ',':
		l.advance()
		return l.token(Comma, start)
	case '.':
		l.advance()
		return l.token(Period, start)
	case ':':
		l.advance()
		return l.token(Colon, start)
	case ';':
		l.advance()
		return l.token(Semicolon, start)
	case '(':
		l.advance()
		return l.token(LeftParen, start)
	case ')':
		l.advance()
		return l.token(RightParen, start)
	case '"', '\u201C', '\u201D':
		l.advance()
		return l.token(Quote, start)
	case '\u2022':
		l.advance()
		return l.token(Bullet, start)
	case '\u2014':
		l.advance()
		return l.token(EmDash, start)
	case '+':
		l.advance()
		return l.token(Plus, start)
	case '-', '\u2212':
		l.advance()
		return l.token(Minus, start)
	case '/':
		l.advance()
		return l.token(Slash, start)
	case '*':
		l.advance()
		return l.token(Asterisk, start)
	case '!':
		l.advance()
		return l.token(Exclamation, start)
	case '?':
		l.advance()
		return l.token(Question, start)
	case '\'', '\u2019':
		l.advance()
		return l.token(Apostrophe, start)
	case '[':
		l.advance()
		return l.token(LeftBracket, start)
	case ']':
		l.advance()
		return l.token(RightBracket, start)
	case '&':
		l.advance()
		return l.token(Ampersand, start)
	case '\u2013':
		l.advance()
		return l.token(EnDash, start)
	default:
		l.advance()
		return l.token(Glyph, start)
	}
}

func (l *Lexer) scanWord(start Position) Token {
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
	return l.token(Word, start)
}

func (l *Lexer) scanInteger(start Position) Token {
	for {
		r, _ := l.peek()
		if !unicode.IsDigit(r) {
			break
		}
		l.advance()
	}
	return l.token(Integer, start)
}

func (l *Lexer) scanSymbol(start Position) Token {
	l.advance()
	for {
		r, width := l.peek()
		switch {
		case r == '}':
			l.advance()
			return l.token(Symbol, start)
		case r == eof || r == '\r' || r == '\n':
			return l.token(Invalid, start)
		case r == utf8.RuneError && width == 1:
			l.advance()
			return l.token(Invalid, start)
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

func (l *Lexer) position() Position {
	return Position{Offset: l.offset, Line: l.line, Column: l.column}
}

func (l *Lexer) token(kind Kind, start Position) Token {
	end := l.position()
	return Token{
		Kind: kind,
		Text: l.source[start.Offset:end.Offset],
		Span: Span{Start: start, End: end},
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

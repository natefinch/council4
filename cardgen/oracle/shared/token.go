// Package shared provides source and token primitives shared by Oracle pipeline
// stages.
package shared

import "fmt"

// Kind identifies a lexical token.
type Kind uint8

// Token kinds emitted by Lexer.
const (
	Invalid Kind = iota
	EOF
	Word
	Integer
	Symbol
	Newline
	Comma
	Period
	Colon
	Semicolon
	LeftParen
	RightParen
	Quote
	Bullet
	EmDash
	Plus
	Minus
	Slash
	Asterisk
	Exclamation
	Question
	Apostrophe
	LeftBracket
	RightBracket
	Ampersand
	EnDash
	Glyph
)

var kindNames = [...]string{
	Invalid:      "invalid",
	EOF:          "EOF",
	Word:         "word",
	Integer:      "integer",
	Symbol:       "symbol",
	Newline:      "newline",
	Comma:        "comma",
	Period:       "period",
	Colon:        "colon",
	Semicolon:    "semicolon",
	LeftParen:    "left parenthesis",
	RightParen:   "right parenthesis",
	Quote:        "quote",
	Bullet:       "bullet",
	EmDash:       "em dash",
	Plus:         "plus",
	Minus:        "minus",
	Slash:        "slash",
	Asterisk:     "asterisk",
	Exclamation:  "exclamation",
	Question:     "question",
	Apostrophe:   "apostrophe",
	LeftBracket:  "left bracket",
	RightBracket: "right bracket",
	Ampersand:    "ampersand",
	EnDash:       "en dash",
	Glyph:        "glyph",
}

func (k Kind) String() string {
	if int(k) >= len(kindNames) {
		return fmt.Sprintf("Kind(%d)", k)
	}
	return kindNames[k]
}

// Position is a location in Oracle text. Offset is a zero-based byte offset;
// Line and Column are one-based rune coordinates.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span is a half-open source range [Start, End).
type Span struct {
	Start Position
	End   Position
}

// Token is one lexical unit. Text is the exact source slice covered by Span.
type Token struct {
	Kind Kind
	Text string
	Span Span
}

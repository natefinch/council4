// Package shared provides source and token primitives shared by Oracle pipeline
// stages.
package shared

// Kind identifies a lexical token.
type Kind string

// Token kinds emitted by Lexer.
const (
	Invalid      Kind = ""
	EOF          Kind = "EOF"
	Word         Kind = "Word"
	Integer      Kind = "Integer"
	Symbol       Kind = "Symbol"
	Newline      Kind = "Newline"
	Comma        Kind = "Comma"
	Period       Kind = "Period"
	Colon        Kind = "Colon"
	Semicolon    Kind = "Semicolon"
	LeftParen    Kind = "LeftParen"
	RightParen   Kind = "RightParen"
	Quote        Kind = "Quote"
	Bullet       Kind = "Bullet"
	EmDash       Kind = "EmDash"
	Plus         Kind = "Plus"
	Minus        Kind = "Minus"
	Slash        Kind = "Slash"
	Asterisk     Kind = "Asterisk"
	Exclamation  Kind = "Exclamation"
	Question     Kind = "Question"
	Apostrophe   Kind = "Apostrophe"
	LeftBracket  Kind = "LeftBracket"
	RightBracket Kind = "RightBracket"
	Ampersand    Kind = "Ampersand"
	EnDash       Kind = "EnDash"
	Glyph        Kind = "Glyph"
)

var kindNames = map[Kind]string{
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

// String returns a human-readable name for the token kind.
func (k Kind) String() string {
	if name, ok := kindNames[k]; ok {
		return name
	}
	return string(k)
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

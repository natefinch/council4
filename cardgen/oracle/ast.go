package oracle

// AbilityKind is the syntactic category of an Oracle-text ability.
type AbilityKind uint8

// Ability kinds recognized by the syntax parser.
const (
	AbilityUnknown AbilityKind = iota
	AbilitySpell
	AbilityActivated
	AbilityLoyalty
	AbilityChapter
	AbilityTriggered
	AbilityReplacement
	AbilityStatic
	AbilityReminder
)

var abilityKindNames = [...]string{
	AbilityUnknown:     "unknown",
	AbilitySpell:       "spell",
	AbilityActivated:   "activated",
	AbilityLoyalty:     "loyalty",
	AbilityChapter:     "chapter",
	AbilityTriggered:   "triggered",
	AbilityReplacement: "replacement",
	AbilityStatic:      "static",
	AbilityReminder:    "reminder",
}

func (k AbilityKind) String() string {
	if int(k) >= len(abilityKindNames) {
		return "unknown"
	}
	return abilityKindNames[k]
}

// ParseContext supplies card-face facts that Oracle text alone cannot express.
type ParseContext struct {
	CardName         string
	InstantOrSorcery bool
	Planeswalker     bool
	Saga             bool
}

// Document is a lossless syntax tree for one card face's Oracle text.
type Document struct {
	Source    string
	Span      Span
	Abilities []Ability
}

// Ability is one Oracle-text paragraph, or one modal header and its options.
type Ability struct {
	Kind        AbilityKind
	Span        Span
	Text        string
	Tokens      []Token
	AbilityWord *Phrase
	Chapters    []int
	ChapterSpan Span
	Cost        *Phrase
	Sentences   []Sentence
	Reminders   []Delimited
	Quoted      []Delimited
	Modal       *Modal
}

// Phrase is a meaningful contiguous token range.
type Phrase struct {
	Span   Span
	Text   string
	Tokens []Token
}

// Sentence is a top-level sentence in an ability.
type Sentence struct {
	Span   Span
	Text   string
	Tokens []Token
}

// Delimited is parenthesized reminder text or a quoted granted ability.
type Delimited struct {
	Span   Span
	Text   string
	Tokens []Token
}

// Modal is a choose header followed by bullet or inline options.
type Modal struct {
	Header  Phrase
	Options []Mode
}

// Mode is one bullet option in a modal ability.
type Mode struct {
	Span      Span
	Text      string
	Tokens    []Token
	Sentences []Sentence
	Reminders []Delimited
	Quoted    []Delimited
}

// Severity is a parser diagnostic severity.
type Severity uint8

// Diagnostic severities.
const (
	SeverityError Severity = iota + 1
	SeverityWarning
)

// Diagnostic describes a localized lexical or syntax problem.
type Diagnostic struct {
	Severity Severity
	Summary  string
	Detail   string
	Span     Span
}

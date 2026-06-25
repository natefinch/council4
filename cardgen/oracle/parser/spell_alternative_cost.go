package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// SpellAlternativeCostKind identifies the rules change attached to an
// alternative spell cost.
type SpellAlternativeCostKind uint8

// Supported alternative spell-cost kinds.
const (
	SpellAlternativeCostUnknown SpellAlternativeCostKind = iota
	SpellAlternativeCostCommander
	SpellAlternativeCostOverload
	// SpellAlternativeCostPitch is the Force of Will family: exile a colored
	// card from hand (optionally paying extra life) rather than pay the
	// spell's mana cost.
	SpellAlternativeCostPitch
	// SpellAlternativeCostFlashback is the alternative-cost form of Flashback,
	// written "Flashback—<cost>" with an em dash before a non-mana (or compound)
	// cost. The cost itself is carried through the ability's CostSyntax; this
	// kind only marks the paragraph as granting Flashback for that cost.
	SpellAlternativeCostFlashback
	// SpellAlternativeCostEscape is the alternative-cost form of Escape, written
	// "Escape—<cost>, Exile N cards from your graveyard." with an em dash before
	// a compound cost. Like Flashback the cost is carried through the ability's
	// CostSyntax; this kind only marks the paragraph as granting Escape for that
	// cost (CR 702.139).
	SpellAlternativeCostEscape
	// SpellAlternativeCostDiscard is the Foil/Outbreak family: discard one or
	// more cards (each an optional subtype filter) from hand rather than pay the
	// spell's mana cost. The discarded cards ride in DiscardCards.
	SpellAlternativeCostDiscard
)

// SpellAlternativeCostCondition identifies a condition on an alternative spell cost.
type SpellAlternativeCostCondition uint8

// Supported alternative spell-cost conditions.
const (
	SpellAlternativeCostConditionUnknown SpellAlternativeCostCondition = iota
	SpellAlternativeCostConditionControlsCommander
	// SpellAlternativeCostConditionNotYourTurn gates a pitch alternative cost
	// behind "If it's not your turn," (Force of Negation).
	SpellAlternativeCostConditionNotYourTurn
)

// SpellAlternativeCost is typed syntax for a paragraph that offers an
// alternative to the spell's printed mana cost.
type SpellAlternativeCost struct {
	Span                  shared.Span
	Kind                  SpellAlternativeCostKind
	Condition             SpellAlternativeCostCondition
	WithoutPayingManaCost bool
	ManaCost              cost.Mana
	ReplaceTargetWithEach bool

	// DiscardCards lists the cards discarded from hand by a
	// SpellAlternativeCostDiscard cost, in printed order. Each entry carries an
	// optional subtype filter ("an Island card"); a bare "another card" or "a
	// card" entry imposes no filter.
	DiscardCards []AlternativeDiscardCard
}

// AlternativeDiscardCard is one card discarded by a discard alternative spell
// cost, optionally constrained to a card subtype ("an Island card").
type AlternativeDiscardCard struct {
	Subtype    types.Sub
	HasSubtype bool
}

func spellAlternativeCostClause(body []shared.Token) (*SpellAlternativeCost, *Cost, bool) {
	if alternative, ok := overloadAlternativeCostClause(body); ok {
		return alternative, nil, true
	}
	if alternative, pitchCost, ok := pitchAlternativeCostClause(body); ok {
		return alternative, pitchCost, true
	}
	if alternative, ok := discardAlternativeCostClause(body); ok {
		return alternative, nil, true
	}
	words := []string{
		"if", "you", "control", "a", "commander", "you", "may", "cast",
		"this", "spell", "without", "paying", "its", "mana", "cost",
	}
	if len(body) != len(words)+2 {
		return nil, nil, false
	}
	for tokenIndex, wordIndex := 0, 0; tokenIndex < len(body); tokenIndex++ {
		switch tokenIndex {
		case 5:
			if body[tokenIndex].Kind != shared.Comma {
				return nil, nil, false
			}
		case 3:
			// The commander determiner is printed as either "a commander" or
			// "your commander"; both name the controller's commander.
			if body[tokenIndex].Kind != shared.Word ||
				(!equalWord(body[tokenIndex], "a") && !equalWord(body[tokenIndex], "your")) {
				return nil, nil, false
			}
			wordIndex++
		case len(body) - 1:
			if body[tokenIndex].Kind != shared.Period {
				return nil, nil, false
			}
		default:
			if body[tokenIndex].Kind != shared.Word || !equalWord(body[tokenIndex], words[wordIndex]) {
				return nil, nil, false
			}
			wordIndex++
		}
	}
	return &SpellAlternativeCost{
		Span:                  shared.SpanOf(body),
		Kind:                  SpellAlternativeCostCommander,
		Condition:             SpellAlternativeCostConditionControlsCommander,
		WithoutPayingManaCost: true,
	}, nil, true
}

// flashbackAlternativeCostClause recognizes the em-dash Flashback form
// "Flashback—<cost>", where the pre-dash label is exactly the Flashback keyword
// and the post-dash body is a non-mana or compound cost. The cost tokens become
// the paragraph's cost phrase so the shared cost machinery types them; the
// returned span covers the whole paragraph so its label and dash are accounted
// for in coverage. It returns ok=false when the label is not Flashback or the
// body is empty.
func flashbackAlternativeCostClause(source string, tokens []shared.Token, dash int) (*SpellAlternativeCost, Phrase, bool) {
	if !slices.Equal(normalizedWords(tokens[:dash]), []string{"flashback"}) {
		return nil, Phrase{}, false
	}
	clause := tokens[dash+1:]
	if period := shared.TopLevelIndex(clause, shared.Period); period >= 0 {
		clause = clause[:period]
	}
	if len(clause) == 0 {
		return nil, Phrase{}, false
	}
	return &SpellAlternativeCost{
		Span: shared.SpanOf(tokens),
		Kind: SpellAlternativeCostFlashback,
	}, phraseFromTokens(source, clause), true
}

// escapeAlternativeCostClause recognizes the em-dash Escape form
// "Escape—<cost>, Exile N cards from your graveyard.", where the pre-dash label
// is exactly the Escape keyword and the post-dash body is the compound escape
// cost (its mana cost plus the graveyard-exile additional cost). The cost tokens
// become the paragraph's cost phrase so the shared cost machinery types them;
// the returned span covers the whole paragraph so its label and dash are
// accounted for in coverage. It returns ok=false when the label is not Escape or
// the body is empty.
func escapeAlternativeCostClause(source string, tokens []shared.Token, dash int) (*SpellAlternativeCost, Phrase, bool) {
	if !slices.Equal(normalizedWords(tokens[:dash]), []string{"escape"}) {
		return nil, Phrase{}, false
	}
	clause := tokens[dash+1:]
	if period := shared.TopLevelIndex(clause, shared.Period); period >= 0 {
		clause = clause[:period]
	}
	if len(clause) == 0 {
		return nil, Phrase{}, false
	}
	return &SpellAlternativeCost{
		Span: shared.SpanOf(tokens),
		Kind: SpellAlternativeCostEscape,
	}, phraseFromTokens(source, clause), true
}

// discardAlternativeCostClause recognizes the Foil/Outbreak discard pitch
// family: "you may discard <card>[ and <card>] rather than pay this spell's mana
// cost", where each card is "a/an [<subtype>] card" or "another card". The cards
// are discarded from hand; an optional subtype constrains a card to that subtype
// ("an Island card").
func discardAlternativeCostClause(body []shared.Token) (*SpellAlternativeCost, bool) {
	if !equalWordSequence(body, 0, "you", "may", "discard") {
		return nil, false
	}
	cursor := 3
	var cards []AlternativeDiscardCard
	for {
		card, next, ok := matchDiscardCardSpec(body, cursor)
		if !ok {
			return nil, false
		}
		cards = append(cards, card)
		cursor = next
		if cursor < len(body) && equalWord(body[cursor], "and") {
			cursor++
			continue
		}
		break
	}
	if !equalWordSequence(body, cursor,
		"rather", "than", "pay", "this", "spell's", "mana", "cost") {
		return nil, false
	}
	cursor += 7
	if cursor != len(body)-1 || body[cursor].Kind != shared.Period {
		return nil, false
	}
	return &SpellAlternativeCost{
		Span:         shared.SpanOf(body),
		Kind:         SpellAlternativeCostDiscard,
		DiscardCards: cards,
	}, true
}

// matchDiscardCardSpec parses one "a/an [<subtype>] card" or "another card"
// discard target starting at start, returning the parsed card and the index
// immediately after the matched "card" noun.
func matchDiscardCardSpec(body []shared.Token, start int) (AlternativeDiscardCard, int, bool) {
	cursor := start
	if cursor >= len(body) || body[cursor].Kind != shared.Word ||
		(!equalWord(body[cursor], "a") && !equalWord(body[cursor], "an") &&
			!equalWord(body[cursor], "another")) {
		return AlternativeDiscardCard{}, 0, false
	}
	cursor++
	var card AlternativeDiscardCard
	if cursor < len(body) && body[cursor].Kind == shared.Word && !equalWord(body[cursor], "card") {
		sub, ok := recognizeSubtypePhrase(body[cursor].Text)
		if !ok {
			return AlternativeDiscardCard{}, 0, false
		}
		card.Subtype = sub
		card.HasSubtype = true
		cursor++
	}
	if cursor >= len(body) || !equalWord(body[cursor], "card") {
		return AlternativeDiscardCard{}, 0, false
	}
	cursor++
	return card, cursor, true
}

func overloadAlternativeCostClause(body []shared.Token) (*SpellAlternativeCost, bool) {
	if len(body) < 2 || body[0].Kind != shared.Word || !equalWord(body[0], "overload") {
		return nil, false
	}
	manaCost, end, ok := parseKeywordManaCost(body, 1)
	if !ok || len(manaCost) == 0 {
		return nil, false
	}
	if end != len(body) && !canonicalOverloadReminder(body[end:]) {
		return nil, false
	}
	return &SpellAlternativeCost{
		Span:                  shared.SpanOf(body),
		Kind:                  SpellAlternativeCostOverload,
		ManaCost:              manaCost,
		ReplaceTargetWithEach: true,
	}, true
}

// pitchAlternativeCostClause recognizes the Force of Will pitch family:
// "[If it's not your turn, ] you may [pay N life and ] exile a/<count> <color>
// card[s] from your hand rather than pay this spell's mana cost". The non-mana
// payment is emitted as ordered typed cost components (an optional pay-life
// component followed by an exile-from-hand component) so it lowers through the
// same cost machinery as any other ability cost.
func pitchAlternativeCostClause(body []shared.Token) (*SpellAlternativeCost, *Cost, bool) {
	cursor := 0
	condition := SpellAlternativeCostConditionUnknown
	if equalWordSequence(body, cursor, "if", "it's", "not", "your", "turn") {
		cursor += 5
		if cursor >= len(body) || body[cursor].Kind != shared.Comma {
			return nil, nil, false
		}
		cursor++
		condition = SpellAlternativeCostConditionNotYourTurn
	}
	if !equalWordSequence(body, cursor, "you", "may") {
		return nil, nil, false
	}
	cursor += 2
	var components []CostComponent
	if cursor < len(body) && equalWord(body[cursor], "pay") {
		if cursor+3 >= len(body) ||
			body[cursor+1].Kind != shared.Integer ||
			!equalWord(body[cursor+2], "life") ||
			!equalWord(body[cursor+3], "and") {
			return nil, nil, false
		}
		value, ok := conditionNumberValue(body[cursor+1])
		if !ok || value <= 0 {
			return nil, nil, false
		}
		components = append(components, CostComponent{
			Kind:        CostComponentPayLife,
			Span:        shared.SpanOf(body[cursor : cursor+3]),
			AmountValue: value,
			AmountKnown: true,
		})
		cursor += 4
	}
	exileStart := cursor
	exile, ok := matchPitchExileClause(body, cursor)
	if !ok {
		return nil, nil, false
	}
	cursor = exile.next
	components = append(components, CostComponent{
		Kind:             CostComponentExile,
		Span:             shared.SpanOf(body[exileStart:cursor]),
		AmountValue:      exile.count,
		AmountKnown:      true,
		ObjectIsCard:     true,
		ObjectNoun:       ObjectNounCard,
		ObjectColor:      exile.color,
		ObjectColorKnown: true,
		SourceZone:       zone.Hand,
	})
	if !equalWordSequence(body, cursor,
		"from", "your", "hand", "rather", "than", "pay", "this", "spell's", "mana", "cost") {
		return nil, nil, false
	}
	cursor += 10
	if cursor != len(body)-1 || body[cursor].Kind != shared.Period {
		return nil, nil, false
	}
	pitchCost := &Cost{
		Span:       shared.SpanOf(body[exileStart:exile.next]),
		Components: components,
	}
	return &SpellAlternativeCost{
		Span:      shared.SpanOf(body),
		Kind:      SpellAlternativeCostPitch,
		Condition: condition,
	}, pitchCost, true
}

// pitchExileClause is the parsed "exile a/<count> <color> card[s]" segment of a
// pitch alternative cost: the card count, the required color, and the token
// index immediately after the matched "card"/"cards" noun.
type pitchExileClause struct {
	count int
	color Color
	next  int
}

// matchPitchExileClause parses "exile a <color> card" or "exile <count> <color>
// cards" starting at start.
func matchPitchExileClause(body []shared.Token, start int) (pitchExileClause, bool) {
	cursor := start
	if cursor >= len(body) || !equalWord(body[cursor], "exile") {
		return pitchExileClause{}, false
	}
	cursor++
	if cursor >= len(body) || body[cursor].Kind != shared.Word {
		return pitchExileClause{}, false
	}
	count := 0
	if equalWord(body[cursor], "a") {
		count = 1
	} else if value, ok := CardinalWordValue(body[cursor].Text); ok && value >= 1 {
		count = value
	} else {
		return pitchExileClause{}, false
	}
	cursor++
	if cursor >= len(body) || body[cursor].Kind != shared.Word {
		return pitchExileClause{}, false
	}
	color, ok := recognizeColorWord(body[cursor].Text)
	if !ok {
		return pitchExileClause{}, false
	}
	cursor++
	cardNoun := "card"
	if count > 1 {
		cardNoun = "cards"
	}
	if cursor >= len(body) || !equalWord(body[cursor], cardNoun) {
		return pitchExileClause{}, false
	}
	cursor++
	return pitchExileClause{count: count, color: color, next: cursor}, true
}

func equalWordSequence(body []shared.Token, start int, words ...string) bool {
	if start < 0 || start+len(words) > len(body) {
		return false
	}
	for offset, word := range words {
		if !equalWord(body[start+offset], word) {
			return false
		}
	}
	return true
}

func canonicalOverloadReminder(tokens []shared.Token) bool {
	if len(tokens) < 2 || tokens[0].Kind != shared.LeftParen || tokens[len(tokens)-1].Kind != shared.RightParen {
		return false
	}
	var normalized strings.Builder
	for _, token := range tokens {
		_, _ = normalized.WriteString(strings.ToLower(token.Text))
	}
	switch normalized.String() {
	case `(youmaycastthisspellforitsoverloadcost.ifyoudo,change"target"initstextto"each.")`,
		`(youmaycastthisspellforitsoverloadcost.ifyoudo,changeitstextbyreplacingallinstancesof"target"with"each.")`:
		return true
	default:
		return false
	}
}

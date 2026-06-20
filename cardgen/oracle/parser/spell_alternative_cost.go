package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
)

// SpellAlternativeCostKind identifies the rules change attached to an
// alternative spell cost.
type SpellAlternativeCostKind uint8

// Supported alternative spell-cost kinds.
const (
	SpellAlternativeCostUnknown SpellAlternativeCostKind = iota
	SpellAlternativeCostCommander
	SpellAlternativeCostOverload
)

// SpellAlternativeCostCondition identifies a condition on an alternative spell cost.
type SpellAlternativeCostCondition uint8

// Supported alternative spell-cost conditions.
const (
	SpellAlternativeCostConditionUnknown SpellAlternativeCostCondition = iota
	SpellAlternativeCostConditionControlsCommander
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
}

func spellAlternativeCostClause(body []shared.Token) (*SpellAlternativeCost, bool) {
	if alternative, ok := overloadAlternativeCostClause(body); ok {
		return alternative, true
	}
	words := []string{
		"if", "you", "control", "a", "commander", "you", "may", "cast",
		"this", "spell", "without", "paying", "its", "mana", "cost",
	}
	if len(body) != len(words)+2 {
		return nil, false
	}
	for tokenIndex, wordIndex := 0, 0; tokenIndex < len(body); tokenIndex++ {
		switch tokenIndex {
		case 5:
			if body[tokenIndex].Kind != shared.Comma {
				return nil, false
			}
		case len(body) - 1:
			if body[tokenIndex].Kind != shared.Period {
				return nil, false
			}
		default:
			if body[tokenIndex].Kind != shared.Word || !equalWord(body[tokenIndex], words[wordIndex]) {
				return nil, false
			}
			wordIndex++
		}
	}
	return &SpellAlternativeCost{
		Span:                  shared.SpanOf(body),
		Kind:                  SpellAlternativeCostCommander,
		Condition:             SpellAlternativeCostConditionControlsCommander,
		WithoutPayingManaCost: true,
	}, true
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
